package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, os.Getenv("HELPER_STDOUT"))
	code := 0
	if v := os.Getenv("HELPER_EXIT"); v != "" {
		code, _ = strconv.Atoi(v)
	}
	os.Exit(code)
}

func helperCommand(stdout string, exitCode int) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_STDOUT="+stdout,
		"HELPER_EXIT="+strconv.Itoa(exitCode),
	)
	return cmd
}

func TestGenerateService(t *testing.T) {
	tmpDir := t.TempDir()
	svcPath := filepath.Join(tmpDir, "velo-test.service")

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	err := GenerateService("test", "/opt/node/20/bin/node", "/opt/deploy/test", "index.js", "deploy-test", "deploy-test")
	assert.NoError(t, err)

	content, err := os.ReadFile(svcPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "Description=Velo Deploy app: test")
	assert.Contains(t, string(content), "User=deploy-test")
	assert.Contains(t, string(content), "Group=deploy-test")
	assert.Contains(t, string(content), "WorkingDirectory=/opt/deploy/test")
	assert.Contains(t, string(content), "ExecStart=/opt/node/20/bin/node index.js")
	assert.Contains(t, string(content), "Restart=on-failure")
	assert.Contains(t, string(content), "NODE_ENV=production")
	assert.Contains(t, string(content), "ProtectSystem=full")
	assert.Contains(t, string(content), "ProtectHome=true")
	assert.Contains(t, string(content), "PrivateTmp=true")
	assert.Contains(t, string(content), "NoNewPrivileges=true")
	assert.Contains(t, string(content), "PrivateDevices=true")
	assert.Contains(t, string(content), "ProtectKernelTunables=true")
	assert.Contains(t, string(content), "RestrictNamespaces=true")
	assert.NotContains(t, string(content), "MemoryDenyWriteExecute")
	assert.Contains(t, string(content), "ReadWritePaths=/opt/deploy/test /tmp")
	assert.Contains(t, string(content), "WantedBy=multi-user.target")
}

func TestGenerateService_NodePathExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	svcPath := filepath.Join(tmpDir, "velo-myapp.service")

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	err := GenerateService("myapp", "/opt/nvm/versions/node/v20.0.0/bin/node", "/opt/deploy/myapp", "server.js", "deploy-myapp", "deploy-myapp")
	assert.NoError(t, err)

	content, err := os.ReadFile(svcPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "PATH=/opt/nvm/versions/node/v20.0.0/bin")
	assert.Contains(t, string(content), "ExecStart=/opt/nvm/versions/node/v20.0.0/bin/node server.js")
}

func TestGenerateCommandService(t *testing.T) {
	tmpDir := t.TempDir()
	svcPath := filepath.Join(tmpDir, "velo-web.service")

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	err := GenerateCommandService("web", "/opt/deploy/node/24/bin/node", "/opt/deploy/apps/web", "npm run start", "deploy-web", "deploy-web")
	assert.NoError(t, err)

	content, err := os.ReadFile(svcPath)
	require.NoError(t, err)

	unit := string(content)
	assert.Contains(t, unit, "WorkingDirectory=/opt/deploy/apps/web")
	assert.Contains(t, unit, "ExecStart=/bin/bash -lc 'exec npm run start'")
	assert.Contains(t, unit, "PATH=/opt/deploy/node/24/bin")
	assert.Contains(t, unit, "ReadWritePaths=/opt/deploy/apps/web /tmp")
}

func TestRemoveService(t *testing.T) {
	tmpDir := t.TempDir()
	svcPath := filepath.Join(tmpDir, "velo-test.service")

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	err := os.WriteFile(svcPath, []byte("test"), 0644)
	require.NoError(t, err)

	err = RemoveService("test")
	assert.NoError(t, err)

	_, err = os.Stat(svcPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveService_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	err := RemoveService("nonexistent")
	assert.NoError(t, err)
}

func TestIsActive(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("active\n", 0)
	}
	defer func() { execCommand = exec.Command }()

	active := IsActive("test")
	assert.True(t, active)
}

