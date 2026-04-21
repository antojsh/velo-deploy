package hosts

import (
	"fmt"
	"os"
	"strings"
)

var hostsFile = "/etc/hosts"

// SetHostsFile sets the hosts file path (for testing)
func SetHostsFile(path string) {
	hostsFile = path
}

// AddAlias adds an entry to /etc/hosts for internal service discovery.
func AddAlias(alias string) error {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, alias) && strings.HasPrefix(line, "127.0.0.1") {
			return nil // Already exists
		}
	}

	f, err := os.OpenFile(hostsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n127.0.0.1  %s\n", alias)
	return err
}

// RemoveAlias removes an entry from /etc/hosts.
func RemoveAlias(alias string) error {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var filtered []string
	for _, line := range lines {
		if strings.Contains(line, alias) && strings.HasPrefix(line, "127.0.0.1") {
			continue
		}
		filtered = append(filtered, line)
	}

	return os.WriteFile(hostsFile, []byte(strings.Join(filtered, "\n")), 0644)
}
