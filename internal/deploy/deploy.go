package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"velo-deploy/internal/config"
	"velo-deploy/internal/hosts"
	"velo-deploy/internal/node"
	"velo-deploy/internal/systemd"
)

type DeployOptions struct {
	RepoURL      string
	Name         string
	Domain       string
	Path         string
	Alias        string
	Branch       string
	Port         int
	OutputDir    string
	BuildCommand string
	StartCommand string
	Type         string
	Force        bool
	LocalDir     string
	DuckDNS      bool
}

func Deploy(cfg *config.Config, repoURL, domain, alias string) error {
	return DeployWithOptions(cfg, DeployOptions{
		RepoURL: repoURL,
		Domain:  domain,
		Alias:   alias,
	})
}

func Register(cfg *config.Config, appName, appDir, domain, alias, typeHint string) error {
	return DeployWithOptions(cfg, DeployOptions{
		Name:     appName,
		LocalDir: appDir,
		Domain:   domain,
		Alias:    alias,
		Type:     typeHint,
	})
}

func DeployWithOptions(cfg *config.Config, opts DeployOptions) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if cfg.Apps == nil {
		cfg.Apps = map[string]*config.AppMeta{}
	}
	if cfg.AppsDir == "" {
		cfg.AppsDir = config.AppsDir
	}
	if cfg.NVMDir == "" {
		cfg.NVMDir = config.NVMDir
	}

	name := opts.Name
	if name == "" {
		name = AppNameFromRepo(opts.RepoURL)
		if name == "app" && opts.LocalDir != "" {
			name = filepath.Base(opts.LocalDir)
		}
	}
	sanitized, err := SanitizeName(name)
	if err != nil {
		return err
	}
	name = sanitized

	workDir := opts.LocalDir
	if workDir == "" {
		workDir = filepath.Join(cfg.AppsDir, name)
	}

	existing := cfg.GetApp(name)
	previousSHA := ""
	if existing != nil && gitIsRepo(workDir) && !opts.Force {
		previousSHA = gitHead(workDir)
	}

	if opts.LocalDir == "" {
		if err := os.MkdirAll(cfg.AppsDir, 0755); err != nil {
			return err
		}
		if opts.Force {
			_ = os.RemoveAll(workDir)
		}
		if _, err := os.Stat(workDir); err == nil && gitIsRepo(workDir) {
			if err := gitPull(workDir, opts.Branch); err != nil {
				return err
			}
		} else {
			if opts.RepoURL == "" {
				return fmt.Errorf("%w: repo url is required", ErrGit)
			}
			if err := gitClone(opts.RepoURL, workDir, opts.Branch); err != nil {
				return err
			}
		}
		if !opts.Force && previousSHA != "" && previousSHA == gitHead(workDir) && existing != nil {
			return applyRuntime(cfg, existing, workDir)
		}
	} else {
		info, err := os.Stat(workDir)
		if err != nil {
			return fmt.Errorf("app directory does not exist: %s", workDir)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", workDir)
		}
		if existing != nil {
			return fmt.Errorf("app %s already exists", name)
		}
	}

	appType := DetectAppType(workDir, opts.Type)
	resolved := resolveDomain(cfg, name, opts.Domain, opts.DuckDNS)
	if opts.Domain == "" {
		opts.Domain = resolved.Domain
	}
	if opts.Alias == "" {
		opts.Alias = resolved.Alias
	}
	alias := DefaultAlias(name, opts.Alias)
	urlPath, err := normalizePath(name, opts.Domain, opts.Path)
	if err != nil {
		return err
	}
	if err := ensurePathAvailable(cfg, name, opts.Domain, urlPath); err != nil {
		return err
	}

	meta := &config.AppMeta{
		Name:         name,
		Type:         appType,
		RepoURL:      opts.RepoURL,
		Branch:       opts.Branch,
		Domain:       opts.Domain,
		Path:         urlPath,
		Alias:        alias,
		BuildCommand: opts.BuildCommand,
		StartCommand: opts.StartCommand,
		OutputDir:    opts.OutputDir,
	}
	if existing != nil {
		if meta.RepoURL == "" {
			meta.RepoURL = existing.RepoURL
		}
		if meta.Branch == "" {
			meta.Branch = existing.Branch
		}
		if meta.BuildCommand == "" {
			meta.BuildCommand = existing.BuildCommand
		}
		if meta.StartCommand == "" {
			meta.StartCommand = existing.StartCommand
		}
	}

	if appType == config.AppTypeNode {
		port := opts.Port
		if port == 0 && existing != nil {
			port = existing.Port
		}
		if port == 0 {
			port = cfg.NextPort()
		}
		if port == 0 {
			return ErrPort
		}
		for otherName, other := range cfg.Apps {
			if otherName != name && other.Type == config.AppTypeNode && other.Port == port {
				return ErrPort
			}
		}
		meta.Port = port

		nodeVer, err := node.DetectVersionFromPackageJSON(workDir)
		if err != nil {
			nodeVer = node.DefaultNodeVersion
		}
		meta.NodeVer = nodeVer
		if err := node.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
			return err
		}
		nodePath, err := node.GetNodePath(cfg.NVMDir, nodeVer)
		if err != nil {
			return err
		}
		meta.NodePath = nodePath
		meta.EntryPoint = DetectEntryPoint(workDir)

		if fileExists(filepath.Join(workDir, "package.json")) {
			if err := node.InstallDeps(workDir, nodePath); err != nil {
				return fmt.Errorf("%w: npm install: %v", ErrBuild, err)
			}
			buildCmd := meta.BuildCommand
			if buildCmd == "" && hasBuildScript(workDir) {
				buildCmd = "npm run build"
			}
			if buildCmd != "" {
				if err := node.RunCommand(workDir, nodePath, buildCmd); err != nil {
					return fmt.Errorf("%w: %s: %v", ErrBuild, buildCmd, err)
				}
			}
		}
	} else {
		if fileExists(filepath.Join(workDir, "package.json")) {
			nodeVer, err := node.DetectVersionFromPackageJSON(workDir)
			if err != nil {
				nodeVer = node.DefaultNodeVersion
			}
			if err := node.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
				return err
			}
			nodePath, err := node.GetNodePath(cfg.NVMDir, nodeVer)
			if err != nil {
				return err
			}
			if err := node.InstallDeps(workDir, nodePath); err != nil {
				return fmt.Errorf("%w: npm install: %v", ErrBuild, err)
			}
			buildCmd := meta.BuildCommand
			if buildCmd == "" && hasBuildScript(workDir) {
				buildCmd = "npm run build"
			}
			if buildCmd != "" {
				if err := node.RunCommand(workDir, nodePath, buildCmd); err != nil {
					return fmt.Errorf("%w: %s: %v", ErrBuild, buildCmd, err)
				}
			}
		}
		meta.OutputDir = DetectOutputDir(workDir, meta.OutputDir)
	}

	cfg.Apps[name] = meta
	if err := applyRuntime(cfg, meta, workDir); err != nil {
		return err
	}
	return cfg.Save()
}

