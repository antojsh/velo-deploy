package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"velo-deploy/internal/caddy"
	"velo-deploy/internal/config"
	"velo-deploy/internal/hosts"
)

func TestDeriveAppName(t *testing.T) {
	tests := []struct {
		repoURL string
		expected string
	}{
		{"https://github.com/user/my-repo.git", "my-repo"},
		{"https://github.com/user/my-repo", "my-repo"},
		{"https://github.com/user/my-repo/", "my-repo"},
		{"https://github.com/user/My-Repo.git", "my-repo"},
		{"https://github.com/user/my_repo.git", "my-repo"},
		{"https://github.com/user/my.repo.git", "my-repo"},
		{"https://github.com/user/AbC123.git", "abc123"},
		{"https://github.com/user/.hidden.git", "hidden"},
		{"my-random-string", "my-random-string"},
		{"https://github.com/org/repo-name-with-dashes.git", "repo-name-with-dashes"},
	}

	for _, tt := range tests {
		t.Run(tt.repoURL, func(t *testing.T) {
			result := deriveAppName(tt.repoURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectEntryPoint(t *testing.T) {
	tmpDir := t.TempDir()

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "index.js", result)

	indexJS := filepath.Join(tmpDir, "index.js")
	err := os.WriteFile(indexJS, []byte("console.log('hello')"), 0644)
	require.NoError(t, err)

	result = detectEntryPoint(tmpDir)
	assert.Equal(t, "index.js", result)
}

func TestDetectEntryPoint_ServerJS(t *testing.T) {
	tmpDir := t.TempDir()

	serverJS := filepath.Join(tmpDir, "server.js")
	err := os.WriteFile(serverJS, []byte("console.log('server')"), 0644)
	require.NoError(t, err)

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "server.js", result)
}

func TestDetectEntryPoint_AppJS(t *testing.T) {
	tmpDir := t.TempDir()

	appJS := filepath.Join(tmpDir, "app.js")
	err := os.WriteFile(appJS, []byte("console.log('app')"), 0644)
	require.NoError(t, err)

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "app.js", result)
}

func TestDetectEntryPoint_MainJS(t *testing.T) {
	tmpDir := t.TempDir()

	mainJS := filepath.Join(tmpDir, "main.js")
	err := os.WriteFile(mainJS, []byte("console.log('main')"), 0644)
	require.NoError(t, err)

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "main.js", result)
}

func TestDetectEntryPoint_SrcIndex(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)

	srcIndexJS := filepath.Join(srcDir, "index.js")
	err = os.WriteFile(srcIndexJS, []byte("console.log('src')"), 0644)
	require.NoError(t, err)

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "src/index.js", result)
}

func TestDetectEntryPoint_DistIndex(t *testing.T) {
	tmpDir := t.TempDir()
	distDir := filepath.Join(tmpDir, "dist")
	err := os.Mkdir(distDir, 0755)
	require.NoError(t, err)

	distIndexJS := filepath.Join(distDir, "index.js")
	err = os.WriteFile(distIndexJS, []byte("console.log('dist')"), 0644)
	require.NoError(t, err)

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "dist/index.js", result)
}

func TestDetectEntryPoint_Priority(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{"index.js", "server.js", "app.js", "main.js"}
	for _, f := range files {
		p := filepath.Join(tmpDir, f)
		err := os.WriteFile(p, []byte("console.log('test')"), 0644)
		require.NoError(t, err)
	}

	result := detectEntryPoint(tmpDir)
	assert.Equal(t, "index.js", result)
}

func TestDetectAppType_StaticWithIndexHTML(t *testing.T) {
	tmpDir := t.TempDir()
	indexHTML := filepath.Join(tmpDir, "index.html")
	err := os.WriteFile(indexHTML, []byte("<html></html>"), 0644)
	require.NoError(t, err)

	result := DetectAppType(tmpDir)
	assert.Equal(t, config.AppTypeStatic, result)
}

func TestDetectAppType_StaticWithPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := filepath.Join(tmpDir, "package.json")
	err := os.WriteFile(pkgJSON, []byte(`{"name":"test"}`), 0644)
	require.NoError(t, err)

	result := DetectAppType(tmpDir)
	assert.Equal(t, config.AppTypeStatic, result)
}

func TestDetectAppType_Node(t *testing.T) {
	tmpDir := t.TempDir()

	result := DetectAppType(tmpDir)
	assert.Equal(t, config.AppTypeNode, result)
}

func TestDetectAppType_IndexHTMLTakesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	indexHTML := filepath.Join(tmpDir, "index.html")
	err := os.WriteFile(indexHTML, []byte("<html></html>"), 0644)
	require.NoError(t, err)

	result := DetectAppType(tmpDir)
	assert.Equal(t, config.AppTypeStatic, result)
}

func TestDetectOutputDir_Dist(t *testing.T) {
	tmpDir := t.TempDir()
	distDir := filepath.Join(tmpDir, "dist")
	err := os.Mkdir(distDir, 0755)
	require.NoError(t, err)

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "dist", result)
}

