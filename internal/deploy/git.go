package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var execCommand = exec.Command

func gitClone(repoURL, dest, branch string) error {
	args := []string{"clone", "--depth", "1"}
	if strings.TrimSpace(branch) != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, repoURL, dest)
	if err := runGit(args...); err != nil {
		return fmt.Errorf("%w: clone %s: %v", ErrGit, repoURL, err)
	}
	return nil
}

func gitPull(dir, branch string) error {
	if err := runGit("-C", dir, "fetch", "--depth", "1", "origin"); err != nil {
		return fmt.Errorf("%w: fetch: %v", ErrGit, err)
	}
	ref := "origin/HEAD"
	if strings.TrimSpace(branch) != "" {
		ref = "origin/" + branch
	}
	if err := runGit("-C", dir, "reset", "--hard", ref); err != nil {
		return fmt.Errorf("%w: reset: %v", ErrGit, err)
	}
	return nil
}

func gitHead(dir string) string {
	out, err := execCommand("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitIsRepo(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return true
	}
	return runGit("-C", dir, "rev-parse", "--is-inside-work-tree") == nil
}

func runGit(args ...string) error {
	cmd := execCommand("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
