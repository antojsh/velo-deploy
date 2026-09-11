package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectAppType(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"start":"node index.js","build":"tsc"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if got := DetectAppType(dir, ""); got != "node" {
		t.Fatalf("start script should win, got %s", got)
	}
	if got := DetectAppType(dir, "static"); got != "static" {
		t.Fatalf("explicit type, got %s", got)
	}

	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := DetectAppType(staticDir, ""); got != "static" {
		t.Fatalf("index.html should be static, got %s", got)
	}
}

func TestAppNameFromRepo(t *testing.T) {
	if got := AppNameFromRepo("https://github.com/user/My_API.git"); got != "my-api" {
		t.Fatalf("got %s", got)
	}
}

func TestNormalizeRepoURL(t *testing.T) {
	a := NormalizeRepoURL("https://github.com/Acme/API.git")
	b := NormalizeRepoURL("git@github.com:Acme/API.git")
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestShouldDeployBranch(t *testing.T) {
	if !ShouldDeployBranch("", "main") || !ShouldDeployBranch("", "master") {
		t.Fatal("default branches")
	}
	if ShouldDeployBranch("", "feat") {
		t.Fatal("feature branch should be ignored")
	}
	if !ShouldDeployBranch("develop", "develop") {
		t.Fatal("configured branch")
	}
}

func TestSanitizeName(t *testing.T) {
	if _, err := SanitizeName("My App"); err == nil {
		t.Fatal("expected error")
	}
	got, err := SanitizeName("api-1")
	if err != nil || got != "api-1" {
		t.Fatalf("got %s %v", got, err)
	}
}
