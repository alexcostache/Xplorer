package filesystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SortMode represents different file sorting modes
type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
	SortByModTime
	SortByExtension
)

// SortModeNames maps sort modes to their display names
var SortModeNames = map[SortMode]string{
	SortByName:      "Alphabetical",
	SortBySize:      "Size",
	SortByModTime:   "Modified Time",
	SortByExtension: "Type",
}

// Navigator handles file system navigation
type Navigator struct {
	currentDir   string
	fileList     []os.FileInfo
	cursor       int
	scrollOffset int
	filter       string
	showHidden   bool
	sortMode     SortMode
	sortReverse  bool
	history      []string
	historyIndex int
}

// NewNavigator creates a new filesystem navigator
func NewNavigator() *Navigator {
	currentDir, _ := os.Getwd()
	nav := &Navigator{
		currentDir:   currentDir,
		cursor:       0,
		scrollOffset: 0,
		filter:       "",
		showHidden:   false,
		sortMode:     SortByName,
		sortReverse:  false,
		history:      []string{currentDir},
		historyIndex: 0,
	}
	nav.RefreshFileList()
	return nav
}

// GetCurrentDir returns the current directory
func (n *Navigator) GetCurrentDir() string {
	return n.currentDir
}

// SetCurrentDir sets the current directory
func (n *Navigator) SetCurrentDir(dir string) {
	n.currentDir = dir
	n.cursor = 0
	n.scrollOffset = 0
	n.RefreshFileList()
}

// GetFileList returns the current file list
func (n *Navigator) GetFileList() []os.FileInfo {
	return n.fileList
}

// GetCursor returns the current cursor position
func (n *Navigator) GetCursor() int {
	return n.cursor
}

// SetCursor sets the cursor position
func (n *Navigator) SetCursor(pos int) {
	if pos >= 0 && pos < len(n.fileList) {
		n.cursor = pos
	}
}

// GetScrollOffset returns the scroll offset
func (n *Navigator) GetScrollOffset() int {
	return n.scrollOffset
}

// SetScrollOffset sets the scroll offset
func (n *Navigator) SetScrollOffset(offset int) {
	n.scrollOffset = offset
}

// GetFilter returns the current filter
func (n *Navigator) GetFilter() string {
	return n.filter
}

// SetFilter sets the filter and refreshes the file list
func (n *Navigator) SetFilter(filter string) {
	n.filter = filter
	n.cursor = 0
	n.RefreshFileList()
}

// ClearFilter clears the filter
func (n *Navigator) ClearFilter() {
	n.filter = ""
	n.cursor = 0
	n.scrollOffset = 0
}

// ToggleHidden toggles showing hidden files
func (n *Navigator) ToggleHidden() {
	n.showHidden = !n.showHidden
	n.cursor = 0
	n.RefreshFileList()
}

// GetShowHidden returns whether hidden files are shown
func (n *Navigator) GetShowHidden() bool {
	return n.showHidden
}

// SetSortMode sets the sorting mode and toggles reverse if same mode
func (n *Navigator) SetSortMode(mode SortMode) {
	if n.sortMode == mode {
		// Toggle reverse if selecting the same mode
		n.sortReverse = !n.sortReverse
	} else {
		// New mode, reset reverse
		n.sortMode = mode
		n.sortReverse = false
	}
	n.RefreshFileList()
}

// GetSortMode returns the current sorting mode
func (n *Navigator) GetSortMode() SortMode {
	return n.sortMode
}

// GetSortReverse returns whether sorting is reversed
func (n *Navigator) GetSortReverse() bool {
	return n.sortReverse
}

// GetSortModeName returns the display name of the current sort mode
func (n *Navigator) GetSortModeName() string {
	name := SortModeNames[n.sortMode]
	if n.sortReverse {
		name += " ↓"
	}
	return name
}

// RefreshFileList refreshes the file list based on current directory and filter
func (n *Navigator) RefreshFileList() {
	entries, err := ioutil.ReadDir(n.currentDir)
	if err != nil {
		n.fileList = nil
		return
	}
	
	n.fileList = nil
	for _, file := range entries {
		name := file.Name()
		
		// Skip hidden files if not showing them
		if !n.showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		
		// Apply filter
		if n.filter == "" || strings.Contains(strings.ToLower(name), strings.ToLower(n.filter)) {
			n.fileList = append(n.fileList, file)
		}
	}
	
	// Sort based on current sort mode
	n.sortFileList()
	
	// Adjust cursor if out of bounds
	if n.cursor >= len(n.fileList) {
		n.cursor = 0
	}
}

// sortFileList sorts the file list based on the current sort mode
func (n *Navigator) sortFileList() {
	switch n.sortMode {
	case SortByName:
		sort.Slice(n.fileList, func(i, j int) bool {
			// Directories first, then alphabetically
			if n.fileList[i].IsDir() != n.fileList[j].IsDir() {
				return n.fileList[i].IsDir()
			}
			result := strings.ToLower(n.fileList[i].Name()) < strings.ToLower(n.fileList[j].Name())
			if n.sortReverse {
				return !result
			}
			return result
		})
	case SortBySize:
		sort.Slice(n.fileList, func(i, j int) bool {
			// Directories first, then by size
			if n.fileList[i].IsDir() != n.fileList[j].IsDir() {
				return n.fileList[i].IsDir()
			}
			result := n.fileList[i].Size() > n.fileList[j].Size()
			if n.sortReverse {
				return !result
			}
			return result
		})
	case SortByModTime:
		sort.Slice(n.fileList, func(i, j int) bool {
			// Directories first, then by modification time
			if n.fileList[i].IsDir() != n.fileList[j].IsDir() {
				return n.fileList[i].IsDir()
			}
			result := n.fileList[i].ModTime().After(n.fileList[j].ModTime())
			if n.sortReverse {
				return !result
			}
			return result
		})
	case SortByExtension:
		sort.Slice(n.fileList, func(i, j int) bool {
			// Directories first, then by extension
			if n.fileList[i].IsDir() != n.fileList[j].IsDir() {
				return n.fileList[i].IsDir()
			}
			extI := strings.ToLower(filepath.Ext(n.fileList[i].Name()))
			extJ := strings.ToLower(filepath.Ext(n.fileList[j].Name()))
			var result bool
			if extI != extJ {
				result = extI < extJ
			} else {
				result = strings.ToLower(n.fileList[i].Name()) < strings.ToLower(n.fileList[j].Name())
			}
			if n.sortReverse {
				return !result
			}
			return result
		})
	}
}

