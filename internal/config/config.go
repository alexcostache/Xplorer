package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nsf/termbox-go"
)

// Config holds application configuration
type Config struct {
	EditorCmd     string
	TerminalApp   string
	ShowHidden    bool
	ShowRawPath   bool
	MouseEnabled  bool
	UseAsciiIcons bool
	Keys          KeyBindings
}

// EditorOption represents an editor choice
type EditorOption struct {
	Name        string
	Command     string
	IsTerminal  bool
	Description string
}

// ConfigFile represents the JSON config file structure
type ConfigFile struct {
	EditorCmd     string `json:"editor_cmd,omitempty"`
	TerminalApp   string `json:"terminal_app,omitempty"`
	MouseEnabled  *bool  `json:"mouse_enabled,omitempty"`
	UseAsciiIcons *bool  `json:"use_ascii_icons,omitempty"`
}

// KeyBindings holds all keyboard shortcuts
type KeyBindings struct {
	Filter         rune
	ToggleHidden   rune
	Quit           rune
	Help           rune
	OpenTerminal   rune
	BookmarkToggle rune
	BookmarkPopup  rune
	EditPath       rune
	ScrollDown     rune
	ScrollUp       rune
	ScrollDownFast rune
	ScrollUpFast   rune
	OpenThemePopup rune
	TogglePath     rune
	OpenWith       rune
	ConfigMenu     rune
}

// New creates a new configuration with platform-specific defaults
func New() *Config {
	cfg := &Config{
		ShowHidden:    false,
		ShowRawPath:   true,
		MouseEnabled:  true, // Enable mouse by default
		UseAsciiIcons: true, // Enable ASCII icons by default
		Keys:          defaultKeyBindings(),
	}

	// Get platform-specific defaults
	var defaultEditor, defaultTerminal string
	switch runtime.GOOS {
	case "windows":
		defaultEditor = "notepad"
		defaultTerminal = "cmd"
	case "darwin":
		defaultEditor = "nvim"
		defaultTerminal = "iTerm"
	default: // Linux/Unix
		defaultEditor = "vim"
		defaultTerminal = "x-terminal-emulator"
	}

	// Load from config file if exists
	configFile := loadConfigFile()
	
	// Priority: config file > environment variable > platform default
	if configFile.EditorCmd != "" {
		cfg.EditorCmd = configFile.EditorCmd
	} else {
		cfg.EditorCmd = getEnvOrDefault("EDITOR_CMD", defaultEditor)
	}
	
	if configFile.TerminalApp != "" {
		cfg.TerminalApp = configFile.TerminalApp
	} else {
		cfg.TerminalApp = getEnvOrDefault("TERMINAL_APP", defaultTerminal)
	}
	
	if configFile.MouseEnabled != nil {
		cfg.MouseEnabled = *configFile.MouseEnabled
	}
	
	if configFile.UseAsciiIcons != nil {
		cfg.UseAsciiIcons = *configFile.UseAsciiIcons
	}

	return cfg
}

// defaultKeyBindings returns the default key bindings
func defaultKeyBindings() KeyBindings {
	return KeyBindings{
		Filter:         '/',
		ToggleHidden:   '.',
		Quit:           'q',
		Help:           '?',
		OpenTerminal:   't',
		BookmarkToggle: 'B',
		BookmarkPopup:  'b',
		EditPath:       'e',
		ScrollDown:     '[',
		ScrollUp:       ']',
		ScrollDownFast: '{',
		ScrollUpFast:   '}',
		OpenThemePopup: 'T',
		TogglePath:     'r',
		OpenWith:       'o',
		ConfigMenu:     'P',
	}
}

// GetAvailableEditors returns a list of editors that are actually installed on the system
func GetAvailableEditors() []EditorOption {
	allEditors := []EditorOption{
		{Name: "Vim", Command: "vim", IsTerminal: true, Description: "Terminal text editor"},
		{Name: "Neovim", Command: "nvim", IsTerminal: true, Description: "Modern Vim"},
		{Name: "Nano", Command: "nano", IsTerminal: true, Description: "Simple terminal editor"},
		{Name: "Emacs", Command: "emacs -nw", IsTerminal: true, Description: "Terminal Emacs"},
		{Name: "Micro", Command: "micro", IsTerminal: true, Description: "Modern terminal editor"},
		{Name: "Helix", Command: "hx", IsTerminal: true, Description: "Post-modern text editor"},
		{Name: "VS Code", Command: "code", IsTerminal: false, Description: "Visual Studio Code"},
		{Name: "Sublime Text", Command: "subl", IsTerminal: false, Description: "Sublime Text editor"},
		{Name: "Atom", Command: "atom", IsTerminal: false, Description: "Atom editor"},
		{Name: "Notepad++", Command: "notepad++", IsTerminal: false, Description: "Notepad++ (Windows)"},
		{Name: "TextEdit", Command: "open -e", IsTerminal: false, Description: "TextEdit (macOS)"},
		{Name: "Notepad", Command: "notepad", IsTerminal: false, Description: "Notepad (Windows)"},
		{Name: "Gedit", Command: "gedit", IsTerminal: false, Description: "GNOME Text Editor"},
		{Name: "Kate", Command: "kate", IsTerminal: false, Description: "KDE Text Editor"},
		{Name: "Geany", Command: "geany", IsTerminal: false, Description: "Lightweight IDE"},
	}
	
	// Filter to only include installed editors
	var availableEditors []EditorOption
	for _, editor := range allEditors {
		if isEditorInstalled(editor.Command) {
			availableEditors = append(availableEditors, editor)
		}
	}
	
	return availableEditors
}