func applyRuntime(cfg *config.Config, meta *config.AppMeta, workDir string) error {
	username := systemd.AppUser(meta.Name)
	if meta.Type == config.AppTypeNode {
		if err := systemd.CreateUser(username); err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		_ = chownRecursive(workDir, username)
		startCmd := meta.StartCommand
		if startCmd == "" && hasStartScript(workDir) {
			startCmd = "npm run start"
		}
		if err := systemd.GenerateAppService(meta.Name, meta.NodePath, workDir, meta.EntryPoint, startCmd, username, username, meta.Port); err != nil {
			return err
		}
		if err := systemd.DaemonReload(); err != nil {
			return err
		}
		_ = systemd.EnableApp(meta.Name)
		if err := systemd.RestartApp(meta.Name); err != nil {
			if err := systemd.StartApp(meta.Name); err != nil {
				return err
			}
		}
		if runtime.GOOS == "linux" {
			if err := systemd.WaitHealthy(meta.Name); err != nil {
				return fmt.Errorf("%w: %v", ErrHealth, err)
			}
		}
	}

	if err := rebuildRoutes(cfg); err != nil {
		return err
	}
	if meta.Alias != "" {
		_ = hosts.AddAlias(meta.Alias)
	}
	if err := reloadCaddy(); err != nil {
		return err
	}
	return nil
}

func Redeploy(cfg *config.Config, app *config.AppMeta) error {
	if app == nil {
		return ErrNotFound
	}
	return DeployWithOptions(cfg, DeployOptions{
		RepoURL:      app.RepoURL,
		Name:         app.Name,
		Domain:       app.Domain,
		Path:         app.Path,
		Alias:        app.Alias,
		Branch:       app.Branch,
		Port:         app.Port,
		OutputDir:    app.OutputDir,
		BuildCommand: app.BuildCommand,
		StartCommand: app.StartCommand,
		Type:         app.Type,
		LocalDir:     "",
		Force:        false,
	})
}

func chownRecursive(path, username string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	cmd := exec.Command("chown", "-R", username+":"+username, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func NeedRoot() bool {
	return runtime.GOOS == "linux" && os.Geteuid() != 0
}