func TestDetectOutputDir_Build(t *testing.T) {
	tmpDir := t.TempDir()
	buildDir := filepath.Join(tmpDir, "build")
	err := os.Mkdir(buildDir, 0755)
	require.NoError(t, err)

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "build", result)
}

func TestDetectOutputDir_Site(t *testing.T) {
	tmpDir := t.TempDir()
	siteDir := filepath.Join(tmpDir, "_site")
	err := os.Mkdir(siteDir, 0755)
	require.NoError(t, err)

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "_site", result)
}

func TestDetectOutputDir_Public(t *testing.T) {
	tmpDir := t.TempDir()
	publicDir := filepath.Join(tmpDir, "public")
	err := os.Mkdir(publicDir, 0755)
	require.NoError(t, err)

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "public", result)
}

func TestDetectOutputDir_Output(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	err := os.Mkdir(outputDir, 0755)
	require.NoError(t, err)

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "output", result)
}

func TestDetectOutputDir_Priority(t *testing.T) {
	tmpDir := t.TempDir()
	dirs := []string{"dist", "build", "_site", "public", "output"}
	for _, d := range dirs {
		dirPath := filepath.Join(tmpDir, d)
		err := os.Mkdir(dirPath, 0755)
		require.NoError(t, err)
	}

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "dist", result)
}

func TestDetectOutputDir_None(t *testing.T) {
	tmpDir := t.TempDir()

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "", result)
}

func TestDetectOutputDir_FileNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	distFile := filepath.Join(tmpDir, "dist")
	err := os.WriteFile(distFile, []byte("not a directory"), 0644)
	require.NoError(t, err)

	result := DetectOutputDir(tmpDir)
	assert.Equal(t, "", result)
}

func TestRebuildSharedCaddyConfig_Empty(t *testing.T) {
	cfg := &config.Config{
		Apps: make(map[string]*config.AppMeta),
	}

	err := rebuildSharedCaddyConfig(cfg)
	assert.NoError(t, err)
}

func TestRebuildSharedCaddyConfig_WithNodeApp(t *testing.T) {
	tmpDir := t.TempDir()
	
	caddy.SetConfDir(tmpDir)
	defer caddy.SetConfDir("/etc/caddy/conf.d")
	
	appsDir := filepath.Join(tmpDir, "apps")
	err := os.Mkdir(appsDir, 0755)
	require.NoError(t, err)
	
	appDir := filepath.Join(appsDir, "myapp")
	err = os.Mkdir(appDir, 0755)
	require.NoError(t, err)
	
	cfg := &config.Config{
		AppsDir: appsDir,
		Apps: map[string]*config.AppMeta{
			"myapp": {
				Name:    "myapp",
				Type:    config.AppTypeNode,
				Port:    3000,
				Alias:   "myapp.local",
				OutputDir: "",
			},
		},
	}

	err = rebuildSharedCaddyConfig(cfg)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "/myapp/*")
	assert.Contains(t, string(content), "reverse_proxy myapp.local:3000")
}

func TestRebuildSharedCaddyConfig_WithStaticApp(t *testing.T) {
	tmpDir := t.TempDir()
	
	caddy.SetConfDir(tmpDir)
	defer caddy.SetConfDir("/etc/caddy/conf.d")
	
	appsDir := filepath.Join(tmpDir, "apps")
	err := os.Mkdir(appsDir, 0755)
	require.NoError(t, err)
	
	appDir := filepath.Join(appsDir, "mysite")
	err = os.Mkdir(appDir, 0755)
	require.NoError(t, err)
	
	distDir := filepath.Join(appDir, "dist")
	err = os.Mkdir(distDir, 0755)
	require.NoError(t, err)
	
	cfg := &config.Config{
		AppsDir: appsDir,
		Apps: map[string]*config.AppMeta{
			"mysite": {
				Name:      "mysite",
				Type:      config.AppTypeStatic,
				Alias:     "mysite.local",
				OutputDir: "dist",
			},
		},
	}

	err = rebuildSharedCaddyConfig(cfg)
	assert.NoError(t, err)

	sharedPath := filepath.Join(tmpDir, "_shared.conf")
	content, err := os.ReadFile(sharedPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "/mysite/*")
	assert.Contains(t, string(content), "root *")
	assert.Contains(t, string(content), "file_server")
}

