package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectVersionFromPackageJSON reads package.json and extracts the "engines.node" field.
// Returns something like "20" or ">=18.0.0".
func DetectVersionFromPackageJSON(workDir string) (string, error) {
	pkgPath := filepath.Join(workDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", fmt.Errorf("package.json not found: %w", err)
	}

	content := string(data)
	// Simple extraction without JSON parsing
	// Look for "engines": { "node": "20" }
	search := `"engines"`
	idx := strings.Index(content, search)
	if idx == -1 {
		// Default to LTS
		return "20", nil
	}

	// Find "node" key after engines
	sub := content[idx:]
	nodeIdx := strings.Index(sub, `"node"`)
	if nodeIdx == -1 {
		return "20", nil
	}

	sub = sub[nodeIdx:]
	// Find the value after "node":
	colonIdx := strings.Index(sub, ":")
	if colonIdx == -1 {
		return "20", nil
	}
	sub = sub[colonIdx+1:]

	// Find the quoted value
	startQuote := strings.Index(sub, `"`)
	if startQuote == -1 {
		return "20", nil
	}
	endQuote := strings.Index(sub[startQuote+1:], `"`)
	if endQuote == -1 {
		return "20", nil
	}

	ver := sub[startQuote+1 : startQuote+1+endQuote]
	// Clean up version specifiers like ">=20.0.0" -> "20"
	ver = strings.TrimPrefix(ver, ">=")
	ver = strings.TrimPrefix(ver, "^")
	ver = strings.TrimPrefix(ver, "~")

	// Extract major version
	major := strings.Split(ver, ".")[0]
	major = strings.Trim(major, " ")
	if major == "" {
		return "20", nil
	}
	return major, nil
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
	cmd := exec.Command("curl", "-fsSL", "-o", tmpFile, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Extract (strip top-level dir)
	cmd = exec.Command("tar", "-xJf", tmpFile, "--strip-components=1", "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	// Make world-readable/executable so system users can run node
	exec.Command("chmod", "-R", "a+rX", destDir).Run()

	fmt.Printf("Node.js v%s installed at %s\n", fullVersion, destDir)
	return nil
}

// resolveFullVersion resolves a major version like "20" to a full version like "20.19.1"
// by querying the Node.js dist index.
func resolveFullVersion(major string) (string, error) {
	out, err := exec.Command("bash", "-c",
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
	cmd := exec.Command(npmPath, "install", "--production")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Prepend node's bin dir to PATH so npm scripts can find 'node'
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
