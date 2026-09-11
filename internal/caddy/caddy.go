package caddy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var confDir = "/etc/caddy/conf.d"
var snippetDir = "/etc/velo-deploy/caddy"
var caddyfile = "/etc/caddy/Caddyfile"
var execCommand = exec.Command
var lookPath = exec.LookPath

func SetConfDir(path string) {
	confDir = path
}

func SetSnippetDir(path string) {
	snippetDir = path
}

// Route represents a single app routing rule.
type Route struct {
	AppName  string
	Upstream string // e.g. "myapp.local:3000"
	Domain   string // empty = path-based
	Path     string // e.g. "/myapp"
	Type     string // "node" | "static"
	RootDir  string // e.g. "/opt/deploy/apps/myapp/dist" (for static)
}

// GenerateConfig creates a Caddy config file for an app.
// If domain is empty, the app is served via path-based routing under
// the shared catch-all config (:443 with tls internal), so no per-app
// file is created.
func GenerateConfig(appName, domain, upstream string) error {
	if domain == "" {
		return nil
	}
	os.MkdirAll(confDir, 0755)

	confPath := filepath.Join(confDir, appName+".conf")
	conf := fmt.Sprintf("%s {\n    reverse_proxy %s\n%s}\n", domain, upstream, snippetImport(appName))
	return os.WriteFile(confPath, []byte(conf), 0644)
}

// GenerateStaticConfig creates a Caddy config file for a static site app.
// If domain is empty, no per-app file is created (path-based routing only).
func GenerateStaticConfig(appName, domain, rootDir string) error {
	if domain == "" {
		return nil
	}
	os.MkdirAll(confDir, 0755)

	confPath := filepath.Join(confDir, appName+".conf")
	conf := fmt.Sprintf("%s {\n    root * %s\n    file_server\n%s}\n", domain, rootDir, snippetImport(appName))
	return os.WriteFile(confPath, []byte(conf), 0644)
}

// RebuildSharedConfig regenerates the catch-all config for all apps without domains.
// Call this after adding/removing any app that has no domain.
// If routes is empty, the shared file is removed instead of regenerated.
func RebuildSharedConfig(routes []Route) error {
	os.MkdirAll(confDir, 0755)

	sharedPath := filepath.Join(confDir, "_shared.conf")
	if len(routes) == 0 {
		if err := os.Remove(sharedPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	grouped := make(map[string][]Route)
	for _, r := range routes {
		grouped[r.Domain] = append(grouped[r.Domain], r)
	}

	var domains []string
	for domain := range grouped {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	var sb strings.Builder
	for _, domain := range domains {
		site := domain
		if site == "" {
			site = ":443"
		}
		sb.WriteString(site + " {\n")
		if domain == "" {
			sb.WriteString("    tls internal\n\n")
		}

		domainRoutes := grouped[domain]
		sort.Slice(domainRoutes, func(i, j int) bool {
			return routePath(domainRoutes[i]) < routePath(domainRoutes[j])
		})
		for _, r := range domainRoutes {
			writeRoute(&sb, r)
		}

		sb.WriteString("    handle {\n")
		sb.WriteString("        respond 404\n")
		sb.WriteString("    }\n")
		sb.WriteString("}\n\n")
	}

	return os.WriteFile(sharedPath, []byte(sb.String()), 0644)
}

func writeRoute(sb *strings.Builder, r Route) {
	path := routePath(r)
	if path == "/" {
		sb.WriteString("    handle {\n")
	} else {
		sb.WriteString(fmt.Sprintf("    handle_path %s/* {\n", path))
	}
	if r.Type == "static" {
		sb.WriteString(fmt.Sprintf("        root * %s\n", r.RootDir))
		sb.WriteString("        file_server\n")
	} else {
		sb.WriteString(fmt.Sprintf("        reverse_proxy %s\n", r.Upstream))
	}
	sb.WriteString(snippetImportIndented(r.AppName, "        "))
	sb.WriteString("    }\n\n")
}

func snippetImport(appName string) string {
	path := filepath.ToSlash(filepath.Join(snippetDir, appName+".snippet"))
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return fmt.Sprintf("    import %s\n", path)
}

func snippetImportIndented(appName, indent string) string {
	path := filepath.ToSlash(filepath.Join(snippetDir, appName+".snippet"))
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return fmt.Sprintf("%simport %s\n", indent, path)
}

func routePath(r Route) string {
	if r.Path != "" {
		return r.Path
	}
	return "/" + r.AppName
}

// RemoveSharedConfig removes the shared catch-all config.
// Missing file is not an error (idempotent).
func RemoveSharedConfig() error {
	sharedPath := filepath.Join(confDir, "_shared.conf")
	if err := os.Remove(sharedPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// RemoveConfig deletes a single app's Caddy config file.
// Missing file is not an error (idempotent).
func RemoveConfig(appName string) error {
	confPath := filepath.Join(confDir, appName+".conf")
	if err := os.Remove(confPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func Reload() error {
	if _, err := lookPath("caddy"); err != nil {
		return nil
	}
	args := []string{"reload"}
	if _, err := os.Stat(caddyfile); err == nil {
		args = []string{"reload", "--config", caddyfile}
	}
	cmd := execCommand("caddy", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("caddy reload failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
