package fileops

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestToggleSelection(t *testing.T) {
	m := NewManager()
	
	// Test selecting a file
	path := "/test/file.txt"
	m.ToggleSelection(path)
	
	if !m.IsSelected(path) {
		t.Errorf("Expected file to be selected")
	}
	
	// Test deselecting a file
	m.ToggleSelection(path)
	
	if m.IsSelected(path) {
		t.Errorf("Expected file to be deselected")
	}
}

func TestMultipleSelections(t *testing.T) {
	m := NewManager()
	
	paths := []string{
		"/test/file1.txt",
		"/test/file2.txt",
		"/test/file3.txt",
	}
	
	// Select multiple files
	for _, path := range paths {
		m.ToggleSelection(path)
	}
	
	// Verify all are selected
	for _, path := range paths {
		if !m.IsSelected(path) {
			t.Errorf("Expected %s to be selected", path)
		}
	}
	
	// Verify count
	if m.GetSelectedCount() != 3 {
		t.Errorf("Expected 3 selected files, got %d", m.GetSelectedCount())
	}
	
	// Verify GetSelectedFiles returns all
	selected := m.GetSelectedFiles()
	if len(selected) != 3 {
		t.Errorf("Expected 3 files in GetSelectedFiles, got %d", len(selected))
	}
}

func TestClearSelection(t *testing.T) {
	m := NewManager()
	
	// Select some files
	m.ToggleSelection("/test/file1.txt")
	m.ToggleSelection("/test/file2.txt")
	
	// Clear selection
	m.ClearSelection()
	
	if m.GetSelectedCount() != 0 {
		t.Errorf("Expected 0 selected files after clear, got %d", m.GetSelectedCount())
	}
	
	if m.IsSelected("/test/file1.txt") {
		t.Errorf("Expected file to not be selected after clear")
	}
}

func TestCopyOperation(t *testing.T) {
	m := NewManager()
	
	files := []string{"/test/file1.txt", "/test/file2.txt"}
	m.Copy(files)
	
	if m.operation != OpCopy {
		t.Errorf("Expected operation to be OpCopy")
	}
	
	if len(m.clipboard) != 2 {
		t.Errorf("Expected 2 files in clipboard, got %d", len(m.clipboard))
	}
}

func TestCutOperation(t *testing.T) {
	m := NewManager()
	
	files := []string{"/test/file1.txt", "/test/file2.txt"}
	m.Cut(files)
	
	if m.operation != OpCut {
		t.Errorf("Expected operation to be OpCut")
	}
	
	if len(m.clipboard) != 2 {
		t.Errorf("Expected 2 files in clipboard, got %d", len(m.clipboard))
	}
}

func TestCreateFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "fileops_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	m := NewManager()
	
	// Test creating a file
	filename := "test.txt"
	err = m.CreateFile(tmpDir, filename)
	if err != nil {
		t.Errorf("Failed to create file: %v", err)
	}
	
	// Verify file exists
	filePath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File was not created")
	}
}

func TestCreateFolder(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "fileops_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	m := NewManager()
	
	// Test creating a folder
	foldername := "testfolder"
	err = m.CreateFolder(tmpDir, foldername)
	if err != nil {
		t.Errorf("Failed to create folder: %v", err)
	}
	
	// Verify folder exists
	folderPath := filepath.Join(tmpDir, foldername)
	if stat, err := os.Stat(folderPath); os.IsNotExist(err) || !stat.IsDir() {
		t.Errorf("Folder was not created")
	}
}

func TestRename(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "fileops_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	m := NewManager()
	
	// Create a test file
	oldPath := filepath.Join(tmpDir, "old.txt")
	if err := ioutil.WriteFile(oldPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Rename the file
	newName := "new.txt"
	err = m.Rename(oldPath, newName)
	if err != nil {
		t.Errorf("Failed to rename file: %v", err)
	}
	
	// Verify old file doesn't exist
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("Old file still exists")
	}
	
	// Verify new file exists
	newPath := filepath.Join(tmpDir, newName)
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("New file was not created")
	}
}

func TestDelete(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "fileops_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	m := NewManager()
	
	// Create a test file
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := ioutil.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Delete the file
	err = m.Delete([]string{filePath})
	if err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}
	
	// Verify file doesn't exist
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("File still exists after delete")
	}
}

func TestPasteCopy(t *testing.T) {
	// Create temp directories
	srcDir, err := ioutil.TempDir("", "fileops_test_src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)
	
	dstDir, err := ioutil.TempDir("", "fileops_test_dst")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dstDir)
	
	m := NewManager()
	
	// Create a test file
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := ioutil.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Copy the file
	m.Copy([]string{srcFile})
	
	// Paste to destination
	err = m.Paste(dstDir)
	if err != nil {
		t.Errorf("Failed to paste file: %v", err)
	}
	
	// Verify source still exists
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Errorf("Source file was removed (should be copy)")
	}
	
	// Verify destination exists
	dstFile := filepath.Join(dstDir, "test.txt")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Errorf("Destination file was not created")
	}
	
	// Verify content
	content, err := ioutil.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "test content" {
		t.Errorf("File content mismatch")
	}
}

func TestPasteCut(t *testing.T) {
	// Create temp directories
	srcDir, err := ioutil.TempDir("", "fileops_test_src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)
	
	dstDir, err := ioutil.TempDir("", "fileops_test_dst")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dstDir)
	
	m := NewManager()
	
	// Create a test file
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := ioutil.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Cut the file
	m.Cut([]string{srcFile})
	
	// Paste to destination
	err = m.Paste(dstDir)
	if err != nil {
		t.Errorf("Failed to paste file: %v", err)
	}
	
	// Verify source no longer exists
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists (should be cut)")
	}
	
	// Verify destination exists
	dstFile := filepath.Join(dstDir, "test.txt")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Errorf("Destination file was not created")
	}
}

// Made with Bob
