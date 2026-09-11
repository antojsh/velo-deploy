package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"velo-deploy/internal/config"
	"velo-deploy/internal/deploy"
)

type Server struct {
	Cfg  *config.Config
	Addr string
}

type pushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
		SSHURL   string `json:"ssh_url"`
	} `json:"repository"`
}

var runAsync = true

func ValidSignature(secret, header string, body []byte) bool {
	if secret == "" || header == "" {
		return false
	}
	header = strings.TrimSpace(header)
	header = strings.TrimPrefix(header, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(header)))
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/webhook", s.handleWebhook)
	return http.ListenAndServe(s.Addr, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	secret := s.Cfg.ResolveWebhookSecret()
	if secret == "" {
		http.Error(w, "webhook secret is not configured", http.StatusUnauthorized)
		return
	}
	if !ValidSignature(secret, r.Header.Get("X-Hub-Signature-256"), body) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event == "ping" || event == "" && looksLikePing(body) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"event":"ping"}`))
		return
	}

	var payload pushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}
	branch := deploy.BranchFromRef(payload.Ref)
	app := s.matchApp(payload)
	if app == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"matched":false}`))
		return
	}
	if !deploy.ShouldDeployBranch(app.Branch, branch) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"ignored":"branch"}`))
		return
	}

	job := func() {
		if err := deploy.Redeploy(s.Cfg, app); err != nil {
			log.Printf("redeploy %s failed: %v", app.Name, err)
		}
	}
	if runAsync {
		go job()
	} else {
		job()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"ok":true,"app":"` + app.Name + `"}`))
}

func (s *Server) matchApp(payload pushPayload) *config.AppMeta {
	candidates := []string{payload.Repository.CloneURL, payload.Repository.HTMLURL, payload.Repository.SSHURL}
	for _, app := range s.Cfg.Apps {
		for _, candidate := range candidates {
			if deploy.ReposMatch(app.RepoURL, candidate) {
				return app
			}
		}
	}
	return nil
}

func looksLikePing(body []byte) bool {
	return strings.Contains(string(body), `"zen"`) && strings.Contains(string(body), `"hook_id"`)
}

func BindAddr(host, port string) string {
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" {
		port = "9999"
	}
	return net.JoinHostPort(host, port)
}