// MoveUp moves the cursor up
func (n *Navigator) MoveUp(visibleLines int) {
	if n.cursor > 0 {
		n.cursor--
		if n.cursor < n.scrollOffset {
			n.scrollOffset--
		}
	} else if len(n.fileList) > 0 {
		// Wrap to bottom
		n.cursor = len(n.fileList) - 1
		n.scrollOffset = max(0, n.cursor-visibleLines+1)
	}
}

// MoveDown moves the cursor down
func (n *Navigator) MoveDown(visibleLines int) {
	if n.cursor < len(n.fileList)-1 {
		n.cursor++
		if n.cursor >= n.scrollOffset+visibleLines {
			n.scrollOffset++
		}
	} else if len(n.fileList) > 0 {
		// Wrap to top
		n.cursor = 0
		n.scrollOffset = 0
	}
}

// MoveUpFast moves the cursor up by 5 lines (Page Up)
func (n *Navigator) MoveUpFast(visibleLines int) {
	if len(n.fileList) == 0 {
		return
	}
	
	// Move up by 5 lines
	n.cursor -= 5
	if n.cursor < 0 {
		n.cursor = 0
	}
	
	// Adjust scroll offset
	if n.cursor < n.scrollOffset {
		n.scrollOffset = n.cursor
	}
}

// MoveDownFast moves the cursor down by 5 lines (Page Down)
func (n *Navigator) MoveDownFast(visibleLines int) {
	if len(n.fileList) == 0 {
		return
	}
	
	// Move down by 5 lines
	n.cursor += 5
	if n.cursor >= len(n.fileList) {
		n.cursor = len(n.fileList) - 1
	}
	
	// Adjust scroll offset
	if n.cursor >= n.scrollOffset+visibleLines {
		n.scrollOffset = n.cursor - visibleLines + 1
	}
}

// GoToParent navigates to the parent directory
func (n *Navigator) GoToParent() bool {
	parent := filepath.Dir(n.currentDir)
	if parent != n.currentDir {
		n.currentDir = parent
		n.ClearFilter()
		n.historyIndex++
		n.history = append(n.history[:n.historyIndex], n.currentDir)
		n.RefreshFileList()
		return true
	}
	return false
}

// EnterDirectory enters the selected directory
func (n *Navigator) EnterDirectory() bool {
	if len(n.fileList) > 0 {
		selected := n.fileList[n.cursor]
		if selected.IsDir() {
			n.currentDir = filepath.Join(n.currentDir, selected.Name())
			n.ClearFilter()
			n.historyIndex++
			n.history = append(n.history[:n.historyIndex], n.currentDir)
			n.RefreshFileList()
			return true
		}
	}
	return false
}

// GetSelectedPath returns the full path of the selected file
func (n *Navigator) GetSelectedPath() string {
	if len(n.fileList) > 0 && n.cursor < len(n.fileList) {
		return filepath.Join(n.currentDir, n.fileList[n.cursor].Name())
	}
	return ""
}

// GetSelectedFile returns the selected file info
func (n *Navigator) GetSelectedFile() os.FileInfo {
	if len(n.fileList) > 0 && n.cursor < len(n.fileList) {
		return n.fileList[n.cursor]
	}
	return nil
}

// GetParentDir returns the parent directory
func (n *Navigator) GetParentDir() string {
	return filepath.Dir(n.currentDir)
}

// GetParentEntries returns filtered entries from the parent directory
func (n *Navigator) GetParentEntries() []os.FileInfo {
	parent := n.GetParentDir()
	entries, err := ioutil.ReadDir(parent)
	if err != nil {
		return nil
	}
	
	var filtered []os.FileInfo
	for _, f := range entries {
		if !n.showHidden && strings.HasPrefix(f.Name(), ".") {
			continue
		}
		filtered = append(filtered, f)
	}
	
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name() < filtered[j].Name()
	})
	
	return filtered
}

// MoveCursorToBestMatch moves cursor to the best matching file
func (n *Navigator) MoveCursorToBestMatch(visibleLines int) {
	if len(n.fileList) == 0 {
		n.cursor = 0
		n.scrollOffset = 0
		return
	}
	
	n.cursor = 0
	lowerFilter := strings.ToLower(n.filter)
	
	// Find first file matching filter
	for i, file := range n.fileList {
		name := strings.ToLower(file.Name())
		if strings.Contains(name, lowerFilter) {
			n.cursor = i
			break
		}
	}
	
	// Adjust scroll offset
	if n.cursor >= n.scrollOffset+visibleLines {
		n.scrollOffset = n.cursor - visibleLines + 1
	} else if n.cursor < n.scrollOffset {
		n.scrollOffset = n.cursor
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Made with Bob

// Refresh refreshes the file list (alias for RefreshFileList)
func (n *Navigator) Refresh() {
	n.RefreshFileList()
}
