package node

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectVersionFromPackageJSON_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","version":"1.0.0","engines":{"node":">=18.0.0"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "18", version)
}

func TestDetectVersionFromPackageJSON_CaretVersion(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"node":"^20.15.0"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_TildeVersion(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"node":"~18.20.0"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "18", version)
}

func TestDetectVersionFromPackageJSON_SimpleVersion(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"node":"20"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_NoEngines(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","version":"1.0.0"}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_NoNodeKey(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"npm":"^9.0.0"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_EmptyEngines(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("not json"), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.Error(t, err)
	assert.Equal(t, "", version)
}

func TestDetectVersionFromPackageJSON_NoPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.Error(t, err)
	assert.Equal(t, "", version)
}

func TestDetectVersionFromPackageJSON_MalformedNodeValue(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"node":":}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_EnginesBeforeNode(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"node":"16","npm":"^9.0.0"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "16", version)
}

func TestGetNodePath_PrimaryExists(t *testing.T) {
	tmpDir := t.TempDir()
	nodeBinDir := filepath.Join(tmpDir, "20", "bin")
	err := os.MkdirAll(nodeBinDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(nodeBinDir, "node"), []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	path, err := GetNodePath(tmpDir, "20")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "20", "bin", "node"), path)
}

func TestGetNodePath_PrimaryNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	path, err := GetNodePath(tmpDir, "20")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}

func TestGetNodePath_NVMLayout(t *testing.T) {
	tmpDir := t.TempDir()
	nvmDir := filepath.Join(tmpDir, "nvm")
	nodeVersionsDir := filepath.Join(nvmDir, "versions", "node")
	v20Dir := filepath.Join(nodeVersionsDir, "v20.0.0", "bin")
	err := os.MkdirAll(v20Dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(v20Dir, "node"), []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	path, err := GetNodePath(nvmDir, "20")
	assert.NoError(t, err)
	assert.Contains(t, path, "v20")
	assert.Contains(t, path, "node")
}

func TestGetNodePath_NVMLayoutWithoutV(t *testing.T) {
	tmpDir := t.TempDir()
	nvmDir := filepath.Join(tmpDir, "nvm")
	nodeVersionsDir := filepath.Join(nvmDir, "versions", "node")
	v18Dir := filepath.Join(nodeVersionsDir, "18.0.0", "bin")
	err := os.MkdirAll(v18Dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(v18Dir, "node"), []byte("#!/bin/bash"), 0755)
	require.NoError(t, err)

	path, err := GetNodePath(nvmDir, "18")
	assert.NoError(t, err)
	assert.Contains(t, path, "18")
}

func TestGetNodePath_VersionNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	path, err := GetNodePath(tmpDir, "999")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}

func TestDetectVersionFromPackageJSON_PackageJSONWithoutEnginesField(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test-app","version":"1.0.0","dependencies":{"express":"^4.18.0"}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestDetectVersionFromPackageJSON_EnginesFieldEmptyString(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := `{"name":"test","engines":{"node":""}}`
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromPackageJSON(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "20", version)
}

func TestGetNodePath_GlobNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	nvmDir := filepath.Join(tmpDir, "nvm", "versions", "node")
	err := os.MkdirAll(nvmDir, 0755)
	require.NoError(t, err)

	path, err := GetNodePath(nvmDir, "20")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}
