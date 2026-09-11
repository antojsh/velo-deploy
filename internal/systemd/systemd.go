package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"velo-deploy/internal/config"
)

var systemdDir = "/etc/systemd/system"

var execCommand = exec.Command

var healthTimeout = 30 * time.Second
var healthInterval = time.Second

func GenerateService(appName, nodePath, workDir, entryPoint, username, group string) error {
	return writeUnit(appName, username, group, workDir, nodePath, entryPoint, "", 0)
}

func GenerateCommandService(appName, nodePath, workDir, startCommand, username, group string) error {
	return writeUnit(appName, username, group, workDir, nodePath, "", startCommand, 0)
}

func GenerateAppService(appName, nodePath, workDir, entryPoint, startCommand, username, group string, port int) error {
	return writeUnit(appName, username, group, workDir, nodePath, entryPoint, startCommand, port)
}

func writeUnit(appName, username, group, workDir, nodePath, entryPoint, startCommand string, port int) error {
	nodePathPosix := filepath.ToSlash(nodePath)
	nodeDir := path.Dir(nodePathPosix)
	workDirPosix := filepath.ToSlash(workDir)
	execStart := fmt.Sprintf("%s %s", nodePathPosix, entryPoint)
	if strings.TrimSpace(startCommand) != "" {
		execStart = "/bin/bash -lc " + shellQuote("exec "+startCommand)
	}

	envFile := filepath.ToSlash(config.EnvFile(appName))
	portLine := ""
	if port > 0 {
		portLine = fmt.Sprintf("Environment=PORT=%d\nEnvironment=HOST=127.0.0.1\n", port)
	}

	unit := fmt.Sprintf(`[Unit]
Description=Velo Deploy app: %s
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=%s
Group=%s
WorkingDirectory=%s
ExecStart=%s
Restart=on-failure
RestartSec=5
Environment=NODE_ENV=production
%sEnvironment=PATH=%s:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=-%s

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
ReadWritePaths=%s /tmp
PrivateDevices=true
ProtectKernelTunables=true
RestrictNamespaces=true

[Install]
WantedBy=multi-user.target
`, appName, username, group, workDirPosix, execStart, portLine, nodeDir, envFile, workDirPosix)

	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return err
	}
	svcPath := filepath.Join(systemdDir, UnitName(appName))
	return os.WriteFile(svcPath, []byte(unit), 0644)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func RemoveService(appName string) error {
	var firstErr error
	for _, name := range unitCandidates(appName) {
		svcPath := filepath.Join(systemdDir, name)
		if err := os.Remove(svcPath); err != nil && !os.IsNotExist(err) && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func DaemonReload() error {
	return runCmd("systemctl", "daemon-reload")
}

func EnableApp(appName string) error {
	return runCmd("systemctl", "enable", UnitName(appName))
}

func StartApp(appName string) error {
	return runCmd("systemctl", "start", resolveExistingUnit(appName))
}

func StopApp(appName string) error {
	return runCmd("systemctl", "stop", resolveExistingUnit(appName))
}

func DisableApp(appName string) error {
	var firstErr error
	for _, name := range unitCandidates(appName) {
		if err := runCmd("systemctl", "disable", name); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func ResetFailed(appName string) error {
	return runCmd("systemctl", "reset-failed", resolveExistingUnit(appName))
}

func ServiceExists(appName string) bool {
	for _, name := range unitCandidates(appName) {
		if _, err := os.Stat(filepath.Join(systemdDir, name)); err == nil {
			return true
		}
	}
	return false
}

func RestartApp(appName string) error {
	return runCmd("systemctl", "restart", resolveExistingUnit(appName))
}

func StatusApp(appName string) (string, error) {
	out, err := execCommand("systemctl", "status", resolveExistingUnit(appName)).CombinedOutput()
	return string(out), err
}

func IsActive(appName string) bool {
	out, _ := execCommand("systemctl", "is-active", resolveExistingUnit(appName)).Output()
	return strings.TrimSpace(string(out)) == "active"
}

func WaitHealthy(appName string) error {
	if healthTimeout <= 0 {
		return nil
	}
	deadline := time.Now().Add(healthTimeout)
	for time.Now().Before(deadline) {
		if IsActive(appName) {
			return nil
		}
		time.Sleep(healthInterval)
	}
	if IsActive(appName) {
		return nil
	}
	return fmt.Errorf("service %s did not become healthy in time", appName)
}

func ShowLogs(appName string, follow bool, lines int) error {
	if lines <= 0 {
		lines = 200
	}
	args := []string{"journalctl", "-u", resolveExistingUnit(appName), "--no-pager", "-n", strconv.Itoa(lines)}
	if follow {
		args = append(args, "-f")
	}
	cmd := execCommand(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ListApps(cfg interface{ Save() error }) {
	names := discoveredApps()
	if len(names) == 0 {
		fmt.Println("No apps deployed.")
		return
	}
	fmt.Printf("%-20s %-10s\n", "APP", "STATUS")
	fmt.Println(strings.Repeat("-", 32))
	for _, appName := range names {
		active := "stopped"
		if IsActive(appName) {
			active = "active"
		}
		fmt.Printf("%-20s %-10s\n", appName, active)
	}
}

func CreateUser(username string) error {
	_, err := execCommand("id", username).CombinedOutput()
	if err == nil {
		return nil
	}
	return runCmd("useradd", "-r", "-s", "/usr/sbin/nologin", "-d", "/nonexistent", "-M", username)
}

func RemoveUser(username string) error {
	return runCmd("userdel", "-r", username)
}

func resolveExistingUnit(appName string) string {
	legacy := filepath.Join(systemdDir, LegacyUnitName(appName))
	current := filepath.Join(systemdDir, UnitName(appName))
	if _, err := os.Stat(current); err == nil {
		return UnitName(appName)
	}
	if _, err := os.Stat(legacy); err == nil {
		return LegacyUnitName(appName)
	}
	return UnitName(appName)
}

func discoveredApps() []string {
	seen := map[string]bool{}
	var names []string
	for _, pattern := range []string{
		filepath.Join(systemdDir, "velo-*.service"),
		filepath.Join(systemdDir, "deploy-*.service"),
	} {
		services, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, svc := range services {
			base := strings.TrimSuffix(filepath.Base(svc), ".service")
			if base+".service" == watcherUnit {
				continue
			}
			appName := strings.TrimPrefix(base, "velo-")
			appName = strings.TrimPrefix(appName, "deploy-")
			if appName == "" || seen[appName] {
				continue
			}
			seen[appName] = true
			names = append(names, appName)
		}
	}
	return names
}

func runCmd(name string, args ...string) error {
	cmd := execCommand(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
