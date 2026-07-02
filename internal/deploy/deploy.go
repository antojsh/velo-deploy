package deploy

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"velo-deploy/internal/caddy"
	"velo-deploy/internal/config"
	"velo-deploy/internal/hosts"
	deploynode "velo-deploy/internal/node"
	"velo-deploy/internal/systemd"
)

var (
	stopApp            = systemd.StopApp
	removeService      = systemd.RemoveService
	daemonReload       = systemd.DaemonReload
	removeUser         = systemd.RemoveUser
	removeCaddyConfig  = caddy.RemoveConfig
	reloadCaddy        = caddy.Reload
	removeHostAlias    = hosts.RemoveAlias
	removeManagedFiles = os.RemoveAll
)

const (
	DefaultBuildCommand = "npm run build"
	DefaultStartCommand = "npm run start"
)

type Options struct {
	BuildCommand string
	StartCommand string
}

func (o Options) withDefaults() Options {
	if o.BuildCommand == "" {
		o.BuildCommand = DefaultBuildCommand
	}
	if o.StartCommand == "" {
		o.StartCommand = DefaultStartCommand
	}
	return o
}

// Deploy performs a full deployment from a git repository.
func Deploy(cfg *config.Config, repoURL, domain, alias string) error {
	return DeployWithOptions(cfg, repoURL, domain, alias, Options{})
}

// DeployWithOptions performs a full deployment from a git repository.
func DeployWithOptions(cfg *config.Config, repoURL, domain, alias string, opts Options) error {
	opts = opts.withDefaults()

	// 1. Derive app name from repo URL
	appName := deriveAppName(repoURL)

	// 2. Check if already exists
	if _, exists := cfg.Apps[appName]; exists {
		fmt.Printf("App '%s' already exists. Removing first...\n", appName)
		if err := Remove(cfg, appName); err != nil {
			return fmt.Errorf("failed to remove existing app: %w", err)
		}
	}

	// 3. Clone repository
	appDir := filepath.Join(cfg.AppsDir, appName)
	os.MkdirAll(appDir, 0755)

	fmt.Printf("Cloning %s into %s...\n", repoURL, appDir)
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, appDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// 4. Detect Node.js version
	nodeVer, err := deploynode.DetectVersionFromPackageJSON(appDir)
	if err != nil {
		fmt.Printf("Warning: %v, using default v%s\n", err, deploynode.DefaultNodeVersion)
		nodeVer = deploynode.DefaultNodeVersion
	}
	fmt.Printf("Detected Node.js version: %s\n", nodeVer)

	// 5. Ensure Node.js is installed via nvm
	if err := deploynode.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
		return fmt.Errorf("failed to install Node.js: %w", err)
	}

	nodePath, err := deploynode.GetNodePath(cfg.NVMDir, nodeVer)
	if err != nil {
		return fmt.Errorf("failed to find node path: %w", err)
	}
	fmt.Printf("Node.js binary at: %s\n", nodePath)

	// 6. Install npm dependencies
	fmt.Println("Installing dependencies...")
	if err := deploynode.InstallDeps(appDir, nodePath); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	// 7. Build application
	fmt.Printf("Building application with: %s\n", opts.BuildCommand)
	if err := deploynode.RunCommand(appDir, nodePath, opts.BuildCommand); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// 8. Determine entry point for metadata/backwards compatibility
	entryPoint := detectEntryPoint(appDir)

	// 9. Assign port
	port := cfg.NextPort()
	if port == 0 {
		return fmt.Errorf("no free ports in range 3000-3999")
	}

	// 10. Create alias
	if alias == "" {
		alias = appName + ".local"
	}
	if err := hosts.AddAlias(alias); err != nil {
		fmt.Printf("Warning: could not add hosts alias: %v\n", err)
	}

	// 11. Create system user
	username := "deploy-" + appName
	if err := systemd.CreateUser(username); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Set correct ownership
	exec.Command("chown", "-R", username+":"+username, appDir).Run()

	// 12. Generate systemd service
	if err := systemd.GenerateCommandService(appName, nodePath, appDir, opts.StartCommand, username, username); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	// 13. Save app metadata
	cfg.Apps[appName] = &config.AppMeta{
		Name:       appName,
		Type:       config.AppTypeNode,
		RepoURL:    repoURL,
		Branch:     "main",
		NodeVer:    nodeVer,
		Port:       port,
		Domain:     domain,
		Alias:      alias,
		NodePath:   nodePath,
		EntryPoint: entryPoint,
		BuildCommand: opts.BuildCommand,
		StartCommand: opts.StartCommand,
	}
	if err := cfg.Save(); err != nil {
		fmt.Printf("Warning: failed to save config: %v\n", err)
	}

	// 14. Configure Caddy
	upstream := fmt.Sprintf("%s:%d", alias, port)

	if domain != "" {
		// Domain-based: dedicated Caddy config with Let's Encrypt
		if err := caddy.GenerateConfig(appName, domain, upstream); err != nil {
			return fmt.Errorf("failed to generate Caddy config: %w", err)
		}
	}

	// Always rebuild shared catch-all for apps without domains
	if err := rebuildSharedCaddyConfig(cfg); err != nil {
		return fmt.Errorf("failed to rebuild shared Caddy config: %w", err)
	}

	if err := caddy.Reload(); err != nil {
		return fmt.Errorf("failed to reload Caddy: %w", err)
	}

	// 15. Enable and start the service
	if err := systemd.DaemonReload(); err != nil {
		return fmt.Errorf("daemon-reload failed: %w", err)
	}
	if err := systemd.EnableApp(appName); err != nil {
		return fmt.Errorf("enable failed: %w", err)
	}
	if err := systemd.StartApp(appName); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	fmt.Printf("\n✅ %s deployed successfully!\n", appName)
	if domain != "" {
		fmt.Printf("   URL: https://%s\n", domain)
	} else {
		fmt.Printf("   URL: https://<tu-ip>/%s/\n", appName)
	}
	fmt.Printf("   Node: %s\n", nodeVer)

	return nil
}

