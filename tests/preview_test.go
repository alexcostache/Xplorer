package tests

import (
	"testing"
	"github.com/alexcostache/Xplorer/internal/preview"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"go file", "main.go", "go"},
		{"python file", "script.py", "python"},
		{"javascript", "app.js", "javascript"},
		{"typescript", "app.ts", "typescript"},
		{"json", "config.json", "json"},
		{"shell", "script.sh", "shell"},
		{"html", "index.html", "html"},
		{"css", "style.css", "css"},
		{"c file", "main.c", "c"},
		{"cpp file", "main.cpp", "cpp"},
		{"java", "Main.java", "java"},
		{"ruby", "app.rb", "ruby"},
		{"rust", "main.rs", "rust"},
		{"php", "index.php", "php"},
		{"unknown", "file.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preview.DetectLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"ascii", 'a', 1},
		{"space", ' ', 1},
		{"number", '5', 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preview.RuneWidth(tt.r)
			if got != tt.want {
				t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func BenchmarkDetectLanguage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		preview.DetectLanguage("main.go")
	}
}

// Made with Bob
