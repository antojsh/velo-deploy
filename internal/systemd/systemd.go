package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const systemdDir = "/etc/systemd/system"

var execCommand = exec.Command

// GenerateService creates a systemd unit file for a Node.js app.
func GenerateService(appName, nodePath, workDir, entryPoint, username, group string) error {
	nodeDir := filepath.Dir(nodePath)
	unit := fmt.Sprintf(`[Unit]
Description=Deploy managed app: %s
After=network.target

[Service]
Type=simple
User=%s
Group=%s
WorkingDirectory=%s
ExecStart=%s %s
Restart=always
RestartSec=3
Environment=NODE_ENV=production
Environment=PATH=%s:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Sandboxing
ProtectSystem=full
PrivateTmp=true
NoNewPrivileges=true
ReadWritePaths=%s %s

[Install]
WantedBy=multi-user.target
`, appName, username, group, workDir, nodePath, entryPoint, nodeDir, workDir, nodeDir)

	svcPath := fmt.Sprintf("%s/deploy-%s.service", systemdDir, appName)
	return os.WriteFile(svcPath, []byte(unit), 0644)
}

// RemoveService deletes the systemd unit file.
func RemoveService(appName string) error {
	svcPath := fmt.Sprintf("%s/deploy-%s.service", systemdDir, appName)
	return os.Remove(svcPath)
}

// DaemonReload runs systemctl daemon-reload.
func DaemonReload() error {
	return runCmd("systemctl", "daemon-reload")
}

// EnableApp enables a service to start on boot.
func EnableApp(appName string) error {
	return runCmd("systemctl", "enable", fmt.Sprintf("deploy-%s.service", appName))
}

// StartApp starts a service.
func StartApp(appName string) error {
	return runCmd("systemctl", "start", fmt.Sprintf("deploy-%s.service", appName))
}

// StopApp stops a service.
func StopApp(appName string) error {
	return runCmd("systemctl", "stop", fmt.Sprintf("deploy-%s.service", appName))
}

// RestartApp restarts a service.
func RestartApp(appName string) error {
	return runCmd("systemctl", "restart", fmt.Sprintf("deploy-%s.service", appName))
}

// StatusApp returns the output of systemctl status.
func StatusApp(appName string) (string, error) {
	out, err := execCommand("systemctl", "status", fmt.Sprintf("deploy-%s.service", appName)).CombinedOutput()
	return string(out), err
}

// IsActive checks if a service is active (running).
func IsActive(appName string) bool {
	out, _ := execCommand("systemctl", "is-active", fmt.Sprintf("deploy-%s.service", appName)).Output()
	return strings.TrimSpace(string(out)) == "active"
}

// ShowLogs streams journalctl output for the service.
func ShowLogs(appName string, follow bool) error {
	args := []string{"journalctl", "-u", fmt.Sprintf("deploy-%s.service", appName), "--no-pager"}
	if follow {
		args = append(args, "-f")
	}
	cmd := execCommand(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListApps lists all deploy-managed services.
func ListApps(cfg interface{ Save() error }) {
	services, err := filepath.Glob(fmt.Sprintf("%s/deploy-*.service", systemdDir))
	if err != nil {
		fmt.Println("No services found.")
		return
	}
	if len(services) == 0 {
		fmt.Println("No apps deployed.")
		return
	}
	fmt.Printf("%-20s %-10s %-10s\n", "APP", "STATUS", "PORT")
	fmt.Println(strings.Repeat("-", 40))
	for _, svc := range services {
		name := strings.TrimSuffix(filepath.Base(svc), ".service")
		appName := strings.TrimPrefix(name, "deploy-")
		active := "stopped"
		if IsActive(appName) {
			active = "active"
		}
		fmt.Printf("%-20s %-10s\n", appName, active)
	}
}

// CreateUser creates a system user without login.
func CreateUser(username string) error {
	_, err := execCommand("id", username).CombinedOutput()
	if err == nil {
		return nil
	}
	return runCmd("useradd", "-r", "-s", "/bin/false", username)
}

// RemoveUser deletes a system user.
func RemoveUser(username string) error {
	return runCmd("userdel", "-r", username)
}

func runCmd(name string, args ...string) error {
	cmd := execCommand(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
