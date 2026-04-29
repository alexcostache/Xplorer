package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoToParentRestoresChildSelection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "navigator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parentDir := filepath.Join(tmpDir, "folder1")
	childDir := filepath.Join(parentDir, "folder2")
	siblingDir := filepath.Join(parentDir, "folder3")

	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(siblingDir, 0755); err != nil {
		t.Fatal(err)
	}

	nav := NewNavigator()
	nav.SetCurrentDir(parentDir)

	foundChild := false
	for i, entry := range nav.GetFileList() {
		if entry.Name() == "folder2" {
			nav.SetCursor(i)
			foundChild = true
			break
		}
	}
	if !foundChild {
		t.Fatal("expected folder2 to be present in parent directory listing")
	}

	if !nav.EnterDirectory() {
		t.Fatal("expected to enter folder2")
	}

	if nav.GetCurrentDir() != childDir {
		t.Fatalf("expected current dir to be %s, got %s", childDir, nav.GetCurrentDir())
	}

	if !nav.GoToParent() {
		t.Fatal("expected to navigate back to parent")
	}

	if nav.GetCurrentDir() != parentDir {
		t.Fatalf("expected current dir to be %s, got %s", parentDir, nav.GetCurrentDir())
	}

	selected := nav.GetSelectedFile()
	if selected == nil {
		t.Fatal("expected a selected entry after returning to parent")
	}
	if selected.Name() != "folder2" {
		t.Fatalf("expected folder2 to be selected after returning, got %s", selected.Name())
	}
}

// Made with Bob
