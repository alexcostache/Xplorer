package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetThemesDirPrefersExecutableThemesDirectory(t *testing.T) {
	execPath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	execThemesDir := filepath.Join(filepath.Dir(execPath), "themes")
	cleanupExecThemes := false
	if _, err := os.Stat(execThemesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(execThemesDir, 0755); err != nil {
			t.Fatal(err)
		}
		cleanupExecThemes = true
	}

	testThemePath := filepath.Join(execThemesDir, "zz-test-installed-theme.json")
	if err := os.WriteFile(testThemePath, []byte(`{
  "name": "Installed Test Theme",
  "colors": {
    "text": "white",
    "background": "black",
    "highlight": "magenta",
    "highlight_text": "white",
    "footer": "cyan",
    "footer_bg": "black",
    "address_bar": "magenta",
    "address_bar_bg": "black",
    "separator": "magenta",
    "dim": "white",
    "filter": "white",
    "filter_bg": "magenta",
    "dir": "cyan"
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testThemePath)

	if cleanupExecThemes {
		defer os.Remove(execThemesDir)
	}

	mgr := NewManager()

	found := false
	for _, th := range mgr.GetThemes() {
		if th.Name == "Installed Test Theme" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected installed theme to be discovered from executable themes directory")
	}
}

func TestGetThemesDirFallsBackToExecutableParentThemesDirectory(t *testing.T) {
	execPath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	execDir := filepath.Dir(execPath)
	execThemesDir := filepath.Join(execDir, "themes")
	parentThemesDir := filepath.Join(execDir, "..", "themes")

	backupDir := ""
	if info, err := os.Stat(execThemesDir); err == nil && info.IsDir() {
		backupDir = execThemesDir + "-backup-test"
		_ = os.RemoveAll(backupDir)
		if err := os.Rename(execThemesDir, backupDir); err != nil {
			t.Fatal(err)
		}
		defer os.Rename(backupDir, execThemesDir)
	}

	if err := os.MkdirAll(parentThemesDir, 0755); err != nil {
		t.Fatal(err)
	}

	testThemePath := filepath.Join(parentThemesDir, "zz-test-parent-theme.json")
	if err := os.WriteFile(testThemePath, []byte(`{
  "name": "Parent Installed Theme",
  "colors": {
    "text": "white",
    "background": "black",
    "highlight": "magenta",
    "highlight_text": "white",
    "footer": "cyan",
    "footer_bg": "black",
    "address_bar": "magenta",
    "address_bar_bg": "black",
    "separator": "magenta",
    "dim": "white",
    "filter": "white",
    "filter_bg": "magenta",
    "dir": "cyan"
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testThemePath)

	mgr := NewManager()

	found := false
	for _, th := range mgr.GetThemes() {
		if th.Name == "Parent Installed Theme" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected installed theme to be discovered from executable parent themes directory")
	}
}

// Made with Bob
