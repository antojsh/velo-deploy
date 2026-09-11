package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"velo-deploy/internal/config"
	"velo-deploy/internal/deploy"
	"velo-deploy/internal/systemd"
	"velo-deploy/internal/tui"
	"velo-deploy/internal/webhook"
)

var version = "dev"

const (
	exitOK       = 0
	exitErr      = 1
	exitUsage    = 2
	exitNotFound = 3
	exitPerm     = 4
	exitPort     = 5
	exitGit      = 6
	exitBuild    = 7
	exitHealth   = 8
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	globals, rest, err := parseGlobals(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitUsage
	}
	if globals.help && len(rest) == 0 {
		printRootHelp(os.Stdout)
		return exitOK
	}
	if globals.version {
		fmt.Println(version)
		return exitOK
	}
	if globals.configPath != "" {
		os.Setenv("VELO_CONFIG", globals.configPath)
	}
	if globals.noColor {
		os.Setenv("NO_COLOR", "1")
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return exitErr
	}

	if len(rest) == 0 {
		if err := tui.NewProgram(cfg).Start(); err != nil {
			fmt.Fprintf(os.Stderr, "tui: %v\n", err)
			return exitErr
		}
		return exitOK
	}

	cmd := rest[0]
	cmdArgs := rest[1:]
	if globals.help {
		printCommandHelp(os.Stdout, cmd)
		return exitOK
	}

	switch cmd {
	case "help", "-h", "--help":
		printRootHelp(os.Stdout)
		return exitOK
	case "version":
		fmt.Println(version)
		return exitOK
	case "deploy":
		return cmdDeploy(cfg, cmdArgs)
	case "add":
		return cmdAdd(cfg, cmdArgs)
	case "list":
		return cmdList(cfg, cmdArgs)
	case "restart":
		return cmdRestart(cfg, cmdArgs)
	case "stop":
		return cmdStop(cfg, cmdArgs)
	case "logs":
		return cmdLogs(cfg, cmdArgs)
	case "remove":
		return cmdRemove(cfg, cmdArgs)
	case "daemon":
		return cmdDaemon(cfg, cmdArgs)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printRootHelp(os.Stderr)
		return exitUsage
	}
}

type globalFlags struct {
	help       bool
	version    bool
	noColor    bool
	configPath string
}

func parseGlobals(args []string) (globalFlags, []string, error) {
	var g globalFlags
	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--":
			return g, args[i+1:], nil
		case arg == "-h" || arg == "--help":
			g.help = true
			i++
		case arg == "-V" || arg == "--version":
			g.version = true
			i++
		case arg == "--no-color":
			g.noColor = true
			i++
		case arg == "--config":
			if i+1 >= len(args) {
				return g, nil, errors.New("--config requires a path")
			}
			g.configPath = args[i+1]
			i += 2
		case strings.HasPrefix(arg, "--config="):
			g.configPath = strings.TrimPrefix(arg, "--config=")
			i++
		case strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "-"):
			return g, nil, fmt.Errorf("unknown flag: %s", arg)
		default:
			if strings.HasPrefix(arg, "-") {
				return g, args[i:], nil
			}
			return g, args[i:], nil
		}
	}
	return g, nil, nil
}

func cmdDeploy(cfg *config.Config, args []string) int {
	if err := requireRoot(); err != nil {
		return mapErr(err)
	}
	opts := deploy.DeployOptions{}
	var repo string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--name":
			opts.Name, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--name="):
			opts.Name = strings.TrimPrefix(arg, "--name=")
		case arg == "--domain":
			opts.Domain, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--domain="):
			opts.Domain = strings.TrimPrefix(arg, "--domain=")
		case arg == "--path":
			opts.Path, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--path="):
			opts.Path = strings.TrimPrefix(arg, "--path=")
		case arg == "--alias":
			opts.Alias, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--alias="):
			opts.Alias = strings.TrimPrefix(arg, "--alias=")
		case arg == "--port":
			opts.Port, i = takeInt(args, i)
		case strings.HasPrefix(arg, "--port="):
			opts.Port = atoi(strings.TrimPrefix(arg, "--port="))
		case arg == "--branch":
			opts.Branch, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--branch="):
			opts.Branch = strings.TrimPrefix(arg, "--branch=")
		case arg == "--output-dir":
			opts.OutputDir, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--output-dir="):
			opts.OutputDir = strings.TrimPrefix(arg, "--output-dir=")
		case arg == "--build-command" || arg == "--build-cmd":
			opts.BuildCommand, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--build-command="):
			opts.BuildCommand = strings.TrimPrefix(arg, "--build-command=")
		case strings.HasPrefix(arg, "--build-cmd="):
			opts.BuildCommand = strings.TrimPrefix(arg, "--build-cmd=")
		case arg == "--start-command" || arg == "--start-cmd":
			opts.StartCommand, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--start-command="):
			opts.StartCommand = strings.TrimPrefix(arg, "--start-command=")
		case strings.HasPrefix(arg, "--start-cmd="):
			opts.StartCommand = strings.TrimPrefix(arg, "--start-cmd=")
		case arg == "--force":
			opts.Force = true
		case arg == "--duckdns":
			opts.DuckDNS = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return exitUsage
		default:
			if repo == "" {
				repo = arg
			} else {
				fmt.Fprintln(os.Stderr, "unexpected argument:", arg)
				return exitUsage
			}
		}
	}
	if repo == "" {
		fmt.Fprintln(os.Stderr, "usage: velo-deploy deploy <repo-url> [flags]")
		return exitUsage
	}
	opts.RepoURL = repo
	if err := deploy.DeployWithOptions(cfg, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return mapErr(err)
	}
	fmt.Println("deployed")
	return exitOK
}

