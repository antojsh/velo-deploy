package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	execCommand = exec.Command
)

// DefaultNodeVersion is the Node.js major version used when package.json does
// not define engines.node or defines a malformed value.
const DefaultNodeVersion = "24"

// DetectVersionFromPackageJSON reads package.json and extracts the "engines.node"
// field. Returns the major version as a string (e.g., "24") or the configured
// default when the field is missing/empty/malformed. Returns an error
// only when the file is missing or the JSON is invalid.
func DetectVersionFromPackageJSON(workDir string) (string, error) {
	pkgPath := filepath.Join(workDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", fmt.Errorf("package.json not found: %w", err)
	}

	var pkg struct {
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("invalid package.json: %w", err)
	}

	raw := strings.TrimSpace(pkg.Engines.Node)
	if raw == "" {
		return DefaultNodeVersion, nil
	}

	// Strip common range operators: >=, <=, >, <, ^, ~, =
	for _, op := range []string{">=", "<=", ">", "<", "^", "~", "="} {
		raw = strings.TrimPrefix(raw, op)
	}
	raw = strings.TrimSpace(raw)

	// Extract the leading numeric part (the major version).
	end := 0
	for end < len(raw) && unicode.IsDigit(rune(raw[end])) {
		end++
	}
	if end == 0 {
		return DefaultNodeVersion, nil
	}
	return raw[:end], nil
}

// nodeInstallBase is the base directory for all node installations.
const nodeInstallBase = "/opt/deploy/node"

// EnsureInstalled checks if the given Node.js major version is installed.
// If not, downloads and installs the official tarball to /opt/deploy/node/<version>/
// with world-readable permissions so any system user can execute it.
func EnsureInstalled(nvmDir, version string) error {
	nodePath, err := GetNodePath(nvmDir, version)
	if err == nil && nodePath != "" {
		return nil // Already installed
	}

	fmt.Printf("Installing Node.js v%s...\n", version)

	// Resolve full LTS version number via unofficial dist index
	fullVersion, err := resolveFullVersion(version)
	if err != nil {
		// Fallback: use major version directly
		fullVersion = version
	}

	arch := "linux-x64"
	tarName := fmt.Sprintf("node-v%s-%s.tar.xz", fullVersion, arch)
	url := fmt.Sprintf("https://nodejs.org/dist/v%s/%s", fullVersion, tarName)
	destDir := filepath.Join(nodeInstallBase, version)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	// Download
	fmt.Printf("Downloading %s...\n", url)
	tmpFile := fmt.Sprintf("/tmp/%s", tarName)
	cmd := execCommand("curl", "-fsSL", "-o", tmpFile, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Extract (strip top-level dir)
	cmd = execCommand("tar", "-xJf", tmpFile, "--strip-components=1", "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	// Make world-readable/executable so system users can run node
	execCommand("chmod", "-R", "a+rX", destDir).Run()

	fmt.Printf("Node.js v%s installed at %s\n", fullVersion, destDir)
	return nil
}

// resolveFullVersion resolves a major version like "20" to a full version like "20.19.1"
// by querying the Node.js dist index.
func resolveFullVersion(major string) (string, error) {
	out, err := execCommand("bash", "-c",
		fmt.Sprintf(`curl -fsSL https://nodejs.org/dist/index.json | grep -o '"version":"v%s\.[^"]*"' | head -1 | grep -o '[0-9][^"]*'`, major),
	).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("could not resolve version")
	}
	return strings.TrimSpace(string(out)), nil
}

// GetNodePath returns the absolute path to the node binary for a given major version.
// Checks /opt/deploy/node/<version>/bin/node first, then falls back to nvm layout.
func GetNodePath(nvmDir, version string) (string, error) {
	// Primary: /opt/deploy/node/<major>/bin/node
	primary := filepath.Join(nodeInstallBase, version, "bin", "node")
	if _, err := os.Stat(primary); err == nil {
		return primary, nil
	}

	// Fallback: nvm layout (legacy / manual installs)
	for _, pattern := range []string{
		filepath.Join(nvmDir, "versions", "node", "v"+version+"*", "bin", "node"),
		filepath.Join(nvmDir, "versions", "node", version+"*", "bin", "node"),
	} {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0], nil
		}
	}

	return "", fmt.Errorf("node v%s not found", version)
}

// InstallDeps runs npm install in the given directory.
func InstallDeps(workDir, nodePath string) error {
	nodeDir := filepath.Dir(nodePath)
	npmPath := filepath.Join(nodeDir, "npm")
	cmd := exec.Command(npmPath, "install")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Prepend node's bin dir to PATH so npm scripts can find 'node'
	cmd.Env = append(os.Environ(), "PATH="+nodeDir+":"+os.Getenv("PATH"))
	return cmd.Run()
}

// RunCommand runs a package command in the given directory with Node's bin dir
// at the front of PATH. This is used for build commands such as "npm run build".
func RunCommand(workDir, nodePath, command string) error {
	nodeDir := filepath.Dir(nodePath)
	cmd := exec.Command("bash", "-lc", command)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PATH="+nodeDir+":"+os.Getenv("PATH"))
	return cmd.Run()
}

// RunApp runs the Node.js application with the given node binary.
func RunApp(nodePath, workDir, entryPoint string) error {
	cmd := exec.Command(nodePath, entryPoint)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
