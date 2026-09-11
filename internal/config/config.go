package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var (
	ConfigDir       = "/etc/velo-deploy"
	ConfigFile      = "config.json"
	AppsDir         = "/opt/deploy/apps"
	LogsDir         = "/var/log/velo-deploy"
	CaddyConf       = "/etc/caddy/conf.d"
	NVMDir          = "/opt/nvm"
	legacyConfigDir = "/etc/deploy"
)

type Config struct {
	CaddyConfDir     string              `json:"caddy_conf_dir"`
	AppsDir          string              `json:"apps_dir"`
	LogsDir          string              `json:"logs_dir"`
	DaemonPort       string              `json:"daemon_port"`
	NVMDir           string              `json:"nvm_dir"`
	WebhookSecret    string              `json:"webhook_secret,omitempty"`
	DuckDNSSubdomain string              `json:"duckdns_subdomain,omitempty"`
	DuckDNSToken     string              `json:"duckdns_token,omitempty"`
	Apps             map[string]*AppMeta `json:"apps"`
}

type AppMeta struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	RepoURL      string `json:"repo_url"`
	Branch       string `json:"branch"`
	NodeVer      string `json:"node_version"`
	Port         int    `json:"port"`
	Domain       string `json:"domain"`
	Path         string `json:"path"`
	Alias        string `json:"alias"`
	NodePath     string `json:"node_path"`
	EntryPoint   string `json:"entry_point"`
	OutputDir    string `json:"output_dir"`
	BuildCommand string `json:"build_command"`
	StartCommand string `json:"start_command"`
}

const (
	AppTypeNode   = "node"
	AppTypeStatic = "static"
)

func ApplyEnv() {
	if v := os.Getenv("VELO_CONFIG"); v != "" {
		ConfigDir = filepath.Dir(v)
		ConfigFile = filepath.Base(v)
	}
	if v := os.Getenv("VELO_APPS_DIR"); v != "" {
		AppsDir = v
	}
	if v := os.Getenv("VELO_LOGS_DIR"); v != "" {
		LogsDir = v
	}
}

func EnvFile(appName string) string {
	return filepath.Join(ConfigDir, "apps", appName+".env")
}

func WebhookSecretFile() string {
	return filepath.Join(ConfigDir, "webhook.secret")
}

func CaddySnippetFile(appName string) string {
	return filepath.Join(ConfigDir, "caddy", appName+".snippet")
}

func DefaultConfig() *Config {
	return &Config{
		CaddyConfDir: CaddyConf,
		AppsDir:      AppsDir,
		LogsDir:      LogsDir,
		DaemonPort:   "9999",
		NVMDir:       NVMDir,
		Apps:         make(map[string]*AppMeta),
	}
}

func Load() (*Config, error) {
	ApplyEnv()
	cfgPath := filepath.Join(ConfigDir, ConfigFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		legacyPath := filepath.Join(legacyConfigDir, ConfigFile)
		if ConfigDir == "/etc/velo-deploy" {
			if legacy, legacyErr := os.ReadFile(legacyPath); legacyErr == nil {
				data = legacy
				err = nil
			}
		}
	}
	if err != nil {
		return DefaultConfig(), nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Apps == nil {
		cfg.Apps = make(map[string]*AppMeta)
	}
	if cfg.CaddyConfDir == "" {
		cfg.CaddyConfDir = CaddyConf
	}
	if cfg.AppsDir == "" {
		cfg.AppsDir = AppsDir
	}
	if cfg.LogsDir == "" {
		cfg.LogsDir = LogsDir
	}
	if cfg.DaemonPort == "" {
		cfg.DaemonPort = "9999"
	}
	if cfg.NVMDir == "" {
		cfg.NVMDir = NVMDir
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(ConfigDir, "apps"), 0750); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(ConfigDir, "caddy"), 0755); err != nil {
		return err
	}
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
	return 0
}

func (c *Config) GetApp(appName string) *AppMeta {
	return c.Apps[appName]
}

func (c *Config) ResolveWebhookSecret() string {
	if v := os.Getenv("VELO_WEBHOOK_SECRET"); v != "" {
		return v
	}
	if c != nil && c.WebhookSecret != "" {
		return c.WebhookSecret
	}
	data, err := os.ReadFile(WebhookSecretFile())
	if err != nil {
		return ""
	}
	return string(bytesTrimSpace(data))
}

func bytesTrimSpace(data []byte) []byte {
	start := 0
	end := len(data)
	for start < end && (data[start] == ' ' || data[start] == '\n' || data[start] == '\r' || data[start] == '\t') {
		start++
	}
	for end > start && (data[end-1] == ' ' || data[end-1] == '\n' || data[end-1] == '\r' || data[end-1] == '\t') {
		end--
	}
	return data[start:end]
}