// GetSystemActions returns system-level actions (terminal, file explorer)
func GetSystemActions() []EditorOption {
	actions := []EditorOption{}
	
	// Add Terminal option
	actions = append(actions, EditorOption{
		Name:        "Terminal",
		Command:     "__TERMINAL__",
		IsTerminal:  false,
		Description: "Open in terminal",
	})
	
	// Add File Explorer option based on OS
	switch runtime.GOOS {
	case "darwin":
		actions = append(actions, EditorOption{
			Name:        "Finder",
			Command:     "__FINDER__",
			IsTerminal:  false,
			Description: "Reveal in Finder",
		})
	case "windows":
		actions = append(actions, EditorOption{
			Name:        "Explorer",
			Command:     "__EXPLORER__",
			IsTerminal:  false,
			Description: "Open in File Explorer",
		})
	default: // Linux
		actions = append(actions, EditorOption{
			Name:        "File Manager",
			Command:     "__FILEMANAGER__",
			IsTerminal:  false,
			Description: "Open in file manager",
		})
	}
	
	return actions
}

// isEditorInstalled checks if an editor command is available on the system
func isEditorInstalled(command string) bool {
	// Extract the base command (first word before any arguments)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	baseCmd := parts[0]
	
	// Special handling for macOS "open -e" command
	if baseCmd == "open" && runtime.GOOS == "darwin" {
		return true
	}
	
	// Use exec.LookPath to check if command exists in PATH
	_, err := exec.LookPath(baseCmd)
	return err == nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getConfigFilePath returns the path to the config file
func getConfigFilePath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".xp_config.json")
}

// loadConfigFile loads configuration from JSON file
func loadConfigFile() ConfigFile {
	var cfg ConfigFile
	
	path := getConfigFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg // Return empty config if file doesn't exist
	}
	
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// SaveConfigFile saves configuration to JSON file
func SaveConfigFile(editorCmd, terminalApp string, mouseEnabled, useAsciiIcons *bool) error {
	cfg := ConfigFile{
		EditorCmd:     editorCmd,
		TerminalApp:   terminalApp,
		MouseEnabled:  mouseEnabled,
		UseAsciiIcons: useAsciiIcons,
	}
	
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(getConfigFilePath(), data, 0644)
}

// GetConfigFilePath returns the config file path (exported for external use)
func GetConfigFilePath() string {
	return getConfigFilePath()
}

// AsciiFileIcon returns an ASCII icon for a file based on its extension
func AsciiFileIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}
	
	ext := getExtension(name)
	
	// Check if it's an image file
	imageExts := []string{".png", ".jpg", ".jpeg", ".svg", ".gif", ".bmp", ".ico", ".webp"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return "🖼"
		}
	}
	
	// All other files use the same icon
	return "📄"
}

// FileIcon returns an icon for a file based on its extension
func FileIcon(name string, isDir bool, useAscii bool) string {
	if useAscii {
		return AsciiFileIcon(name, isDir)
	}
	
	if isDir {
		return ""
	}
	
	ext := getExtension(name)
	icons := map[string]string{
		".go":   "",
		".py":   "",
		".js":   "",
		".ts":   "",
		".json": "",
		".html": "",
		".css":  "",
		".md":   "",
		".sh":   "", ".zsh": "", ".bash": "",
		".c": "", ".h": "", ".cpp": "",
		".java": "",
		".txt":  "", ".log": "",
		".yml":  "", ".yaml": "", ".toml": "",
		".pdf":  "",
		".zip":  "", ".tar": "", ".gz": "", ".rar": "",
		".png":  "", ".jpg": "", ".jpeg": "", ".svg": "", ".gif": "",
		".mp3":  "", ".wav": "", ".flac": "",
		".mp4":  "", ".mkv": "", ".webm": "",
	}
	
	if icon, ok := icons[ext]; ok {
		return icon
	}
	return ""
}

// DescribeFileByExt returns a human-readable description of a file type
func DescribeFileByExt(name string) string {
	ext := getExtension(name)
	
	descriptions := map[string]string{
		".exe":  "EXE File",
		".dll":  "DLL File",
		".png":  "Image File", ".jpg": "Image File", ".jpeg": "Image File", ".gif": "Image File", ".svg": "Image File",
		".zip":  "Archive File", ".tar": "Archive File", ".gz": "Archive File", ".rar": "Archive File",
		".pdf":  "PDF Document",
		".mp4":  "Video File", ".mkv": "Video File", ".avi": "Video File",
		".mp3":  "Audio File", ".wav": "Audio File", ".flac": "Audio File",
		".bin":  "Binary File", ".dat": "Binary File",
	}
	
	if desc, ok := descriptions[ext]; ok {
		if ext != "" {
			return desc + " (" + ext + ")"
		}
		return desc
	}
	
	if ext != "" {
		return "Unknown File (" + ext + ")"
	}
	return "Unknown File"
}

// getExtension returns the lowercase file extension
func getExtension(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return toLower(name[i:])
		}
	}
	return ""
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// ColorAttributes for syntax highlighting
var (
	ColorKeyword termbox.Attribute = termbox.ColorCyan
	ColorString  termbox.Attribute = termbox.ColorYellow
	ColorComment termbox.Attribute = termbox.ColorGreen
)

// Made with Bob
