package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ConfigDir  = "/etc/deploy"
	ConfigFile = "config.json"
	AppsDir    = "/opt/deploy/apps"
	LogsDir    = "/var/log/deploy"
	CaddyConf  = "/etc/caddy/conf.d"
)

type Config struct {
	CaddyConfDir string            `json:"caddy_conf_dir"`
	AppsDir      string            `json:"apps_dir"`
	LogsDir      string            `json:"logs_dir"`
	DaemonPort   string            `json:"daemon_port"`
	NVMDir       string            `json:"nvm_dir"`
	Apps         map[string]*AppMeta `json:"apps"`
}

type AppMeta struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // "node" | "static"
	RepoURL    string `json:"repo_url"`
	Branch     string `json:"branch"`
	NodeVer    string `json:"node_version"`
	Port       int    `json:"port"`
	Domain     string `json:"domain"`
	Alias      string `json:"alias"`
	NodePath   string `json:"node_path"`
	EntryPoint string `json:"entry_point"`   // e.g. "index.js" or "server.js"
	OutputDir  string `json:"output_dir"`   // e.g. "dist", "build", "_site" (for static)
}

const (
	AppTypeNode   = "node"
	AppTypeStatic = "static"
)

func DefaultConfig() *Config {
	return &Config{
		CaddyConfDir: CaddyConf,
		AppsDir:      AppsDir,
		LogsDir:      LogsDir,
		DaemonPort:   "9999",
		NVMDir:       "/opt/nvm",
		Apps:         make(map[string]*AppMeta),
	}
}

func Load() (*Config, error) {
	cfgPath := filepath.Join(ConfigDir, ConfigFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		// First run, return defaults
		return DefaultConfig(), nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Apps == nil {
		cfg.Apps = make(map[string]*AppMeta)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	os.MkdirAll(ConfigDir, 0755)
	cfgPath := filepath.Join(ConfigDir, ConfigFile)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0644)
}

func (c *Config) NextPort() int {
	minPort := 3000
	maxPort := 3999
	used := make(map[int]bool)
	for _, app := range c.Apps {
		if app.Type == AppTypeNode {
			used[app.Port] = true
		}
	}
	for p := minPort; p <= maxPort; p++ {
		if !used[p] {
			return p
		}
	}
	return 0 // no free port
}

func (c *Config) GetApp(appName string) *AppMeta {
	return c.Apps[appName]
}
