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

// SetConfDir sets the Caddy config directory (for testing)
func SetConfDir(path string) {
	confDir = path
}

// Route represents a single app routing rule.
type Route struct {
	AppName   string
	Upstream  string // e.g. "myapp.local:3000"
	Domain    string // empty = path-based
	Path      string // e.g. "/myapp"
	Type      string // "node" | "static"
	RootDir   string // e.g. "/opt/deploy/apps/myapp/dist" (for static)
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
	conf := fmt.Sprintf(`%s {
    reverse_proxy %s
}
`, domain, upstream)
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
	conf := fmt.Sprintf(`%s {
    root * %s
    file_server
}
`, domain, rootDir)
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
	sb.WriteString("    }\n\n")
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

// Reload tells Caddy to reload its configuration.
// Returns an error but callers should treat it as a warning — the app is
// already deployed; Caddy may simply not be running yet.
func Reload() error {
	cmd := exec.Command("caddy", "reload")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Warning: caddy reload failed (is Caddy running?): %s\n", strings.TrimSpace(string(out)))
	}
	return nil
}
