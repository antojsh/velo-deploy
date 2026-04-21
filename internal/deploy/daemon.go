package deploy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"deploy/internal/caddy"
	"deploy/internal/config"
	deploynode "deploy/internal/node"
	"deploy/internal/systemd"
)

type webhookPayload struct {
	Ref  string `json:"ref"`
	Repo struct {
		CloneURL string `json:"clone_url"`
		Name     string `json:"name"`
	} `json:"repository"`
}

// RunDaemon starts the HTTP listener for GitHub webhooks.
func RunDaemon(cfg *config.Config, port string) error {
	fmt.Printf("Starting deploy daemon on :%s\n", port)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var payload webhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Only process pushes to main branch
		if payload.Ref != "refs/heads/main" && payload.Ref != "refs/heads/master" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ignored branch"))
			return
		}

		// Try to match by repo name
		appName := deriveAppName(payload.Repo.CloneURL)
		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Webhook received for unknown app: %s\n", appName)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("app not found"))
			return
		}

		// Run the deploy in a goroutine
		go func() {
			fmt.Printf("Auto-deploy triggered for %s\n", appName)
			if err := autoDeploy(cfg, app); err != nil {
				fmt.Fprintf(os.Stderr, "Auto-deploy failed for %s: %v\n", appName, err)
			} else {
				fmt.Printf("Auto-deploy successful for %s\n", appName)
			}
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deploying..."))
	})

	return http.ListenAndServe(":"+port, nil)
}

func autoDeploy(cfg *config.Config, app *config.AppMeta) error {
	appDir := filepath.Join(cfg.AppsDir, app.Name)

	// 1. Git pull
	fmt.Printf("[%s] Pulling latest code...\n", app.Name)
	cmd := exec.Command("git", "pull")
	cmd.Dir = appDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	// 2. Check Node version (may have changed)
	nodeVer, err := deploynode.DetectVersionFromPackageJSON(appDir)
	if err != nil {
		nodeVer = app.NodeVer // Use existing
	}

	// 3. Ensure node is installed
	if err := deploynode.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
		return fmt.Errorf("node install failed: %w", err)
	}

	nodePath, err := deploynode.GetNodePath(cfg.NVMDir, nodeVer)
	if err != nil {
		return fmt.Errorf("node path failed: %w", err)
	}

	// 4. npm install
	fmt.Printf("[%s] Installing deps...\n", app.Name)
	if err := deploynode.InstallDeps(appDir, nodePath); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	// 5. Set ownership
	username := "deploy-" + app.Name
	exec.Command("chown", "-R", username+":"+username, appDir).Run()

	// 6. Update config if node version changed
	app.NodeVer = nodeVer
	app.NodePath = nodePath
	cfg.Save()

	// 7. Regenerate service file (in case node path changed)
	systemd.GenerateService(app.Name, nodePath, appDir, app.EntryPoint, username, username)
	systemd.DaemonReload()

	// 8. Restart the app
	if err := systemd.RestartApp(app.Name); err != nil {
		return fmt.Errorf("restart failed: %w", err)
	}

	// 9. Rebuild shared Caddy config and reload
	rebuildSharedCaddyConfig(cfg)
	caddy.Reload()

	return nil
}
