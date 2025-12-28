package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/alexcostache/Xplorer/internal/fileops"
)

// TestProgressTracking tests the progress tracking functionality
func TestProgressTracking(t *testing.T) {
	manager := fileops.NewManager()
	
	// Test initial state
	if manager.IsOperationActive() {
		t.Error("Expected no active operation initially")
	}
	
	progress := manager.GetProgress()
	if progress == nil {
		t.Fatal("Expected progress info to be initialized")
	}
	
	if progress.Active {
		t.Error("Expected progress to be inactive initially")
	}
}

// TestProgressPercentage tests progress percentage calculation
func TestProgressPercentage(t *testing.T) {
	progress := &fileops.ProgressInfo{
		TotalBytes:     1000,
		ProcessedBytes: 500,
	}
	
	percent := progress.GetProgressPercent()
	if percent != 50 {
		t.Errorf("Expected 50%%, got %d%%", percent)
	}
	
	// Test zero total bytes
	progress.TotalBytes = 0
	percent = progress.GetProgressPercent()
	if percent != 0 {
		t.Errorf("Expected 0%% for zero total bytes, got %d%%", percent)
	}
}

// TestProgressSpeed tests speed calculation
func TestProgressSpeed(t *testing.T) {
	progress := &fileops.ProgressInfo{
		ProcessedBytes: 1000,
		StartTime:      time.Now().Add(-1 * time.Second),
	}
	
	speed := progress.GetSpeed()
	// Speed should be approximately 1000 bytes/second
	if speed < 900 || speed > 1100 {
		t.Errorf("Expected speed around 1000 B/s, got %.2f B/s", speed)
	}
}

// TestCopyWithProgress tests file copy operation with progress tracking
func TestCopyWithProgress(t *testing.T) {
	// Create temporary directories
	srcDir, err := ioutil.TempDir("", "xp-test-src-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)
	
	dstDir, err := ioutil.TempDir("", "xp-test-dst-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dstDir)
	
	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, filename := range testFiles {
		path := filepath.Join(srcDir, filename)
		content := []byte("Test content for " + filename)
		if err := ioutil.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Create manager and copy files
	manager := fileops.NewManager()
	
	// Select files for copy
	var srcPaths []string
	for _, filename := range testFiles {
		srcPaths = append(srcPaths, filepath.Join(srcDir, filename))
	}
	
	manager.Copy(srcPaths)
	
	// Verify clipboard
	if !manager.HasClipboard() {
		t.Error("Expected clipboard to have files")
	}
	
	count, op := manager.GetClipboardInfo()
	if count != len(testFiles) {
		t.Errorf("Expected %d files in clipboard, got %d", len(testFiles), count)
	}
	if op != fileops.OpCopy {
		t.Error("Expected copy operation")
	}
	
	// Paste files
	done := make(chan error, 1)
	go func() {
		done <- manager.Paste(dstDir)
	}()
	
	// Wait a bit to check progress
	time.Sleep(50 * time.Millisecond)
	
	// Check if operation is active
	if !manager.IsOperationActive() {
		t.Log("Operation completed too quickly to check progress")
	} else {
		progress := manager.GetProgress()
		if progress.Operation != fileops.OpCopy {
			t.Error("Expected copy operation in progress")
		}
		if progress.TotalFiles != len(testFiles) {
			t.Errorf("Expected %d total files, got %d", len(testFiles), progress.TotalFiles)
		}
	}
	
	// Wait for completion
	if err := <-done; err != nil {
		t.Errorf("Paste operation failed: %v", err)
	}
	
	// Verify files were copied
	for _, filename := range testFiles {
		dstPath := filepath.Join(dstDir, filename)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist in destination", filename)
		}
	}
	
	// Verify operation is no longer active
	time.Sleep(100 * time.Millisecond)
	if manager.IsOperationActive() {
		t.Error("Expected operation to be inactive after completion")
	}
}

// TestDeleteWithProgress tests file delete operation with progress tracking
func TestDeleteWithProgress(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "xp-test-delete-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create test files
	testFiles := []string{"delete1.txt", "delete2.txt"}
	var filePaths []string
	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := ioutil.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		filePaths = append(filePaths, path)
	}
	
	// Create manager and delete files
	manager := fileops.NewManager()
	
	done := make(chan error, 1)
	go func() {
		done <- manager.Delete(filePaths)
	}()
	
	// Wait a bit to check progress
	time.Sleep(50 * time.Millisecond)
	
	// Check if operation is active or completed
	if manager.IsOperationActive() {
		progress := manager.GetProgress()
		if progress.Operation != fileops.OpDelete {
			t.Error("Expected delete operation in progress")
		}
	}
	
	// Wait for completion
	if err := <-done; err != nil {
		t.Errorf("Delete operation failed: %v", err)
	}
	
	// Verify files were deleted
	for _, path := range filePaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to be deleted", path)
		}
	}
}

// TestProgressConcurrency tests thread-safe progress updates
func TestProgressConcurrency(t *testing.T) {
	progress := &fileops.ProgressInfo{
		TotalBytes: 10000,
		StartTime:  time.Now(),
	}
	
	// Simulate concurrent updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(val int64) {
			for j := 0; j < 100; j++ {
				progress.Mu.Lock()
				progress.ProcessedBytes += val
				progress.Mu.Unlock()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}(int64(i + 1))
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify we can safely read progress
	percent := progress.GetProgressPercent()
	speed := progress.GetSpeed()
	
	if percent < 0 || percent > 100 {
		t.Errorf("Invalid progress percentage: %d", percent)
	}
	
	if speed < 0 {
		t.Errorf("Invalid speed: %.2f", speed)
	}
}

// TestProgressWithLargeFiles tests progress tracking with larger files
func TestProgressWithLargeFiles(t *testing.T) {
	// Create temporary directories
	srcDir, err := ioutil.TempDir("", "xp-test-large-src-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)
	
	dstDir, err := ioutil.TempDir("", "xp-test-large-dst-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dstDir)
	
	// Create a larger test file (1MB)
	largeFile := filepath.Join(srcDir, "large.bin")
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := ioutil.WriteFile(largeFile, data, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Copy the file
	manager := fileops.NewManager()
	manager.Copy([]string{largeFile})
	
	progressChecked := false
	done := make(chan error, 1)
	go func() {
		done <- manager.Paste(dstDir)
	}()
	
	// Monitor progress
	for i := 0; i < 20; i++ {
		time.Sleep(10 * time.Millisecond)
		if manager.IsOperationActive() {
			progressChecked = true
			progress := manager.GetProgress()
			
			// Verify progress is being tracked
			if progress.TotalBytes == 0 {
				t.Error("Expected total bytes to be set")
			}
			
			percent := progress.GetProgressPercent()
			if percent < 0 || percent > 100 {
				t.Errorf("Invalid progress: %d%%", percent)
			}
			
			speed := progress.GetSpeed()
			if speed < 0 {
				t.Errorf("Invalid speed: %.2f", speed)
			}
		}
	}
	
	// Wait for completion
	if err := <-done; err != nil {
		t.Errorf("Paste operation failed: %v", err)
	}
	
	if !progressChecked {
		t.Log("Operation completed too quickly to check progress (this is OK for fast systems)")
	}
	
	// Verify file was copied
	dstFile := filepath.Join(dstDir, "large.bin")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Expected large file to be copied")
	}
}

// Made with Bob