func TestIsActive_Inactive(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("inactive\n", 0)
	}
	defer func() { execCommand = exec.Command }()

	active := IsActive("test")
	assert.False(t, active)
}

func TestIsActive_Failed(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 1)
	}
	defer func() { execCommand = exec.Command }()

	active := IsActive("test")
	assert.False(t, active)
}

func TestStatusApp(t *testing.T) {
	expectedOutput := "active (running)"
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand(expectedOutput+"\n", 0)
	}
	defer func() { execCommand = exec.Command }()

	output, err := StatusApp("test")
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput+"\n", output)
}

func TestShowLogs_NoFollow(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "journalctl" {
			return helperCommand("log line 1\nlog line 2\n", 0)
		}
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := ShowLogs("test", false, 0)
	assert.NoError(t, err)
}

func TestShowLogs_Follow(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "journalctl" {
			return helperCommand("following logs\n", 0)
		}
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := ShowLogs("test", true, 50)
	assert.NoError(t, err)
}

func TestListApps_NoServices(t *testing.T) {
	tmpDir := t.TempDir()

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("inactive\n", 0)
	}
	defer func() { execCommand = exec.Command }()

	ListApps(nil)
}

func TestListApps_WithServices(t *testing.T) {
	tmpDir := t.TempDir()

	svcPath1 := filepath.Join(tmpDir, "deploy-app1.service")
	svcPath2 := filepath.Join(tmpDir, "deploy-app2.service")
	err := os.WriteFile(svcPath1, []byte("test"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(svcPath2, []byte("test"), 0644)
	require.NoError(t, err)

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	callCount := 0
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "systemctl" && args[0] == "is-active" {
			callCount++
			if callCount%2 == 0 {
				return helperCommand("inactive\n", 0)
			}
			return helperCommand("active\n", 0)
		}
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	ListApps(nil)
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "id" {
			return helperCommand("uid=1000(test) gid=1000(test)\n", 0)
		}
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := CreateUser("test")
	assert.NoError(t, err)
}

func TestCreateUser_NewUser(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "id" {
			return helperCommand("", 1)
		}
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := CreateUser("newuser")
	assert.NoError(t, err)
}

func TestRemoveUser(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := RemoveUser("testuser")
	assert.NoError(t, err)
}

func TestDaemonReload(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := DaemonReload()
	assert.NoError(t, err)
}

func TestEnableApp(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := EnableApp("test")
	assert.NoError(t, err)
}

func TestStartApp(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := StartApp("test")
	assert.NoError(t, err)
}

func TestStopApp(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := StopApp("test")
	assert.NoError(t, err)
}

func TestRestartApp(t *testing.T) {
	execCommand = func(name string, args ...string) *exec.Cmd {
		return helperCommand("", 0)
	}
	defer func() { execCommand = exec.Command }()

	err := RestartApp("test")
	assert.NoError(t, err)
}

func TestGenerateService_AllSandboxOptions(t *testing.T) {
	tmpDir := t.TempDir()
	svcPath := filepath.Join(tmpDir, "velo-sandbox.service")

	originalSystemdDir := systemdDir
	systemdDir = tmpDir
	defer func() { systemdDir = originalSystemdDir }()

	err := GenerateService("sandbox", "/usr/bin/node", "/opt/deploy/sandbox", "app.js", "deploy-sandbox", "deploy-sandbox")
	assert.NoError(t, err)

	content, err := os.ReadFile(svcPath)
	require.NoError(t, err)
	conf := string(content)

	assert.Contains(t, conf, "ProtectSystem=full")
	assert.Contains(t, conf, "ProtectHome=true")
	assert.Contains(t, conf, "PrivateTmp=true")
	assert.Contains(t, conf, "NoNewPrivileges=true")
	assert.Contains(t, conf, "PrivateDevices=true")
	assert.Contains(t, conf, "ReadWritePaths=/opt/deploy/sandbox /tmp")
	assert.NotContains(t, conf, "MemoryDenyWriteExecute")
}
