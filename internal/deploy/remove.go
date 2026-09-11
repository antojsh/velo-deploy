package deploy

import (
	"fmt"
	"path/filepath"
	"strings"

	"velo-deploy/internal/config"
	"velo-deploy/internal/systemd"
)

func Remove(cfg *config.Config, name string) error {
	return RemoveApp(cfg, name, true)
}

func RemoveApp(cfg *config.Config, name string, purge bool) error {
	if cfg == nil || cfg.GetApp(name) == nil {
		return fmt.Errorf("app %s not found", name)
	}
	app := cfg.GetApp(name)
	var failures []string

	hasSystemdService := app.Type == config.AppTypeNode || serviceExists(name)
	if hasSystemdService {
		if err := stopApp(name); err != nil {
			failures = append(failures, "stop systemd service: "+err.Error())
		}
		if err := disableApp(name); err != nil {
			failures = append(failures, "disable systemd service: "+err.Error())
		}
		if err := removeService(name); err != nil {
			failures = append(failures, "remove systemd service: "+err.Error())
		}
		if err := daemonReload(); err != nil {
			failures = append(failures, "daemon-reload: "+err.Error())
		}
		if err := resetFailed(name); err != nil {
			failures = append(failures, "reset systemd failed state: "+err.Error())
		}
		for _, username := range []string{systemd.AppUser(name), "deploy-" + name} {
			if err := removeUser(username); err != nil {
				failures = append(failures, "remove user: "+err.Error())
			}
		}
	}
	if err := removeCaddyConfig(name); err != nil {
		failures = append(failures, "remove caddy config: "+err.Error())
	}
	if app.Alias != "" {
		if err := removeHostAlias(app.Alias); err != nil {
			failures = append(failures, "remove hosts alias: "+err.Error())
		}
	}
	delete(cfg.Apps, name)
	if err := rebuildSharedCaddyConfig(cfg); err != nil {
		failures = append(failures, "rebuild caddy routes: "+err.Error())
	}
	if err := reloadCaddy(); err != nil {
		failures = append(failures, "caddy reload: "+err.Error())
	}
	if purge {
		appDir := filepath.Join(cfg.AppsDir, name)
		if err := removeManagedFiles(appDir); err != nil {
			failures = append(failures, "remove files: "+err.Error())
		}
	}
	if err := cfg.Save(); err != nil {
		failures = append(failures, "save config: "+err.Error())
	}
	if len(failures) > 0 {
		return fmt.Errorf("failed to fully remove app '%s': %s", name, strings.Join(failures, "; "))
	}
	return nil
}
