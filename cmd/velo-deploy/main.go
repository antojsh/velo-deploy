package main

import (
	"fmt"
	"os"

	"velo-deploy/internal/config"
	"velo-deploy/internal/deploy"
	"velo-deploy/internal/systemd"
	"velo-deploy/internal/tui"
)

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		// No subcommand -> launch TUI
		p := tui.NewProgram(cfg)
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("velo-deploy %s\n", version)
	case "deploy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: velo-deploy deploy <repo-url> [--domain example.com] [--build-command \"npm run build\"] [--start-command \"npm run start\"]")
			os.Exit(1)
		}
		repoURL := os.Args[2]
		var domain, alias string
		opts := deploy.Options{}
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--domain" && i+1 < len(os.Args) {
				domain = os.Args[i+1]
				i++
			}
			if os.Args[i] == "--alias" && i+1 < len(os.Args) {
				alias = os.Args[i+1]
				i++
			}
			if (os.Args[i] == "--build-command" || os.Args[i] == "--build-cmd") && i+1 < len(os.Args) {
				opts.BuildCommand = os.Args[i+1]
				i++
			}
			if (os.Args[i] == "--start-command" || os.Args[i] == "--start-cmd") && i+1 < len(os.Args) {
				opts.StartCommand = os.Args[i+1]
				i++
			}
		}
		if err := deploy.DeployWithOptions(cfg, repoURL, domain, alias, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Deploy failed: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: velo-deploy add <name> <directory> [--domain example.com] [--type auto|static|node]")
			os.Exit(1)
		}
		appName := os.Args[2]
		appDir := os.Args[3]
		var domain, appType string
		for i := 4; i < len(os.Args); i++ {
			if os.Args[i] == "--domain" && i+1 < len(os.Args) {
				domain = os.Args[i+1]
				i++
			}
			if os.Args[i] == "--type" && i+1 < len(os.Args) {
				appType = os.Args[i+1]
				i++
			}
		}
		if err := deploy.Register(cfg, appName, appDir, domain, "", appType); err != nil {
			fmt.Fprintf(os.Stderr, "Add failed: %v\n", err)
			os.Exit(1)
		}
	case "list":
		systemd.ListApps(cfg)
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: velo-deploy remove <app-name>")
			os.Exit(1)
		}
		if err := deploy.Remove(cfg, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Remove failed: %v\n", err)
			os.Exit(1)
		}
	case "restart":
		if len(os.Args) < 3 {
			fmt.Println("Usage: velo-deploy restart <app-name>")
			os.Exit(1)
		}
		if err := systemd.RestartApp(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Restart failed: %v\n", err)
			os.Exit(1)
		}
	case "logs":
		if len(os.Args) < 3 {
			fmt.Println("Usage: velo-deploy logs <app-name>")
			os.Exit(1)
		}
		follow := false
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "-f" {
				follow = true
			}
		}
		if err := systemd.ShowLogs(os.Args[2], follow); err != nil {
			fmt.Fprintf(os.Stderr, "Logs failed: %v\n", err)
			os.Exit(1)
		}
	case "daemon":
		// Git-watcher daemon
		port := "9999"
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--port" && i+1 < len(os.Args) {
				port = os.Args[i+1]
				i++
			}
		}
		if err := deploy.RunDaemon(cfg, port); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Commands: deploy, add, list, remove, restart, logs, daemon, version")
		os.Exit(1)
	}
}
