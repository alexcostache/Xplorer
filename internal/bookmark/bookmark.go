package bookmark

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

// Bookmark represents a saved directory location
type Bookmark struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Manager handles bookmark operations
type Manager struct {
	bookmarks []Bookmark
}

// NewManager creates a new bookmark manager
func NewManager() *Manager {
	m := &Manager{
		bookmarks: []Bookmark{},
	}
	m.Load()
	return m
}

// GetAll returns all bookmarks
func (m *Manager) GetAll() []Bookmark {
	return m.bookmarks
}

// IsBookmarked checks if a path is bookmarked
func (m *Manager) IsBookmarked(path string) bool {
	cleanPath := filepath.Clean(path)
	for _, b := range m.bookmarks {
		if filepath.Clean(b.Path) == cleanPath {
			return true
		}
	}
	return false
}

// Toggle adds or removes a bookmark
func (m *Manager) Toggle(path string) bool {
	cleanPath := filepath.Clean(path)
	
	// Check if already bookmarked
	for i, b := range m.bookmarks {
		if filepath.Clean(b.Path) == cleanPath {
			// Remove bookmark
			m.bookmarks = append(m.bookmarks[:i], m.bookmarks[i+1:]...)
			m.Save()
			return false // removed
		}
	}
	
	// Add new bookmark
	name := filepath.Base(cleanPath)
	m.bookmarks = append(m.bookmarks, Bookmark{
		Name: name,
		Path: cleanPath,
	})
	m.Save()
	return true // added
}

// Remove removes a bookmark at the given index
func (m *Manager) Remove(index int) {
	if index >= 0 && index < len(m.bookmarks) {
		m.bookmarks = append(m.bookmarks[:index], m.bookmarks[index+1:]...)
		m.Save()
	}
}

// GetPath returns the path of a bookmark at the given index
func (m *Manager) GetPath(index int) string {
	if index >= 0 && index < len(m.bookmarks) {
		return m.bookmarks[index].Path
	}
	return ""
}

// RemoveByPath removes a bookmark by its path
func (m *Manager) RemoveByPath(path string) bool {
	cleanPath := filepath.Clean(path)
	for i, b := range m.bookmarks {
		if filepath.Clean(b.Path) == cleanPath {
			m.bookmarks = append(m.bookmarks[:i], m.bookmarks[i+1:]...)
			m.Save()
			return true
		}
	}
	return false
}

// Count returns the number of bookmarks
func (m *Manager) Count() int {
	return len(m.bookmarks)
}

// getBookmarkFile returns the path to the bookmark file
func (m *Manager) getBookmarkFile() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".xp_bookmarks.json")
}

// Load loads bookmarks from disk
func (m *Manager) Load() {
	path := m.getBookmarkFile()
	data, err := os.ReadFile(path)
	if err != nil {
		return // File doesn't exist yet, that's ok
	}
	_ = json.Unmarshal(data, &m.bookmarks)
}

// Save saves bookmarks to disk
func (m *Manager) Save() {
	data, _ := json.MarshalIndent(m.bookmarks, "", "  ")
	_ = os.WriteFile(m.getBookmarkFile(), data, 0644)
}

// Made with Bob
