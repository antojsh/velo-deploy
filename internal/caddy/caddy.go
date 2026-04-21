package caddy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	Type      string // "node" | "static"
	RootDir   string // e.g. "/opt/deploy/apps/myapp/dist" (for static)
}

// GenerateConfig creates a Caddy config file for an app.
// If domain is empty, the app will be served via path-based routing
// under the shared catch-all config (:443 with tls internal).
func GenerateConfig(appName, domain, upstream string) error {
	os.MkdirAll(confDir, 0755)

	confPath := filepath.Join(confDir, appName+".conf")

	var conf string
	if domain != "" {
		// Domain → Let's Encrypt (automatic HTTPS)
		conf = fmt.Sprintf(`%s {
    reverse_proxy %s
}
`, domain, upstream)
	}

	return os.WriteFile(confPath, []byte(conf), 0644)
}

// GenerateStaticConfig creates a Caddy config file for a static site app.
func GenerateStaticConfig(appName, domain, rootDir string) error {
	os.MkdirAll(confDir, 0755)

	confPath := filepath.Join(confDir, appName+".conf")

	var conf string
	if domain != "" {
		conf = fmt.Sprintf(`%s {
    root * %s
    file_server
}
`, domain, rootDir)
	}

	return os.WriteFile(confPath, []byte(conf), 0644)
}

// RebuildSharedConfig regenerates the catch-all config for all apps without domains.
// Call this after adding/removing any app that has no domain.
func RebuildSharedConfig(routes []Route) error {
	os.MkdirAll(confDir, 0755)

	var sb strings.Builder
	sb.WriteString(":443 {\n")
	sb.WriteString("    tls internal\n\n")

	for _, r := range routes {
		if r.Type == "static" {
			sb.WriteString(fmt.Sprintf("    handle_path /%s/* {\n", r.AppName))
			sb.WriteString(fmt.Sprintf("        root * %s\n", r.RootDir))
			sb.WriteString("        file_server\n")
			sb.WriteString("    }\n\n")
		} else {
			sb.WriteString(fmt.Sprintf("    handle_path /%s/* {\n", r.AppName))
			sb.WriteString(fmt.Sprintf("        reverse_proxy %s\n", r.Upstream))
			sb.WriteString("    }\n\n")
		}
	}

	sb.WriteString("    handle {\n")
	sb.WriteString("        respond 404\n")
	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	sharedPath := filepath.Join(confDir, "_shared.conf")
	return os.WriteFile(sharedPath, []byte(sb.String()), 0644)
}

// RemoveSharedConfig removes the shared catch-all config.
func RemoveSharedConfig() error {
	sharedPath := filepath.Join(confDir, "_shared.conf")
	return os.Remove(sharedPath)
}

// RemoveConfig deletes a single app's Caddy config file.
func RemoveConfig(appName string) error {
	confPath := filepath.Join(confDir, appName+".conf")
	return os.Remove(confPath)
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