// DetectAppType detects whether an app is "static" or "node" based on its files.
func DetectAppType(appDir string) string {
	if _, err := os.Stat(filepath.Join(appDir, "index.html")); err == nil {
		return config.AppTypeStatic
	}
	if _, err := os.Stat(filepath.Join(appDir, "package.json")); err == nil {
		return config.AppTypeStatic
	}
	return config.AppTypeNode
}

// DetectOutputDir detects the output directory for static sites.
func DetectOutputDir(appDir string) string {
	candidates := []string{"dist", "build", "_site", "public", "output"}
	for _, candidate := range candidates {
		path := filepath.Join(appDir, candidate)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

// Register adds an already-cloned app to deploy's management.
func Register(cfg *config.Config, appName, appDir, domain, alias, appType string) error {
	if _, exists := cfg.Apps[appName]; exists {
		return fmt.Errorf("app '%s' already exists. Use 'remove' first to re-register", appName)
	}

	info, err := os.Stat(appDir)
	if err != nil {
		return fmt.Errorf("app directory does not exist: %s", appDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", appDir)
	}

	if appType == "" {
		appType = DetectAppType(appDir)
	}

	var meta *config.AppMeta

	if appType == config.AppTypeStatic {
		outputDir := DetectOutputDir(appDir)
		if outputDir == "" {
			outputDir = "."
		}

		if alias == "" {
			alias = appName + ".local"
		}
		if err := hosts.AddAlias(alias); err != nil {
			fmt.Printf("Warning: could not add hosts alias: %v\n", err)
		}

		rootDir := appDir
		if outputDir != "." {
			rootDir = filepath.Join(appDir, outputDir)
		}

		meta = &config.AppMeta{
			Name:      appName,
			Type:      config.AppTypeStatic,
			Domain:    domain,
			Alias:     alias,
			OutputDir: outputDir,
		}

		if domain != "" {
			if err := caddy.GenerateStaticConfig(appName, domain, rootDir); err != nil {
				return fmt.Errorf("failed to generate Caddy config: %w", err)
			}
		}

		cfg.Apps[appName] = meta
		if err := rebuildSharedCaddyConfig(cfg); err != nil {
			return fmt.Errorf("failed to rebuild shared Caddy config: %w", err)
		}
		if err := caddy.Reload(); err != nil {
			return fmt.Errorf("failed to reload Caddy: %w", err)
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to save config: %v\n", err)
		}

		fmt.Printf("\n✅ %s registered successfully (static)!\n", appName)
		if domain != "" {
			fmt.Printf("   URL: https://%s\n", domain)
		} else {
			fmt.Printf("   URL: https://<tu-ip>/%s/\n", appName)
		}
		fmt.Printf("   Output: %s\n", outputDir)
		return nil
	}

	nodeVer, err := deploynode.DetectVersionFromPackageJSON(appDir)
	if err != nil {
		fmt.Printf("Warning: %v, using default v%s\n", err, deploynode.DefaultNodeVersion)
		nodeVer = deploynode.DefaultNodeVersion
	}

	if err := deploynode.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
		return fmt.Errorf("failed to install Node.js: %w", err)
	}

	nodePath, err := deploynode.GetNodePath(cfg.NVMDir, nodeVer)
	if err != nil {
		return fmt.Errorf("failed to find node path: %w", err)
	}

	if err := deploynode.InstallDeps(appDir, nodePath); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	if err := deploynode.RunCommand(appDir, nodePath, DefaultBuildCommand); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	port := cfg.NextPort()
	if port == 0 {
		return fmt.Errorf("no free ports in range 3000-3999")
	}

	entryPoint := detectEntryPoint(appDir)

	if alias == "" {
		alias = appName + ".local"
	}
	if err := hosts.AddAlias(alias); err != nil {
		fmt.Printf("Warning: could not add hosts alias: %v\n", err)
	}

	username := "deploy-" + appName
	if err := systemd.CreateUser(username); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	exec.Command("chown", "-R", username+":"+username, appDir).Run()

	if err := systemd.GenerateCommandService(appName, nodePath, appDir, DefaultStartCommand, username, username); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	meta = &config.AppMeta{
		Name:       appName,
		Type:       config.AppTypeNode,
		Port:       port,
		Domain:     domain,
		Alias:      alias,
		NodePath:   nodePath,
		NodeVer:    nodeVer,
		EntryPoint: entryPoint,
		BuildCommand: DefaultBuildCommand,
		StartCommand: DefaultStartCommand,
	}

	upstream := fmt.Sprintf("%s:%d", alias, port)
	if domain != "" {
		if err := caddy.GenerateConfig(appName, domain, upstream); err != nil {
			return fmt.Errorf("failed to generate Caddy config: %w", err)
		}
	}

	cfg.Apps[appName] = meta
	if err := rebuildSharedCaddyConfig(cfg); err != nil {
		return fmt.Errorf("failed to rebuild shared Caddy config: %w", err)
	}
	if err := caddy.Reload(); err != nil {
		return fmt.Errorf("failed to reload Caddy: %w", err)
	}

	if err := systemd.DaemonReload(); err != nil {
		return fmt.Errorf("daemon-reload failed: %w", err)
	}
	if err := systemd.EnableApp(appName); err != nil {
		return fmt.Errorf("enable failed: %w", err)
	}
	if err := systemd.StartApp(appName); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Warning: failed to save config: %v\n", err)
	}

	fmt.Printf("\n✅ %s registered successfully (node)!\n", appName)
	if domain != "" {
		fmt.Printf("   URL: https://%s\n", domain)
	} else {
		fmt.Printf("   URL: https://<tu-ip>/%s/\n", appName)
	}
	fmt.Printf("   Node: %s\n", nodeVer)

	return nil
}

// Remove removes a deployed app entirely.
func Remove(cfg *config.Config, appName string) error {
	app, exists := cfg.Apps[appName]
	if !exists {
		return fmt.Errorf("app '%s' not found", appName)
	}

	fmt.Printf("Removing %s...\n", appName)
	var cleanupErrs []error

	if app.Type == config.AppTypeNode {
		if err := stopApp(appName); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("stop systemd service: %w", err))
		}
		if err := removeService(appName); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("remove systemd service: %w", err))
		}
		if err := daemonReload(); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("reload systemd daemon: %w", err))
		}
		if err := removeUser("deploy-" + appName); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("remove system user: %w", err))
		}
	}

	if err := removeCaddyConfig(appName); err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("remove Caddy config: %w", err))
	}
	if app.Alias != "" {
		if err := removeHostAlias(app.Alias); err != nil {
			cleanupErrs = append(cleanupErrs, fmt.Errorf("remove hosts alias: %w", err))
		}
	}
	if err := removeManagedFiles(filepath.Join(cfg.AppsDir, appName)); err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("remove app files: %w", err))
	}

	delete(cfg.Apps, appName)
	if err := cfg.Save(); err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("save config: %w", err))
	}

	if err := rebuildSharedCaddyConfig(cfg); err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("rebuild shared Caddy config: %w", err))
	}
	if err := reloadCaddy(); err != nil {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("reload Caddy: %w", err))
	}

	if err := errors.Join(cleanupErrs...); err != nil {
		return fmt.Errorf("failed to fully remove app '%s': %w", appName, err)
	}

	fmt.Printf("✅ %s removed.\n", appName)
	return nil
}

