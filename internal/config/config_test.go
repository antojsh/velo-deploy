package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, CaddyConf, cfg.CaddyConfDir)
	assert.Equal(t, AppsDir, cfg.AppsDir)
	assert.Equal(t, LogsDir, cfg.LogsDir)
	assert.Equal(t, "9999", cfg.DaemonPort)
	assert.Equal(t, "/opt/nvm", cfg.NVMDir)
	assert.NotNil(t, cfg.Apps)
	assert.Equal(t, 0, len(cfg.Apps))
}

func TestLoad_FileNotFound(t *testing.T) {
	originalConfigDir := ConfigDir
	ConfigDir = "/nonexistent/path/that/does/not/exist"
	defer func() { ConfigDir = originalConfigDir }()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, CaddyConf, cfg.CaddyConfDir)
	assert.NotNil(t, cfg.Apps)
}

func TestLoad_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	err := os.WriteFile(configFile, []byte("not valid json{"), 0644)
	require.NoError(t, err)

	originalConfigDir := ConfigDir
	ConfigDir = tmpDir
	defer func() { ConfigDir = originalConfigDir }()

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	testCfg := &Config{
		CaddyConfDir: "/custom/caddy",
		AppsDir:      "/custom/apps",
		LogsDir:      "/custom/logs",
		DaemonPort:   "8888",
		NVMDir:       "/custom/nvm",
		Apps: map[string]*AppMeta{
			"test-app": {
				Name:       "test-app",
				Type:       AppTypeNode,
				Port:       3000,
				Domain:     "test.example.com",
				Alias:      "test.local",
				NodeVer:    "20",
				EntryPoint: "index.js",
			},
		},
	}

	data, err := json.Marshal(testCfg)
	require.NoError(t, err)

	err = os.WriteFile(configFile, data, 0644)
	require.NoError(t, err)

	originalConfigDir := ConfigDir
	ConfigDir = tmpDir
	defer func() { ConfigDir = originalConfigDir }()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "/custom/caddy", cfg.CaddyConfDir)
	assert.Equal(t, "/custom/apps", cfg.AppsDir)
	assert.Equal(t, "/custom/logs", cfg.LogsDir)
	assert.Equal(t, "8888", cfg.DaemonPort)
	assert.Equal(t, "/custom/nvm", cfg.NVMDir)
	assert.Len(t, cfg.Apps, 1)

	app := cfg.Apps["test-app"]
	assert.NotNil(t, app)
	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, AppTypeNode, app.Type)
	assert.Equal(t, 3000, app.Port)
	assert.Equal(t, "test.example.com", app.Domain)
	assert.Equal(t, "test.local", app.Alias)
	assert.Equal(t, "20", app.NodeVer)
	assert.Equal(t, "index.js", app.EntryPoint)
}

func TestLoad_NullApps(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	jsonData := `{"apps": null}`
	err := os.WriteFile(configFile, []byte(jsonData), 0644)
	require.NoError(t, err)

	originalConfigDir := ConfigDir
	ConfigDir = tmpDir
	defer func() { ConfigDir = originalConfigDir }()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Apps)
	assert.Equal(t, 0, len(cfg.Apps))
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		CaddyConfDir: "/test/caddy",
		AppsDir:      "/test/apps",
		LogsDir:      "/test/logs",
		DaemonPort:   "7777",
		NVMDir:       "/test/nvm",
		Apps: map[string]*AppMeta{
			"myapp": {
				Name:       "myapp",
				Type:       AppTypeStatic,
				OutputDir:  "dist",
				Alias:      "myapp.local",
			},
		},
	}

	originalConfigDir := ConfigDir
	ConfigDir = tmpDir
	defer func() { ConfigDir = originalConfigDir }()

	err := cfg.Save()
	assert.NoError(t, err)

	loadedCfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "/test/caddy", loadedCfg.CaddyConfDir)
	assert.Equal(t, "7777", loadedCfg.DaemonPort)
	assert.Len(t, loadedCfg.Apps, 1)
	assert.Equal(t, AppTypeStatic, loadedCfg.Apps["myapp"].Type)
	assert.Equal(t, "dist", loadedCfg.Apps["myapp"].OutputDir)
}

func TestNextPort_NoApps(t *testing.T) {
	cfg := &Config{
		Apps: make(map[string]*AppMeta),
	}

	port := cfg.NextPort()
	assert.Equal(t, 3000, port)
}

func TestNextPort_SomeApps(t *testing.T) {
	cfg := &Config{
		Apps: map[string]*AppMeta{
			"app1": {Name: "app1", Type: AppTypeNode, Port: 3000},
			"app2": {Name: "app2", Type: AppTypeNode, Port: 3002},
		},
	}

	port := cfg.NextPort()
	assert.Equal(t, 3001, port)
}