func TestRegister_StaticWithDomainPersistsMetadataAndCaddyConfig(t *testing.T) {
    tmpDir := t.TempDir()
    appsDir := filepath.Join(tmpDir, "apps")
    appDir := filepath.Join(appsDir, "site")
    distDir := filepath.Join(appDir, "dist")
    require.NoError(t, os.MkdirAll(distDir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<h1>ok</h1>"), 0644))

    confDir := filepath.Join(tmpDir, "caddy")
    hostsPath := filepath.Join(tmpDir, "hosts")
    require.NoError(t, os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644))
	hosts.SetHostsFile(hostsPath)
	defer hosts.SetHostsFile("/etc/hosts")

    caddy.SetConfDir(confDir)
    defer caddy.SetConfDir("/etc/caddy/conf.d")

    originalConfigDir := config.ConfigDir
    config.ConfigDir = filepath.Join(tmpDir, "config")
    defer func() { config.ConfigDir = originalConfigDir }()

    cfg := &config.Config{AppsDir: appsDir, Apps: map[string]*config.AppMeta{}}

    err := Register(cfg, "site", appDir, "site.example.com", "site.local", config.AppTypeStatic)

    assert.NoError(t, err)
    require.Contains(t, cfg.Apps, "site")
    assert.Equal(t, config.AppTypeStatic, cfg.Apps["site"].Type)
    assert.Equal(t, "dist", cfg.Apps["site"].OutputDir)
    assert.Equal(t, "site.example.com", cfg.Apps["site"].Domain)

    caddyConfig, err := os.ReadFile(filepath.Join(confDir, "site.conf"))
    require.NoError(t, err)
    assert.Contains(t, string(caddyConfig), "site.example.com")
    assert.Contains(t, string(caddyConfig), "file_server")
}

func TestRegister_StaticWithoutDomainRebuildsSharedConfig(t *testing.T) {
    tmpDir := t.TempDir()
    appsDir := filepath.Join(tmpDir, "apps")
    appDir := filepath.Join(appsDir, "docs")
    require.NoError(t, os.MkdirAll(appDir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(appDir, "index.html"), []byte("<h1>docs</h1>"), 0644))

    confDir := filepath.Join(tmpDir, "caddy")
    hostsPath := filepath.Join(tmpDir, "hosts")
    require.NoError(t, os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644))
	hosts.SetHostsFile(hostsPath)
	defer hosts.SetHostsFile("/etc/hosts")

    caddy.SetConfDir(confDir)
    defer caddy.SetConfDir("/etc/caddy/conf.d")

    originalConfigDir := config.ConfigDir
    config.ConfigDir = filepath.Join(tmpDir, "config")
    defer func() { config.ConfigDir = originalConfigDir }()

    cfg := &config.Config{AppsDir: appsDir, Apps: map[string]*config.AppMeta{}}

    err := Register(cfg, "docs", appDir, "", "", config.AppTypeStatic)

    assert.NoError(t, err)
    assert.Equal(t, "docs.local", cfg.Apps["docs"].Alias)
    assert.Equal(t, ".", cfg.Apps["docs"].OutputDir)

    sharedConfig, err := os.ReadFile(filepath.Join(confDir, "_shared.conf"))
    require.NoError(t, err)
    assert.Contains(t, string(sharedConfig), "handle_path /docs/*")
    assert.Contains(t, string(sharedConfig), "file_server")
}

func TestRegisterRejectsInvalidInputs(t *testing.T) {
    tmpDir := t.TempDir()
    cfg := &config.Config{AppsDir: tmpDir, Apps: map[string]*config.AppMeta{
        "existing": {Name: "existing"},
    }}

    err := Register(cfg, "existing", tmpDir, "", "", config.AppTypeStatic)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already exists")

    err = Register(cfg, "missing", filepath.Join(tmpDir, "missing"), "", "", config.AppTypeStatic)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "does not exist")

    filePath := filepath.Join(tmpDir, "file.txt")
    require.NoError(t, os.WriteFile(filePath, []byte("x"), 0644))
    err = Register(cfg, "file", filePath, "", "", config.AppTypeStatic)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "is not a directory")
}

func TestRemoveStaticAppDeletesManagedFilesAndMetadata(t *testing.T) {
    tmpDir := t.TempDir()
    appsDir := filepath.Join(tmpDir, "apps")
    appDir := filepath.Join(appsDir, "site")
    require.NoError(t, os.MkdirAll(appDir, 0755))

    confDir := filepath.Join(tmpDir, "caddy")
    caddy.SetConfDir(confDir)
    defer caddy.SetConfDir("/etc/caddy/conf.d")
    require.NoError(t, os.MkdirAll(confDir, 0755))
    require.NoError(t, os.WriteFile(filepath.Join(confDir, "site.conf"), []byte("site config"), 0644))

    hostsPath := filepath.Join(tmpDir, "hosts")
    require.NoError(t, os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n127.0.0.1  site.local\n"), 0644))
	hosts.SetHostsFile(hostsPath)
	defer hosts.SetHostsFile("/etc/hosts")

    originalConfigDir := config.ConfigDir
    config.ConfigDir = filepath.Join(tmpDir, "config")
    defer func() { config.ConfigDir = originalConfigDir }()

    cfg := &config.Config{AppsDir: appsDir, Apps: map[string]*config.AppMeta{
        "site": {Name: "site", Type: config.AppTypeStatic, Alias: "site.local"},
    }}

    err := Remove(cfg, "site")

    assert.NoError(t, err)
    assert.NotContains(t, cfg.Apps, "site")
    _, err = os.Stat(appDir)
    assert.True(t, os.IsNotExist(err))
}

func TestRemoveRejectsUnknownApp(t *testing.T) {
    cfg := &config.Config{Apps: map[string]*config.AppMeta{}}

    err := Remove(cfg, "missing")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}



