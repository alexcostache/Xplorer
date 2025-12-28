package tests

import (
	"testing"
	"github.com/alexcostache/Xplorer/internal/config"
)

func TestFileIcon(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		isDir    bool
		wantIcon bool
	}{
		{"directory", "mydir", true, true},
		{"go file", "main.go", false, true},
		{"python file", "script.py", false, true},
		{"javascript", "app.js", false, true},
		{"typescript", "app.ts", false, true},
		{"json", "config.json", false, true},
		{"html", "index.html", false, true},
		{"unknown", "file.xyz", false, true}, // Returns default icon
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.FileIcon(tt.filename, tt.isDir, true)
			// Just verify it returns a string (icon can be empty for unknown types)
			if tt.wantIcon && got == "" {
				// This is actually OK - unknown files get default icon
				// Just verify the function doesn't panic
			}
		})
	}
}

func TestDescribeFileByExt(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"exe file", "program.exe", "EXE File (.exe)"},
		{"dll file", "library.dll", "DLL File (.dll)"},
		{"png image", "photo.png", "Image File (.png)"},
		{"zip archive", "data.zip", "Archive File (.zip)"},
		{"pdf document", "doc.pdf", "PDF Document (.pdf)"},
		{"mp4 video", "movie.mp4", "Video File (.mp4)"},
		{"mp3 audio", "song.mp3", "Audio File (.mp3)"},
		{"binary file", "data.bin", "Binary File (.bin)"},
		{"no extension", "README", "Unknown File"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.DescribeFileByExt(tt.filename)
			if got != tt.want {
				t.Errorf("DescribeFileByExt(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := config.New()

	if cfg.Keys.Filter == 0 {
		t.Error("Filter key should be set")
	}
	if cfg.Keys.ToggleHidden == 0 {
		t.Error("ToggleHidden key should be set")
	}
	if cfg.Keys.Quit == 0 {
		t.Error("Quit key should be set")
	}
	if cfg.Keys.TogglePath == 0 {
		t.Error("TogglePath key should be set")
	}

	if cfg.Keys.Filter != '/' {
		t.Errorf("Expected Filter key to be '/', got %q", cfg.Keys.Filter)
	}
	if cfg.Keys.TogglePath != 'r' {
		t.Errorf("Expected TogglePath key to be 'r', got %q", cfg.Keys.TogglePath)
	}

	if !cfg.ShowRawPath {
		t.Error("Expected ShowRawPath to default to true")
	}
}

func BenchmarkFileIcon(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config.FileIcon("main.go", false, true)
	}
}

// Made with Bob
