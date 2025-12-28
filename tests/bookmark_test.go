package tests

import (
	"testing"
	"github.com/alexcostache/Xplorer/internal/bookmark"
)

func TestBookmarkOperations(t *testing.T) {
	// Create a new bookmark manager
	m := bookmark.NewManager()
	initialCount := m.Count()
	
	testPath := "/home/user/test_bookmark_ops"
	
	// Clean up any existing bookmark
	if m.IsBookmarked(testPath) {
		m.Toggle(testPath)
	}
	
	// Test adding bookmark
	added := m.Toggle(testPath)
	if !added {
		t.Error("Toggle should return true when adding")
	}
	if !m.IsBookmarked(testPath) {
		t.Error("Path should be bookmarked after toggle")
	}
	if m.Count() != initialCount+1 {
		t.Errorf("Expected %d bookmarks, got %d", initialCount+1, m.Count())
	}
	
	// Test removing bookmark
	removed := m.Toggle(testPath)
	if removed {
		t.Error("Toggle should return false when removing")
	}
	if m.IsBookmarked(testPath) {
		t.Error("Path should not be bookmarked after second toggle")
	}
	if m.Count() != initialCount {
		t.Errorf("Expected %d bookmarks, got %d", initialCount, m.Count())
	}
}

func TestBookmarkWithTrailingSlash(t *testing.T) {
	m := bookmark.NewManager()
	testPath := "/home/user/test_trailing"
	
	// Clean up
	if m.IsBookmarked(testPath) {
		m.Toggle(testPath)
	}
	
	m.Toggle(testPath)
	
	if !m.IsBookmarked(testPath + "/") {
		t.Error("Should match path with trailing slash")
	}
	
	// Clean up
	m.Toggle(testPath)
}

func TestGetPath(t *testing.T) {
	m := bookmark.NewManager()
	
	// Test invalid index
	if path := m.GetPath(999); path != "" {
		t.Errorf("Expected empty string for invalid index, got %s", path)
	}
	
	// Test valid index if bookmarks exist
	if m.Count() > 0 {
		path := m.GetPath(0)
		if path == "" {
			t.Error("Expected non-empty path for valid index")
		}
	}
}

func TestRemove(t *testing.T) {
	m := bookmark.NewManager()
	
	testPath := "/test_remove_path"
	
	// Add a bookmark
	if !m.IsBookmarked(testPath) {
		m.Toggle(testPath)
	}
	
	// Find its index
	var idx int = -1
	for i := 0; i < m.Count(); i++ {
		if m.GetPath(i) == testPath {
			idx = i
			break
		}
	}
	
	if idx >= 0 {
		m.Remove(idx)
		if m.IsBookmarked(testPath) {
			t.Error("Bookmark should be removed")
		}
	}
}

// Made with Bob