func cmdAdd(cfg *config.Config, args []string) int {
	if err := requireRoot(); err != nil {
		return mapErr(err)
	}
	var name, path, domain, urlPath, alias, typeHint, outputDir string
	var port int
	duckDNS := false
	positionals := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--type":
			typeHint, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--type="):
			typeHint = strings.TrimPrefix(arg, "--type=")
		case arg == "--domain":
			domain, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--domain="):
			domain = strings.TrimPrefix(arg, "--domain=")
		case arg == "--path":
			urlPath, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--path="):
			urlPath = strings.TrimPrefix(arg, "--path=")
		case arg == "--alias":
			alias, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--alias="):
			alias = strings.TrimPrefix(arg, "--alias=")
		case arg == "--port":
			port, i = takeInt(args, i)
		case strings.HasPrefix(arg, "--port="):
			port = atoi(strings.TrimPrefix(arg, "--port="))
		case arg == "--output-dir":
			outputDir, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--output-dir="):
			outputDir = strings.TrimPrefix(arg, "--output-dir=")
		case arg == "--duckdns":
			duckDNS = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return exitUsage
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) < 2 {
		fmt.Fprintln(os.Stderr, "usage: velo-deploy add <name> <path> [flags]")
		return exitUsage
	}
	name, path = positionals[0], positionals[1]
	if err := deploy.DeployWithOptions(cfg, deploy.DeployOptions{
		Name:      name,
		LocalDir:  path,
		Domain:    domain,
		Path:      urlPath,
		Alias:     alias,
		Port:      port,
		OutputDir: outputDir,
		Type:      typeHint,
		DuckDNS:   duckDNS,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return mapErr(err)
	}
	fmt.Println("added", name)
	return exitOK
}

func cmdList(cfg *config.Config, args []string) int {
	jsonOut := false
	quiet := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOut = true
		case "--quiet":
			quiet = true
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return exitUsage
		}
	}
	type row struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Domain string `json:"domain"`
		Alias  string `json:"alias"`
		Path   string `json:"path"`
		Port   int    `json:"port,omitempty"`
		Status string `json:"status"`
	}
	var rows []row
	for name, app := range cfg.Apps {
		status := "running"
		if app.Type == config.AppTypeNode {
			if systemd.IsActive(name) {
				status = "running"
			} else {
				status = "stopped"
			}
		}
		rows = append(rows, row{
			Name:   app.Name,
			Type:   app.Type,
			Domain: app.Domain,
			Alias:  app.Alias,
			Path:   app.Path,
			Port:   app.Port,
			Status: status,
		})
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rows)
		return exitOK
	}
	if quiet {
		for _, r := range rows {
			fmt.Println(r.Name)
		}
		return exitOK
	}
	fmt.Printf("%-14s %-8s %-20s %-18s %s\n", "NAME", "TYPE", "DOMAIN", "ALIAS", "STATUS")
	for _, r := range rows {
		domain := r.Domain
		if domain == "" {
			domain = "-"
		}
		fmt.Printf("%-14s %-8s %-20s %-18s %s\n", r.Name, r.Type, domain, r.Alias, r.Status)
	}
	return exitOK
}

func cmdRestart(cfg *config.Config, args []string) int {
	if err := requireRoot(); err != nil {
		return mapErr(err)
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: velo-deploy restart <name>")
		return exitUsage
	}
	app := cfg.GetApp(args[0])
	if app == nil {
		fmt.Fprintln(os.Stderr, deploy.ErrNotFound)
		return exitNotFound
	}
	if app.Type == config.AppTypeStatic {
		fmt.Println("static apps are served by Caddy; nothing to restart")
		return exitOK
	}
	if err := systemd.RestartApp(app.Name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitErr
	}
	return exitOK
}

