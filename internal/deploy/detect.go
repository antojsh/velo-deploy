package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"velo-deploy/internal/config"
)

var (
	ErrGit      = errors.New("git operation failed")
	ErrBuild    = errors.New("build failed")
	ErrPort     = errors.New("port already in use")
	ErrNotFound = errors.New("app not found")
	ErrHealth   = errors.New("service did not become healthy")
	ErrName     = errors.New("invalid app name")
)

var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

var outputCandidates = []string{"dist", "build", "_site", "public", "output", "out", ".output/public"}

const (
	DefaultBuildCommand = "npm run build"
	DefaultStartCommand = "npm run start"
)

type packageJSON struct {
	Name    string            `json:"name"`
	Main    string            `json:"main"`
	Scripts map[string]string `json:"scripts"`
}

func DetectAppType(workDir string, typeHint ...string) string {
	hint := ""
	if len(typeHint) > 0 {
		hint = typeHint[0]
	}
	switch hint {
	case config.AppTypeNode, config.AppTypeStatic:
		return hint
	}
	pkg := readPackageJSON(workDir)
	hasIndex := fileExists(filepath.Join(workDir, "index.html"))
	hasStart := pkg != nil && strings.TrimSpace(pkg.Scripts["start"]) != ""
	if hasStart {
		return config.AppTypeNode
	}
	if hasIndex || pkg != nil {
		return config.AppTypeStatic
	}
	return config.AppTypeNode
}

func AppNameFromRepo(repoURL string) string {
	trimmed := strings.TrimSpace(repoURL)
	trimmed = strings.TrimSuffix(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, ".git")
	parts := strings.Split(trimmed, "/")
	name := parts[len(parts)-1]
	if idx := strings.LastIndex(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	name = strings.Trim(b.String(), "-")
	if name == "" {
		name = "app"
	}
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

func deriveAppName(repoURL string) string {
	return AppNameFromRepo(repoURL)
}

func DetectOutputDir(workDir string, override ...string) string {
	if len(override) > 0 && override[0] != "" {
		return override[0]
	}
	for _, candidate := range outputCandidates {
		info, err := os.Stat(filepath.Join(workDir, candidate))
		if err == nil && info.IsDir() {
			return candidate
		}
	}
	if fileExists(filepath.Join(workDir, "index.html")) {
		return "."
	}
	return ""
}

func DetectEntryPoint(workDir string) string {
	pkg := readPackageJSON(workDir)
	if pkg != nil && strings.TrimSpace(pkg.Main) != "" && fileExists(filepath.Join(workDir, pkg.Main)) {
		return pkg.Main
	}
	for _, candidate := range []string{"index.js", "server.js", "app.js", "main.js", "src/index.js", "dist/index.js", "dist/main.js"} {
		if fileExists(filepath.Join(workDir, candidate)) {
			return candidate
		}
	}
	return "index.js"
}

func detectEntryPoint(workDir string) string {
	return DetectEntryPoint(workDir)
}

func SanitizeName(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if !nameRe.MatchString(name) {
		return "", fmt.Errorf("%w: %s", ErrName, name)
	}
	return name, nil
}

func DefaultPath(appName, domain string) string {
	if domain != "" {
		return "/"
	}
	return "/" + appName
}

func DefaultAlias(appName, alias string) string {
	if strings.TrimSpace(alias) != "" {
		return alias
	}
	return appName + ".local"
}

func NormalizeRepoURL(raw string) string {
	u := strings.TrimSpace(strings.ToLower(raw))
	u = strings.TrimSuffix(u, ".git")
	u = strings.TrimSuffix(u, "/")
	u = strings.TrimPrefix(u, "git+")
	if strings.HasPrefix(u, "git@") {
		u = strings.TrimPrefix(u, "git@")
		u = strings.Replace(u, ":", "/", 1)
	}
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "ssh://git@")
	u = strings.TrimPrefix(u, "ssh://")
	return u
}

func ReposMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return NormalizeRepoURL(a) == NormalizeRepoURL(b)
}

func BranchFromRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}

func ShouldDeployBranch(appBranch, pushedBranch string) bool {
	if pushedBranch == "" {
		return false
	}
	if appBranch != "" {
		return appBranch == pushedBranch
	}
	return pushedBranch == "main" || pushedBranch == "master"
}

func readPackageJSON(workDir string) *packageJSON {
	data, err := os.ReadFile(filepath.Join(workDir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg packageJSON
	if json.Unmarshal(data, &pkg) != nil {
		return nil
	}
	if pkg.Scripts == nil {
		pkg.Scripts = map[string]string{}
	}
	return &pkg
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasBuildScript(workDir string) bool {
	pkg := readPackageJSON(workDir)
	return pkg != nil && strings.TrimSpace(pkg.Scripts["build"]) != ""
}

func hasStartScript(workDir string) bool {
	pkg := readPackageJSON(workDir)
	return pkg != nil && strings.TrimSpace(pkg.Scripts["start"]) != ""
}
