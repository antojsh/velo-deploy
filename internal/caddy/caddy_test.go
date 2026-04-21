package caddy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateConfig_WithDomain(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "test-app.conf")
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := GenerateConfig("test-app", "example.com", "localhost:3000")

	assert.NoError(t, err)
	content, err := os.ReadFile(confPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "example.com")
	assert.Contains(t, string(content), "reverse_proxy localhost:3000")
}

func TestGenerateConfig_WithoutDomain(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "test-app.conf")
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := GenerateConfig("test-app", "", "localhost:3000")

	assert.NoError(t, err)
	_, err = os.Stat(confPath)
	assert.True(t, os.IsNotExist(err), "should not create config file when domain is empty")
}

func TestGenerateStaticConfig_WithDomain(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "static-app.conf")
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := GenerateStaticConfig("static-app", "static.example.com", "/var/www/static")

	assert.NoError(t, err)
	content, err := os.ReadFile(confPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "static.example.com")
	assert.Contains(t, string(content), "root * /var/www/static")
	assert.Contains(t, string(content), "file_server")
}

func TestGenerateStaticConfig_WithoutDomain(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "static-app.conf")
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := GenerateStaticConfig("static-app", "", "/var/www/static")

	assert.NoError(t, err)
	_, err = os.Stat(confPath)
	assert.True(t, os.IsNotExist(err), "should not create config file when domain is empty")
}

func TestRebuildSharedConfig_WithRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	routes := []Route{
		{AppName: "app1", Upstream: "app1.local:3000", Type: "node"},
		{AppName: "app2", Upstream: "app2.local:3001", Type: "node"},
	}

	err := RebuildSharedConfig(routes)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), ":443")
	assert.Contains(t, string(content), "tls internal")
	assert.Contains(t, string(content), "/app1/*")
	assert.Contains(t, string(content), "reverse_proxy app1.local:3000")
	assert.Contains(t, string(content), "/app2/*")
	assert.Contains(t, string(content), "reverse_proxy app2.local:3001")
	assert.Contains(t, string(content), "respond 404")
}

func TestRebuildSharedConfig_WithStaticRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	routes := []Route{
		{AppName: "static1", Type: "static", RootDir: "/opt/deploy/apps/static1/dist"},
	}

	err := RebuildSharedConfig(routes)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "/static1/*")
	assert.Contains(t, string(content), "root * /opt/deploy/apps/static1/dist")
	assert.Contains(t, string(content), "file_server")
	assert.NotContains(t, string(content), "reverse_proxy")
}

func TestRebuildSharedConfig_MixedRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	routes := []Route{
		{AppName: "nodeapp", Upstream: "nodeapp.local:3000", Type: "node"},
		{AppName: "staticapp", Type: "static", RootDir: "/opt/deploy/static"},
	}

	err := RebuildSharedConfig(routes)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "/nodeapp/*")
	assert.Contains(t, string(content), "reverse_proxy nodeapp.local:3000")
	assert.Contains(t, string(content), "/staticapp/*")
	assert.Contains(t, string(content), "root * /opt/deploy/static")
	assert.Contains(t, string(content), "file_server")
}

func TestRebuildSharedConfig_EmptyRoutes(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := RebuildSharedConfig([]Route{})
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	_, err = os.Stat(sharedPath)
	assert.True(t, os.IsNotExist(err), "should remove shared config when no routes")
}

func TestRemoveSharedConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	err := os.WriteFile(sharedPath, []byte("test"), 0644)
	require.NoError(t, err)

	err = RemoveSharedConfig()
	assert.NoError(t, err)

	_, err = os.Stat(sharedPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveSharedConfig_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := RemoveSharedConfig()
	assert.NoError(t, err)
}

func TestRemoveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	confPath := filepath.Join(tmpDir, "test-app.conf")
	err := os.WriteFile(confPath, []byte("test"), 0644)
	require.NoError(t, err)

	err = RemoveConfig("test-app")
	assert.NoError(t, err)

	_, err = os.Stat(confPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveConfig_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	err := RemoveConfig("nonexistent-app")
	assert.NoError(t, err)
}

func TestReload(t *testing.T) {
	err := Reload()
	assert.NoError(t, err)
}

func TestRoute_Struct(t *testing.T) {
	route := Route{
		AppName:  "myapp",
		Upstream: "myapp.local:3000",
		Domain:   "myapp.com",
		Type:     "node",
		RootDir:  "/opt/deploy/myapp",
	}

	assert.Equal(t, "myapp", route.AppName)
	assert.Equal(t, "myapp.local:3000", route.Upstream)
	assert.Equal(t, "myapp.com", route.Domain)
	assert.Equal(t, "node", route.Type)
	assert.Equal(t, "/opt/deploy/myapp", route.RootDir)
}

func TestRebuildSharedConfig_HandlePathCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	routes := []Route{
		{AppName: "my-app", Upstream: "my-app.local:3000", Type: "node"},
	}

	err := RebuildSharedConfig(routes)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "handle_path /my-app/*")
}

func TestRebuildSharedConfig_PreservesNewlines(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	confDir = tmpDir
	defer func() { confDir = originalConfDir }()

	routes := []Route{
		{AppName: "app1", Upstream: "app1.local:3000", Type: "node"},
		{AppName: "app2", Upstream: "app2.local:3001", Type: "node"},
	}

	err := RebuildSharedConfig(routes)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	require.NoError(t, err)

	app1Idx := strings.Index(content, "/app1/")
	app2Idx := strings.Index(content, "/app2/")
	assert.True(t, app1Idx < app2Idx, "app1 should come before app2")
}

func TestSetConfDir(t *testing.T) {
	tmpDir := t.TempDir()
	originalConfDir := confDir
	SetConfDir(tmpDir)
	defer func() { confDir = originalConfDir }()

	assert.Equal(t, tmpDir, confDir)

	confPath := filepath.Join(tmpDir, "test.conf")
	err := os.WriteFile(confPath, []byte("test"), 0644)
	require.NoError(t, err)

	RemoveConfig("test")
	_, err = os.Stat(confPath)
	assert.True(t, os.IsNotExist(err))
}
