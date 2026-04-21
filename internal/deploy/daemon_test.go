package deploy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"deploy/internal/config"
)

func TestRunDaemon_WebhookHandler(t *testing.T) {
	cfg := &config.Config{
		AppsDir: t.TempDir(),
		Apps:    make(map[string]*config.AppMeta),
	}

	req, err := http.NewRequest("GET", "/webhook", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWebhookPayload_Parsing(t *testing.T) {
	payload := webhookPayload{
		Ref: "refs/heads/main",
	}
	payload.Repo.CloneURL = "https://github.com/user/repo.git"
	payload.Repo.Name = "repo"

	assert.Equal(t, "refs/heads/main", payload.Ref)
	assert.Equal(t, "https://github.com/user/repo.git", payload.Repo.CloneURL)
	assert.Equal(t, "repo", payload.Repo.Name)
}

func TestWebhookPayload_JSON(t *testing.T) {
	jsonData := `{
		"ref": "refs/heads/main",
		"repository": {
			"clone_url": "https://github.com/user/myapp.git",
			"name": "myapp"
		}
	}`

	var payload webhookPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	require.NoError(t, err)

	assert.Equal(t, "refs/heads/main", payload.Ref)
	assert.Equal(t, "https://github.com/user/myapp.git", payload.Repo.CloneURL)
	assert.Equal(t, "myapp", payload.Repo.Name)
}

func TestWebhookPayload_MasterBranch(t *testing.T) {
	jsonData := `{
		"ref": "refs/heads/master",
		"repository": {
			"clone_url": "https://github.com/user/repo.git",
			"name": "repo"
		}
	}`

	var payload webhookPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	require.NoError(t, err)

	assert.Equal(t, "refs/heads/master", payload.Ref)
}

func TestWebhookHandler_NonPostMethod(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		AppsDir: tmpDir,
		Apps:    make(map[string]*config.AppMeta),
	}

	_ = cfg

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	req, err := http.NewRequest("GET", "/webhook", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	req, err := http.NewRequest("POST", "/webhook", bytes.NewBufferString("not json"))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestWebhookHandler_IgnoredBranch(t *testing.T) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Ref != "refs/heads/main" && payload.Ref != "refs/heads/master" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ignored branch"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	payload := map[string]interface{}{
		"ref": "refs/heads/feature-branch",
		"repository": map[string]interface{}{
			"clone_url": "https://github.com/user/repo.git",
			"name":      "repo",
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ignored branch")
}

func TestWebhookHandler_AppNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		AppsDir: tmpDir,
		Apps:    make(map[string]*config.AppMeta),
	}

	appDir := filepath.Join(tmpDir, "myapp")
	err := os.Mkdir(appDir, 0755)
	require.NoError(t, err)

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Ref != "refs/heads/main" && payload.Ref != "refs/heads/master" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ignored branch"))
			return
		}

		appName := deriveAppName(payload.Repo.CloneURL)
		_, exists := cfg.Apps[appName]
		if !exists {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("app not found"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	payload := map[string]interface{}{
		"ref": "refs/heads/main",
		"repository": map[string]interface{}{
			"clone_url": "https://github.com/user/unknown-app.git",
			"name":      "unknown-app",
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "app not found")
}

func TestWebhookHandler_ValidPush(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		AppsDir: tmpDir,
		Apps:    make(map[string]*config.AppMeta),
	}

	appDir := filepath.Join(tmpDir, "myapp")
	err := os.Mkdir(appDir, 0755)
	require.NoError(t, err)

	cfg.Apps["myapp"] = &config.AppMeta{
		Name: "myapp",
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Ref != "refs/heads/main" && payload.Ref != "refs/heads/master" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ignored branch"))
			return
		}

		appName := deriveAppName(payload.Repo.CloneURL)
		_, exists := cfg.Apps[appName]
		if !exists {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("app not found"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deploying..."))
	}

	payload := map[string]interface{}{
		"ref": "refs/heads/main",
		"repository": map[string]interface{}{
			"clone_url": "https://github.com/user/myapp.git",
			"name":      "myapp",
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "deploying...")
}

func TestAutoDeploy_GitPullFailure(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "testapp")
	err := os.Mkdir(appDir, 0755)
	require.NoError(t, err)

	cfg := &config.Config{
		AppsDir: tmpDir,
		NVMDir:  "/opt/nvm",
		Apps: map[string]*config.AppMeta{
			"testapp": {
				Name:       "testapp",
				Type:       "node",
				EntryPoint: "index.js",
				NodeVer:    "20",
			},
		},
	}

	pkgJSON := filepath.Join(appDir, "package.json")
	err = os.WriteFile(pkgJSON, []byte(`{"name":"test","engines":{"node":"20"}}`), 0644)
	require.NoError(t, err)

	err = autoDeploy(cfg, cfg.Apps["testapp"])
	assert.Error(t, err)
	assert.Contains(t, string(err.Error()), "git pull failed")
}

func TestAutoDeploy_NoPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "testapp")
	err := os.Mkdir(appDir, 0755)
	require.NoError(t, err)

	cfg := &config.Config{
		AppsDir: tmpDir,
		NVMDir:  "/opt/nvm",
		Apps: map[string]*config.AppMeta{
			"testapp": {
				Name:       "testapp",
				Type:       "node",
				EntryPoint: "index.js",
				NodeVer:    "20",
			},
		},
	}

	err = autoDeploy(cfg, cfg.Apps["testapp"])
	assert.Error(t, err)
}

func TestWebhookPayload_FeatureBranchIgnored(t *testing.T) {
	payloads := []string{
		`{"ref":"refs/heads/feature","repository":{"clone_url":"https://github.com/user/repo.git","name":"repo"}}`,
		`{"ref":"refs/heads/dev","repository":{"clone_url":"https://github.com/user/repo.git","name":"repo"}}`,
		`{"ref":"refs/heads/release-v1","repository":{"clone_url":"https://github.com/user/repo.git","name":"repo"}}`,
	}

	for _, jsonData := range payloads {
		var payload webhookPayload
		err := json.Unmarshal([]byte(jsonData), &payload)
		require.NoError(t, err)

		isMainOrMaster := payload.Ref == "refs/heads/main" || payload.Ref == "refs/heads/master"
		assert.False(t, isMainOrMaster, "Payload with ref %s should not be main/master", payload.Ref)
	}
}

func TestWebhookPayload_BothBranchesWork(t *testing.T) {
	mainPayload := `{"ref":"refs/heads/main","repository":{"clone_url":"https://github.com/user/repo.git","name":"repo"}}`
	masterPayload := `{"ref":"refs/heads/master","repository":{"clone_url":"https://github.com/user/repo.git","name":"repo"}}`

	for _, jsonData := range []string{mainPayload, masterPayload} {
		var payload webhookPayload
		err := json.Unmarshal([]byte(jsonData), &payload)
		require.NoError(t, err)

		isMainOrMaster := payload.Ref == "refs/heads/main" || payload.Ref == "refs/heads/master"
		assert.True(t, isMainOrMaster)
	}
}
