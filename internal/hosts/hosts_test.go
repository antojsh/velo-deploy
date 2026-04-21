package hosts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAlias_Success(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	
	initialContent := "127.0.0.1 localhost\n::1 localhost\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)
	
	originalHostsFile := hostsFile
	hostsFile = hostsPath
	defer func() { hostsFile = originalHostsFile }()
	
	err = AddAlias("myapp.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "127.0.0.1  myapp.local")
}

func TestAddAlias_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	
	initialContent := "127.0.0.1 localhost\n127.0.0.1  myapp.local\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)
	
	originalHostsFile := hostsFile
	hostsFile = hostsPath
	defer func() { hostsFile = originalHostsFile }()
	
	err = AddAlias("myapp.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	lines := splitLines(string(content))
	count := 0
	for _, line := range lines {
		if line == "127.0.0.1  myapp.local" {
			count++
		}
	}
	assert.Equal(t, 1, count, "should not add duplicate entry")
}

func TestAddAlias_FileNotFound(t *testing.T) {
	originalHostsFile := hostsFile
	hostsFile = "/nonexistent/path/hosts"
	defer func() { hostsFile = originalHostsFile }()
	
	err := AddAlias("myapp.local")
	assert.Error(t, err)
}

func TestRemoveAlias_Success(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	
	initialContent := "127.0.0.1 localhost\n127.0.0.1  myapp.local\n::1 localhost\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)
	
	originalHostsFile := hostsFile
	hostsFile = hostsPath
	defer func() { hostsFile = originalHostsFile }()
	
	err = RemoveAlias("myapp.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "myapp.local")
	assert.Contains(t, string(content), "localhost")
}

func TestRemoveAlias_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	
	initialContent := "127.0.0.1 localhost\n::1 localhost\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)
	
	originalHostsFile := hostsFile
	hostsFile = hostsPath
	defer func() { hostsFile = originalHostsFile }()
	
	err = RemoveAlias("nonexistent.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.Equal(t, initialContent, string(content))
}

func TestRemoveAlias_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	
	initialContent := "127.0.0.1 localhost\n127.0.0.1  myapp.local\n127.0.0.1  myapp.local\n::1 localhost\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)
	
	originalHostsFile := hostsFile
	hostsFile = hostsPath
	defer func() { hostsFile = originalHostsFile }()
	
	err = RemoveAlias("myapp.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "myapp.local")
}

func TestRemoveAlias_FileNotFound(t *testing.T) {
	originalHostsFile := hostsFile
	hostsFile = "/nonexistent/path/hosts"
	defer func() { hostsFile = originalHostsFile }()
	
	err := RemoveAlias("myapp.local")
	assert.Error(t, err)
}

func TestSetHostsFile(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "custom-hosts")
	
	originalHostsFile := hostsFile
	SetHostsFile(hostsPath)
	defer func() { hostsFile = originalHostsFile }()
	
	assert.Equal(t, hostsPath, hostsFile)
	
	err := os.WriteFile(hostsPath, []byte("127.0.0.1 localhost\n"), 0644)
	require.NoError(t, err)
	
	err = AddAlias("test.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test.local")
}

func TestAddAlias_DifferentPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	
	initialContent := "127.0.0.1 localhost\n192.168.1.1  myapp.local\n"
	err := os.WriteFile(hostsPath, []byte(initialContent), 0644)
	require.NoError(t, err)
	
	originalHostsFile := hostsFile
	hostsFile = hostsPath
	defer func() { hostsFile = originalHostsFile }()
	
	err = AddAlias("myapp.local")
	assert.NoError(t, err)
	
	content, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	lines := splitLines(string(content))
	count := 0
	for _, line := range lines {
		if line == "127.0.0.1  myapp.local" {
			count++
		}
	}
	assert.Equal(t, 1, count, "should add 127.0.0.1 entry even when 192.168.x.x exists")
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range split(s, '\n') {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func split(s string, sep rune) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || rune(s[i]) == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	return result
}
