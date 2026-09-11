package deploy

import (
	"fmt"
	"path/filepath"

	"velo-deploy/internal/caddy"
	"velo-deploy/internal/config"
)

func rebuildRoutes(cfg *config.Config) error {
	for name := range cfg.Apps {
		_ = caddy.RemoveConfig(name)
	}
	var routes []caddy.Route
	for _, app := range cfg.Apps {
		route := caddy.Route{
			AppName: app.Name,
			Domain:  app.Domain,
			Path:    app.Path,
			Type:    app.Type,
		}
		if app.Type == config.AppTypeStatic {
			workDir := filepath.Join(cfg.AppsDir, app.Name)
			output := app.OutputDir
			if output == "" {
				output = "."
			}
			route.RootDir = filepath.Join(workDir, output)
		} else {
			upstreamHost := "127.0.0.1"
			if app.Alias != "" {
				upstreamHost = app.Alias
			}
			route.Upstream = fmt.Sprintf("%s:%d", upstreamHost, app.Port)
		}
		routes = append(routes, route)
	}
	return caddy.RebuildSharedConfig(routes)
}
