package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Operation represents a file operation type
type Operation int

const (
	OpNone Operation = iota
	OpCopy
	OpCut
	OpDelete
)

// ProgressInfo contains information about ongoing file operation
type ProgressInfo struct {
	Operation     Operation
	TotalBytes    int64
	ProcessedBytes int64
	TotalFiles    int
	ProcessedFiles int
	CurrentFile   string
	StartTime     time.Time
	Active        bool
	Mu            sync.RWMutex
}

// Manager handles file operations
type Manager struct {
	clipboard      []string  // Files in clipboard
	operation      Operation // Current operation (copy or cut)
	selectedFiles  map[string]bool // Selected files in current directory
	progress       *ProgressInfo
}

// NewManager creates a new file operations manager
func NewManager() *Manager {
	return &Manager{
		clipboard:     make([]string, 0),
		operation:     OpNone,
		selectedFiles: make(map[string]bool),
		progress: &ProgressInfo{
			Active: false,
		},
	}
}

// GetProgress returns the current progress information
func (m *Manager) GetProgress() *ProgressInfo {
	return m.progress
}

// IsOperationActive returns whether a file operation is in progress
func (m *Manager) IsOperationActive() bool {
	m.progress.Mu.RLock()
	defer m.progress.Mu.RUnlock()
	return m.progress.Active
}

// GetProgressPercent returns the progress percentage (0-100)
func (p *ProgressInfo) GetProgressPercent() int {
	p.Mu.RLock()
	defer p.Mu.RUnlock()
	if p.TotalBytes == 0 {
		return 0
	}
	return int((p.ProcessedBytes * 100) / p.TotalBytes)
}

// GetSpeed returns the current operation speed in bytes per second
func (p *ProgressInfo) GetSpeed() float64 {
	p.Mu.RLock()
	defer p.Mu.RUnlock()
	elapsed := time.Since(p.StartTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(p.ProcessedBytes) / elapsed
}

// GetETA returns estimated time remaining in seconds
func (p *ProgressInfo) GetETA() float64 {
	p.Mu.RLock()
	defer p.Mu.RUnlock()
	if p.ProcessedBytes == 0 {
		return 0
	}
	speed := p.GetSpeed()
	if speed == 0 {
		return 0
	}
	remaining := p.TotalBytes - p.ProcessedBytes
	return float64(remaining) / speed
}

// startProgress initializes progress tracking
func (m *Manager) startProgress(op Operation, totalFiles int, totalBytes int64) {
	m.progress.Mu.Lock()
	defer m.progress.Mu.Unlock()
	m.progress.Operation = op
	m.progress.TotalFiles = totalFiles
	m.progress.TotalBytes = totalBytes
	m.progress.ProcessedFiles = 0
	m.progress.ProcessedBytes = 0
	m.progress.CurrentFile = ""
	m.progress.StartTime = time.Now()
	m.progress.Active = true
}

// updateProgress updates the current progress
func (m *Manager) updateProgress(processedBytes int64, currentFile string) {
	m.progress.Mu.Lock()
	defer m.progress.Mu.Unlock()
	m.progress.ProcessedBytes = processedBytes
	m.progress.CurrentFile = currentFile
}

// finishProgress marks the operation as complete
func (m *Manager) finishProgress() {
	m.progress.Mu.Lock()
	defer m.progress.Mu.Unlock()
	m.progress.Active = false
}

// calculateTotalSize calculates total size of files to be processed
func (m *Manager) calculateTotalSize(files []string) (int64, error) {
	var total int64
	for _, path := range files {
		size, err := m.getPathSize(path)
		if err != nil {
			return 0, err
		}
		total += size
	}
	return total, nil
}

// getPathSize returns the total size of a file or directory
func (m *Manager) getPathSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	
	if !info.IsDir() {
		return info.Size(), nil
	}
	
	var total int64
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	
	return total, err
}

// ToggleSelection toggles selection for a file
func (m *Manager) ToggleSelection(path string) {
	if m.selectedFiles[path] {
		delete(m.selectedFiles, path)
	} else {
		m.selectedFiles[path] = true
	}
}

// IsSelected checks if a file is selected
func (m *Manager) IsSelected(path string) bool {
	return m.selectedFiles[path]
}

// ClearSelection clears all selections
func (m *Manager) ClearSelection() {
	m.selectedFiles = make(map[string]bool)
}

// GetSelectedFiles returns list of selected files
func (m *Manager) GetSelectedFiles() []string {
	files := make([]string, 0, len(m.selectedFiles))
	for path := range m.selectedFiles {
		files = append(files, path)
	}
	return files
}

// GetSelectedCount returns number of selected files
func (m *Manager) GetSelectedCount() int {
	return len(m.selectedFiles)
}

// Copy copies selected files to clipboard
func (m *Manager) Copy(files []string) {
	m.clipboard = make([]string, len(files))
	copy(m.clipboard, files)
	m.operation = OpCopy
}

// Cut cuts selected files to clipboard
func (m *Manager) Cut(files []string) {
	m.clipboard = make([]string, len(files))
	copy(m.clipboard, files)
	m.operation = OpCut
}