// rebuildSharedCaddyConfig regenerates the catch-all config for all apps without domains.
func rebuildSharedCaddyConfig(cfg *config.Config) error {
	var routes []caddy.Route
	for _, app := range cfg.Apps {
		if app.Domain == "" {
			route := caddy.Route{
				AppName:  app.Name,
				Upstream: fmt.Sprintf("%s:%d", app.Alias, app.Port),
				Type:     app.Type,
				RootDir:  filepath.Join(cfg.AppsDir, app.Name, app.OutputDir),
			}
			routes = append(routes, route)
		}
	}

	if len(routes) == 0 {
		caddy.RemoveSharedConfig()
		return nil
	}

	return caddy.RebuildSharedConfig(routes)
}

func deriveAppName(repoURL string) string {
	// Strip .git suffix and any trailing slash, then take the last path component.
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimRight(repoURL, "/")
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		// Replace non-alphanumeric with hyphens
		var clean strings.Builder
		for _, c := range name {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				clean.WriteRune(c)
			} else if clean.Len() > 0 && clean.String()[clean.Len()-1] != '-' {
				clean.WriteRune('-')
			}
		}
		result := clean.String()
		if result != "" {
			return strings.ToLower(result)
		}
	}
	return "app"
}

func detectEntryPoint(workDir string) string {
	// Check common entry points
	candidates := []string{"index.js", "server.js", "app.js", "main.js", "src/index.js", "dist/index.js"}
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(workDir, candidate)); err == nil {
			return candidate
		}
	}
	// Default
	return "index.js"
}
