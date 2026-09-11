package deploy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"velo-deploy/internal/config"
	deploynode "velo-deploy/internal/node"
	"velo-deploy/internal/systemd"
)

type webhookPayload struct {
	Ref  string `json:"ref"`
	Repo struct {
		CloneURL string `json:"clone_url"`
		Name     string `json:"name"`
	} `json:"repository"`
}

func RunDaemon(cfg *config.Config, port string) error {
	fmt.Printf("Starting deploy daemon on :%s\n", port)
	http.HandleFunc("/webhook", webhookHandler(cfg))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	return http.ListenAndServe(":"+port, nil)
}

func webhookHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if payload.Ref != "refs/heads/main" && payload.Ref != "refs/heads/master" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ignored branch"))
			return
		}

		appName := deriveAppName(payload.Repo.CloneURL)
		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Webhook received for unknown app: %s\n", appName)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("app not found"))
			return
		}

		go func() {
			fmt.Printf("Auto-deploy triggered for %s\n", appName)
			if err := autoDeploy(cfg, app); err != nil {
				fmt.Fprintf(os.Stderr, "Auto-deploy failed for %s: %v\n", appName, err)
			} else {
				fmt.Printf("Auto-deploy successful for %s\n", appName)
			}
		}()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("deploying..."))
	}
}

func autoDeploy(cfg *config.Config, app *config.AppMeta) error {
	appDir := filepath.Join(cfg.AppsDir, app.Name)

	fmt.Printf("[%s] Pulling latest code...\n", app.Name)
	cmd := exec.Command("git", "pull")
	cmd.Dir = appDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	nodeVer, err := deploynode.DetectVersionFromPackageJSON(appDir)
	if err != nil {
		nodeVer = app.NodeVer
	}
	if nodeVer == "" {
		nodeVer = deploynode.DefaultNodeVersion
	}
	if err := deploynode.EnsureInstalled(cfg.NVMDir, nodeVer); err != nil {
		return fmt.Errorf("node install failed: %w", err)
	}
	nodePath, err := deploynode.GetNodePath(cfg.NVMDir, nodeVer)
	if err != nil {
		return fmt.Errorf("node path failed: %w", err)
	}
	if err := deploynode.InstallDeps(appDir, nodePath); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}
	buildCommand := app.BuildCommand
	if buildCommand == "" {
		buildCommand = DefaultBuildCommand
	}
	if err := deploynode.RunCommand(appDir, nodePath, buildCommand); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	username := systemd.AppUser(app.Name)
	_ = exec.Command("chown", "-R", username+":"+username, appDir).Run()
	app.NodeVer = nodeVer
	app.NodePath = nodePath
	_ = cfg.Save()

	startCommand := app.StartCommand
	if startCommand == "" {
		startCommand = DefaultStartCommand
	}
	_ = systemd.GenerateCommandService(app.Name, nodePath, appDir, startCommand, username, username)
	_ = systemd.DaemonReload()
	if err := systemd.RestartApp(app.Name); err != nil {
		return fmt.Errorf("restart failed: %w", err)
	}
	_ = rebuildSharedCaddyConfig(cfg)
	_ = reloadCaddy()
	return nil
}
