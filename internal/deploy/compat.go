package deploy

import (
	"fmt"
	"os"
	"strings"

	"velo-deploy/internal/caddy"
	"velo-deploy/internal/config"
	"velo-deploy/internal/hosts"
	"velo-deploy/internal/systemd"
)

type Options = DeployOptions

func (o DeployOptions) withDefaults() DeployOptions {
	if strings.TrimSpace(o.BuildCommand) == "" {
		o.BuildCommand = DefaultBuildCommand
	}
	if strings.TrimSpace(o.StartCommand) == "" {
		o.StartCommand = DefaultStartCommand
	}
	return o
}

func normalizePath(appName, domain, input string) (string, error) {
	path := strings.TrimSpace(input)
	if path == "" {
		return DefaultPath(appName, domain), nil
	}
	if strings.ContainsAny(path, " \t") {
		return "", fmt.Errorf("invalid path: %s", path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	return path, nil
}

func ensurePathAvailable(cfg *config.Config, appName, domain, path string) error {
	if cfg == nil {
		return nil
	}
	for name, app := range cfg.Apps {
		if name == appName {
			continue
		}
		if app.Domain == domain && app.Path == path {
			return fmt.Errorf("path %s on %s is already used", path, domain)
		}
	}
	return nil
}

func rebuildSharedCaddyConfig(cfg *config.Config) error {
	return rebuildRoutes(cfg)
}

var (
	stopApp            = systemd.StopApp
	disableApp         = systemd.DisableApp
	removeService      = systemd.RemoveService
	daemonReload       = systemd.DaemonReload
	resetFailed        = systemd.ResetFailed
	removeUser         = systemd.RemoveUser
	serviceExists      = systemd.ServiceExists
	removeCaddyConfig  = caddy.RemoveConfig
	reloadCaddy        = caddy.Reload
	removeHostAlias    = hosts.RemoveAlias
	removeManagedFiles = os.RemoveAll
)