// Paste pastes files from clipboard to destination
func (m *Manager) Paste(destDir string) error {
	if len(m.clipboard) == 0 {
		return fmt.Errorf("clipboard is empty")
	}

	// Calculate total size for progress tracking
	totalSize, err := m.calculateTotalSize(m.clipboard)
	if err != nil {
		return fmt.Errorf("failed to calculate total size: %v", err)
	}

	// Start progress tracking
	m.startProgress(m.operation, len(m.clipboard), totalSize)
	defer m.finishProgress()

	var processedBytes int64

	for _, srcPath := range m.clipboard {
		fileName := filepath.Base(srcPath)
		destPath := filepath.Join(destDir, fileName)

		// Handle name conflicts
		destPath = m.getUniqueDestPath(destPath)

		if m.operation == OpCopy {
			if err := m.copyFileOrDirWithProgress(srcPath, destPath, &processedBytes); err != nil {
				return fmt.Errorf("failed to copy %s: %v", srcPath, err)
			}
		} else if m.operation == OpCut {
			m.updateProgress(processedBytes, fileName)
			if err := os.Rename(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to move %s: %v", srcPath, err)
			}
			// For move operations, add the file size to processed bytes
			size, _ := m.getPathSize(srcPath)
			processedBytes += size
		}
		
		m.progress.Mu.Lock()
		m.progress.ProcessedFiles++
		m.progress.Mu.Unlock()
	}

	// Clear clipboard after cut operation
	if m.operation == OpCut {
		m.clipboard = make([]string, 0)
		m.operation = OpNone
	}

	return nil
}

// Delete deletes specified files
func (m *Manager) Delete(files []string) error {
	// Calculate total size for progress tracking
	totalSize, err := m.calculateTotalSize(files)
	if err != nil {
		return fmt.Errorf("failed to calculate total size: %v", err)
	}

	// Start progress tracking
	m.startProgress(OpDelete, len(files), totalSize)
	defer m.finishProgress()

	var processedBytes int64

	for _, path := range files {
		fileName := filepath.Base(path)
		m.updateProgress(processedBytes, fileName)
		
		// Get size before deleting
		size, _ := m.getPathSize(path)
		
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to delete %s: %v", path, err)
		}
		
		processedBytes += size
		m.progress.Mu.Lock()
		m.progress.ProcessedFiles++
		m.progress.Mu.Unlock()
	}
	return nil
}

// Rename renames a file
func (m *Manager) Rename(oldPath, newName string) error {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
	
	if oldPath == newPath {
		return nil // No change
	}
	
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file already exists: %s", newName)
	}
	
	return os.Rename(oldPath, newPath)
}

// CreateFile creates a new empty file
func (m *Manager) CreateFile(dir, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	filePath := filepath.Join(dir, filename)
	
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists: %s", filename)
	}
	
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	
	return nil
}

// CreateFolder creates a new directory
func (m *Manager) CreateFolder(dir, foldername string) error {
	if foldername == "" {
		return fmt.Errorf("folder name cannot be empty")
	}
	
	folderPath := filepath.Join(dir, foldername)
	
	if _, err := os.Stat(folderPath); err == nil {
		return fmt.Errorf("folder already exists: %s", foldername)
	}
	
	err := os.Mkdir(folderPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create folder: %v", err)
	}
	
	return nil
}

// GetClipboardInfo returns clipboard status
func (m *Manager) GetClipboardInfo() (count int, op Operation) {
	return len(m.clipboard), m.operation
}

// HasClipboard checks if clipboard has files
func (m *Manager) HasClipboard() bool {
	return len(m.clipboard) > 0
}

// copyFileOrDir copies a file or directory recursively
func (m *Manager) copyFileOrDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return m.copyDir(src, dst)
	}
	return m.copyFile(src, dst)
}

// copyFileOrDirWithProgress copies a file or directory recursively with progress tracking
func (m *Manager) copyFileOrDirWithProgress(src, dst string, processedBytes *int64) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return m.copyDirWithProgress(src, dst, processedBytes)
	}
	return m.copyFileWithProgress(src, dst, processedBytes)
}

// copyFile copies a single file
func (m *Manager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// copyFileWithProgress copies a single file with progress tracking
func (m *Manager) copyFileWithProgress(src, dst string, processedBytes *int64) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Update progress with current file
	m.updateProgress(*processedBytes, filepath.Base(src))

	// Copy with progress tracking
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := dstFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			*processedBytes += int64(n)
			m.updateProgress(*processedBytes, filepath.Base(src))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// copyDir copies a directory recursively
func (m *Manager) copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := m.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := m.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyDirWithProgress copies a directory recursively with progress tracking
func (m *Manager) copyDirWithProgress(src, dst string, processedBytes *int64) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := m.copyDirWithProgress(srcPath, dstPath, processedBytes); err != nil {
				return err
			}
		} else {
			if err := m.copyFileWithProgress(srcPath, dstPath, processedBytes); err != nil {
				return err
			}
		}
	}

	return nil
}

// getUniqueDestPath generates a unique destination path if file exists
func (m *Manager) getUniqueDestPath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	nameWithoutExt := path[:len(path)-len(ext)]
	
	counter := 1
	for {
		newPath := fmt.Sprintf("%s_copy%d%s", nameWithoutExt, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// Made with Bob
