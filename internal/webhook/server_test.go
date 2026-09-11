package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"velo-deploy/internal/config"
)

func TestValidSignature(t *testing.T) {
	body := []byte(`{"ok":true}`)
	mac := hmac.New(sha256.New, []byte("s3cret"))
	mac.Write(body)
	header := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !ValidSignature("s3cret", header, body) {
		t.Fatal("expected valid signature")
	}
	if ValidSignature("s3cret", "sha256=deadbeef", body) {
		t.Fatal("expected invalid signature")
	}
	if ValidSignature("", header, body) {
		t.Fatal("empty secret must fail")
	}
}

func TestHealth(t *testing.T) {
	srv := &Server{Cfg: config.DefaultConfig()}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWebhookRejectsMissingSecret(t *testing.T) {
	srv := &Server{Cfg: config.DefaultConfig()}
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	srv.handleWebhook(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWebhookPing(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WebhookSecret = "s3cret"
	body := []byte(`{"zen":"keep it simple","hook_id":1}`)
	mac := hmac.New(sha256.New, []byte("s3cret"))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	srv := &Server{Cfg: cfg}
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(body))
	req.Header.Set("X-Hub-Signature-256", sig)
	req.Header.Set("X-GitHub-Event", "ping")
	rec := httptest.NewRecorder()
	srv.handleWebhook(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
