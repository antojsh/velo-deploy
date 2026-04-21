package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"deploy/internal/caddy"
	"deploy/internal/config"
	"deploy/internal/hosts"
	deploynode "deploy/internal/node"
	"deploy/internal/systemd"
)

// Deploy performs a full deployment from a git repository.
func Deploy(cfg *config.Config, repoURL, domain, alias string) error {
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
		fmt.Printf("Warning: %v, using default v20\n", err)
		nodeVer = "20"
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

	// 7. Determine entry point
	entryPoint := detectEntryPoint(appDir)

	// 8. Assign port
	port := cfg.NextPort()
	if port == 0 {
		return fmt.Errorf("no free ports in range 3000-3999")
	}

	// 9. Create alias
	if alias == "" {
		alias = appName + ".local"
	}
	if err := hosts.AddAlias(alias); err != nil {
		fmt.Printf("Warning: could not add hosts alias: %v\n", err)
	}

	// 10. Create system user
	username := "deploy-" + appName
	if err := systemd.CreateUser(username); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Set correct ownership
	exec.Command("chown", "-R", username+":"+username, appDir).Run()

	// 11. Generate systemd service
	if err := systemd.GenerateService(appName, nodePath, appDir, entryPoint, username, username); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	// 12. Save app metadata
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
	}
	if err := cfg.Save(); err != nil {
		fmt.Printf("Warning: failed to save config: %v\n", err)
	}

	// 13. Configure Caddy
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

	// 14. Enable and start the service
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

		if err := rebuildSharedCaddyConfig(cfg); err != nil {
			return fmt.Errorf("failed to rebuild shared Caddy config: %w", err)
		}
		if err := caddy.Reload(); err != nil {
			return fmt.Errorf("failed to reload Caddy: %w", err)
		}

		cfg.Apps[appName] = meta
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
		fmt.Printf("Warning: %v, using default v20\n", err)
		nodeVer = "20"
	}

	if err := deploynode.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
		return fmt.Errorf("failed to install Node.js: %w", err)
	}

	nodePath, err := deploynode.GetNodePath(cfg.NVMDir, nodeVer)
	if err != nil {
		return fmt.Errorf("failed to find node path: %w", err)
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

	if err := systemd.GenerateService(appName, nodePath, appDir, entryPoint, username, username); err != nil {
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
	}

	upstream := fmt.Sprintf("%s:%d", alias, port)
	if domain != "" {
		if err := caddy.GenerateConfig(appName, domain, upstream); err != nil {
			return fmt.Errorf("failed to generate Caddy config: %w", err)
		}
	}

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

	cfg.Apps[appName] = meta
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

	if app.Type == config.AppTypeNode {
		systemd.StopApp(appName)
		systemd.RemoveService(appName)
		systemd.DaemonReload()
		systemd.RemoveUser("deploy-" + appName)
	}

	caddy.RemoveConfig(appName)
	hosts.RemoveAlias(app.Alias)

	if app.Type != config.AppTypeNode {
		os.RemoveAll(filepath.Join(cfg.AppsDir, appName))
	}

	delete(cfg.Apps, appName)
	cfg.Save()

	rebuildSharedCaddyConfig(cfg)
	caddy.Reload()

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
	// Strip .git suffix and get last path component
	repoURL = strings.TrimSuffix(repoURL, ".git")
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