func cmdStop(cfg *config.Config, args []string) int {
	if err := requireRoot(); err != nil {
		return mapErr(err)
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: velo-deploy stop <name>")
		return exitUsage
	}
	app := cfg.GetApp(args[0])
	if app == nil {
		fmt.Fprintln(os.Stderr, deploy.ErrNotFound)
		return exitNotFound
	}
	if err := systemd.StopApp(app.Name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitErr
	}
	return exitOK
}

func cmdLogs(cfg *config.Config, args []string) int {
	follow := false
	lines := 200
	var name string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-f" || arg == "--follow":
			follow = true
		case arg == "-n" || arg == "--lines":
			lines, i = takeInt(args, i)
		case strings.HasPrefix(arg, "--lines="):
			lines = atoi(strings.TrimPrefix(arg, "--lines="))
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return exitUsage
		default:
			name = arg
		}
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: velo-deploy logs <name> [flags]")
		return exitUsage
	}
	if cfg.GetApp(name) == nil {
		fmt.Fprintln(os.Stderr, deploy.ErrNotFound)
		return exitNotFound
	}
	if err := systemd.ShowLogs(name, follow, lines); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitErr
	}
	return exitOK
}

func cmdRemove(cfg *config.Config, args []string) int {
	if err := requireRoot(); err != nil {
		return mapErr(err)
	}
	purge := false
	var name string
	for _, arg := range args {
		switch {
		case arg == "--purge":
			purge = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return exitUsage
		default:
			name = arg
		}
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: velo-deploy remove <name> [flags]")
		return exitUsage
	}
	if err := deploy.RemoveApp(cfg, name, purge); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return mapErr(err)
	}
	fmt.Println("removed", name)
	return exitOK
}

func cmdDaemon(cfg *config.Config, args []string) int {
	if err := requireRoot(); err != nil {
		return mapErr(err)
	}
	port := cfg.DaemonPort
	if port == "" {
		port = "9999"
	}
	host := "0.0.0.0"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--port":
			port, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--port="):
			port = strings.TrimPrefix(arg, "--port=")
		case arg == "--host":
			host, i = takeValue(args, i)
		case strings.HasPrefix(arg, "--host="):
			host = strings.TrimPrefix(arg, "--host=")
		case arg == "--secret":
			secret, next := takeValue(args, i)
			os.Setenv("VELO_WEBHOOK_SECRET", secret)
			i = next
		case strings.HasPrefix(arg, "--secret="):
			os.Setenv("VELO_WEBHOOK_SECRET", strings.TrimPrefix(arg, "--secret="))
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return exitUsage
		}
	}
	if cfg.ResolveWebhookSecret() == "" {
		fmt.Fprintln(os.Stderr, "webhook secret is required: set VELO_WEBHOOK_SECRET or "+config.WebhookSecretFile())
		return exitErr
	}
	addr := webhook.BindAddr(host, port)
	fmt.Println("listening on", addr)
	srv := &webhook.Server{Cfg: cfg, Addr: addr}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitErr
	}
	return exitOK
}

func requireRoot() error {
	if deploy.NeedRoot() {
		return errPerm
	}
	return nil
}

var errPerm = errors.New("permission denied: re-run as root")

func mapErr(err error) int {
	if err == nil {
		return exitOK
	}
	if errors.Is(err, errPerm) {
		return exitPerm
	}
	if errors.Is(err, deploy.ErrNotFound) {
		return exitNotFound
	}
	if errors.Is(err, deploy.ErrPort) {
		return exitPort
	}
	if errors.Is(err, deploy.ErrGit) {
		return exitGit
	}
	if errors.Is(err, deploy.ErrBuild) {
		return exitBuild
	}
	if errors.Is(err, deploy.ErrHealth) {
		return exitHealth
	}
	if errors.Is(err, deploy.ErrName) {
		return exitUsage
	}
	return exitErr
}

func takeValue(args []string, i int) (string, int) {
	if i+1 >= len(args) {
		return "", i
	}
	return args[i+1], i + 1
}

func takeInt(args []string, i int) (int, int) {
	v, i := takeValue(args, i)
	return atoi(v), i
}

func atoi(v string) int {
	n, _ := strconv.Atoi(v)
	return n
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, `velo-deploy — Bare Metal PaaS

Usage:
  velo-deploy [global flags] <command> [flags]

Commands:
  deploy    Clone, build and register an app from Git
  add       Register an app that already exists on disk
  list      List registered apps
  restart   Restart a Node.js app
  stop      Stop a Node.js app
  logs      Show app logs
  remove    Unregister an app
  daemon    Run the GitHub webhook listener

Global flags:
  -h, --help          Show help
  -V, --version       Print version
      --no-color      Disable ANSI colors
      --config PATH   Override config file`)
}

func printCommandHelp(w io.Writer, cmd string) {
	printRootHelp(w)
	fmt.Fprintln(w, "\nCommand:", cmd)
}