func TestNextPort_AllPortsUsed(t *testing.T) {
	apps := make(map[string]*AppMeta)
	for i := 3000; i <= 3999; i++ {
		apps[string(rune(i))] = &AppMeta{Name: string(rune(i)), Type: AppTypeNode, Port: i}
	}

	cfg := &Config{Apps: apps}

	port := cfg.NextPort()
	assert.Equal(t, 0, port)
}

func TestNextPort_OnlyNodeApps(t *testing.T) {
	cfg := &Config{
		Apps: map[string]*AppMeta{
			"nodeapp1":  {Name: "nodeapp1", Type: AppTypeNode, Port: 3000},
			"staticapp": {Name: "staticapp", Type: AppTypeStatic, Port: 3001},
			"nodeapp2":  {Name: "nodeapp2", Type: AppTypeNode, Port: 3002},
		},
	}

	port := cfg.NextPort()
	assert.Equal(t, 3001, port)
}

func TestNextPort_OnlyStaticApps(t *testing.T) {
	cfg := &Config{
		Apps: map[string]*AppMeta{
			"staticapp1": {Name: "staticapp1", Type: AppTypeStatic, Port: 3000},
			"staticapp2": {Name: "staticapp2", Type: AppTypeStatic, Port: 3001},
		},
	}

	port := cfg.NextPort()
	assert.Equal(t, 3000, port)
}

func TestNextPort_GapInMiddle(t *testing.T) {
	cfg := &Config{
		Apps: map[string]*AppMeta{
			"app1": {Name: "app1", Type: AppTypeNode, Port: 3000},
			"app2": {Name: "app2", Type: AppTypeNode, Port: 3003},
			"app3": {Name: "app3", Type: AppTypeNode, Port: 3005},
		},
	}

	port := cfg.NextPort()
	assert.Equal(t, 3001, port)
}

func TestGetApp_Exists(t *testing.T) {
	cfg := &Config{
		Apps: map[string]*AppMeta{
			"myapp": {Name: "myapp", Type: AppTypeNode, Port: 3000},
		},
	}

	app := cfg.GetApp("myapp")
	assert.NotNil(t, app)
	assert.Equal(t, "myapp", app.Name)
}

func TestGetApp_NotExists(t *testing.T) {
	cfg := &Config{
		Apps: map[string]*AppMeta{
			"myapp": {Name: "myapp", Type: AppTypeNode, Port: 3000},
		},
	}

	app := cfg.GetApp("nonexistent")
	assert.Nil(t, app)
}

func TestGetApp_EmptyApps(t *testing.T) {
	cfg := &Config{
		Apps: make(map[string]*AppMeta),
	}

	app := cfg.GetApp("myapp")
	assert.Nil(t, app)
}

func TestAppMeta_Struct(t *testing.T) {
	app := &AppMeta{
		Name:       "test-app",
		Type:       AppTypeNode,
		RepoURL:    "https://github.com/user/repo",
		Branch:     "main",
		NodeVer:    "20",
		Port:       3000,
		Domain:     "api.example.com",
		Alias:      "test.local",
		NodePath:   "/opt/nvm/versions/node/v20.0.0/bin/node",
		EntryPoint: "server.js",
		OutputDir:  "",
	}

	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, AppTypeNode, app.Type)
	assert.Equal(t, "https://github.com/user/repo", app.RepoURL)
	assert.Equal(t, "main", app.Branch)
	assert.Equal(t, "20", app.NodeVer)
	assert.Equal(t, 3000, app.Port)
	assert.Equal(t, "api.example.com", app.Domain)
	assert.Equal(t, "test.local", app.Alias)
	assert.Equal(t, "/opt/nvm/versions/node/v20.0.0/bin/node", app.NodePath)
	assert.Equal(t, "server.js", app.EntryPoint)
	assert.Equal(t, "", app.OutputDir)
}

func TestAppMeta_StaticType(t *testing.T) {
	app := &AppMeta{
		Name:      "static-site",
		Type:      AppTypeStatic,
		Alias:     "site.local",
		OutputDir: "dist",
	}

	assert.Equal(t, "static-site", app.Name)
	assert.Equal(t, AppTypeStatic, app.Type)
	assert.Equal(t, "dist", app.OutputDir)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "/etc/deploy", ConfigDir)
	assert.Equal(t, "config.json", ConfigFile)
	assert.Equal(t, "/opt/deploy/apps", AppsDir)
	assert.Equal(t, "/var/log/deploy", LogsDir)
	assert.Equal(t, "/etc/caddy/conf.d", CaddyConf)
	assert.Equal(t, "node", AppTypeNode)
	assert.Equal(t, "static", AppTypeStatic)
}
