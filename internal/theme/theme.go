package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/nsf/termbox-go"
)

// Theme represents a color scheme
type Theme struct {
	Name               string
	ColorText          termbox.Attribute
	ColorBackground    termbox.Attribute
	ColorHighlight     termbox.Attribute
	ColorHighlightText termbox.Attribute
	ColorFooter        termbox.Attribute
	ColorFooterBg      termbox.Attribute
	ColorAddressBar    termbox.Attribute
	ColorAddressBarBg  termbox.Attribute
	ColorSeparator     termbox.Attribute
	ColorDim           termbox.Attribute
	ColorFilter        termbox.Attribute
	ColorFilterBg      termbox.Attribute
	FileColors         map[string]termbox.Attribute
	DirColor           termbox.Attribute
}

// ThemeJSON represents the JSON structure for themes
type ThemeJSON struct {
	Name       string            `json:"name"`
	Colors     map[string]string `json:"colors"`
	FileColors map[string]string `json:"file_colors,omitempty"`
}

// Manager handles theme operations
type Manager struct {
	themes       []Theme
	current      *Theme
	fileColorMap map[string]termbox.Attribute
}

// NewManager creates a new theme manager
func NewManager() *Manager {
	m := &Manager{
		fileColorMap: getDefaultFileColors(),
	}
	
	// Load themes from JSON files
	m.themes = m.loadThemesFromJSON()
	
	// If no themes loaded, use default
	if len(m.themes) == 0 {
		m.themes = []Theme{getDefaultTheme()}
	}
	
	return m
}

// GetCurrent returns the current theme
func (m *Manager) GetCurrent() *Theme {
	if m.current == nil {
		m.current = &m.themes[0]
	}
	return m.current
}

// GetThemes returns all available themes
func (m *Manager) GetThemes() []Theme {
	return m.themes
}

// SetThemeByName sets the theme by name
func (m *Manager) SetThemeByName(name string) bool {
	for i := range m.themes {
		if m.themes[i].Name == name {
			m.current = &m.themes[i]
			m.saveThemeName(name)
			return true
		}
	}
	return false
}

// LoadSavedTheme loads the previously saved theme
func (m *Manager) LoadSavedTheme() {
	name := m.loadThemeName()
	if name != "" && !m.SetThemeByName(name) {
		m.current = &m.themes[0]
	}
}

// GetFileColor returns the color for a file
func (m *Manager) GetFileColor(name string, isDir bool) termbox.Attribute {
	if isDir {
		return m.GetCurrent().DirColor
	}
	ext := strings.ToLower(filepath.Ext(name))
	// Check theme-specific file colors first
	if color, ok := m.GetCurrent().FileColors[ext]; ok {
		return color
	}
	// Fall back to default text color
	return m.GetCurrent().ColorText
}

// loadThemesFromJSON loads all theme JSON files from the themes directory
func (m *Manager) loadThemesFromJSON() []Theme {
	var themes []Theme
	
	// Get themes directory path
	themesDir := "themes"
	
	// Read all JSON files in themes directory
	files, err := os.ReadDir(themesDir)
	if err != nil {
		return themes
	}
	
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		themePath := filepath.Join(themesDir, file.Name())
		theme, err := m.loadThemeFromFile(themePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load theme %s: %v\n", file.Name(), err)
			continue
		}
		
		themes = append(themes, theme)
	}
	
	return themes
}

// loadThemeFromFile loads a single theme from a JSON file
func (m *Manager) loadThemeFromFile(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, err
	}
	
	var themeJSON ThemeJSON
	if err := json.Unmarshal(data, &themeJSON); err != nil {
		return Theme{}, err
	}
	
	theme := Theme{
		Name:       themeJSON.Name,
		FileColors: make(map[string]termbox.Attribute),
	}
	
	// Parse colors
	theme.ColorText = parseColor(themeJSON.Colors["text"])
	theme.ColorBackground = parseColor(themeJSON.Colors["background"])
	theme.ColorHighlight = parseColor(themeJSON.Colors["highlight"])
	theme.ColorHighlightText = parseColor(themeJSON.Colors["highlight_text"])
	theme.ColorFooter = parseColor(themeJSON.Colors["footer"])
	theme.ColorFooterBg = parseColor(themeJSON.Colors["footer_bg"])
	theme.ColorAddressBar = parseColor(themeJSON.Colors["address_bar"])
	theme.ColorAddressBarBg = parseColor(themeJSON.Colors["address_bar_bg"])
	theme.ColorSeparator = parseColor(themeJSON.Colors["separator"])
	theme.ColorDim = parseColor(themeJSON.Colors["dim"])
	theme.ColorFilter = parseColor(themeJSON.Colors["filter"])
	theme.ColorFilterBg = parseColor(themeJSON.Colors["filter_bg"])
	theme.DirColor = parseColor(themeJSON.Colors["dir"])
	
	// Parse file colors if provided, otherwise use defaults
	if len(themeJSON.FileColors) > 0 {
		for ext, colorName := range themeJSON.FileColors {
			theme.FileColors[ext] = parseColor(colorName)
		}
	} else {
		// Use default file colors
		theme.FileColors = getDefaultFileColors()
	}
	
	// Validate: ensure text and background are different
	if theme.ColorText == theme.ColorBackground {
		theme.ColorText = termbox.ColorWhite
		if theme.ColorBackground == termbox.ColorWhite {
			theme.ColorText = termbox.ColorBlack
		}
	}
	
	// Validate: ensure footer text and background are different
	if theme.ColorFooter == theme.ColorFooterBg {
		theme.ColorFooter = termbox.ColorWhite
		if theme.ColorFooterBg == termbox.ColorWhite {
			theme.ColorFooter = termbox.ColorBlack
		}
	}
	
	// Validate: ensure address bar text and background are different
	if theme.ColorAddressBar == theme.ColorAddressBarBg {
		theme.ColorAddressBar = termbox.ColorWhite
		if theme.ColorAddressBarBg == termbox.ColorWhite {
			theme.ColorAddressBar = termbox.ColorBlack
		}
	}
	
	// Validate: ensure filter text and background are different
	if theme.ColorFilter == theme.ColorFilterBg {
		theme.ColorFilter = termbox.ColorWhite
		if theme.ColorFilterBg == termbox.ColorWhite {
			theme.ColorFilter = termbox.ColorBlack
		}
	}
	
	// Validate: ensure highlight text and background are different
	if theme.ColorHighlightText == theme.ColorHighlight {
		theme.ColorHighlightText = termbox.ColorWhite
		if theme.ColorHighlight == termbox.ColorWhite {
			theme.ColorHighlightText = termbox.ColorBlack
		}
	}
	
	return theme, nil
}

// parseColor converts a color name string to termbox.Attribute
func parseColor(colorName string) termbox.Attribute {
	colorMap := map[string]termbox.Attribute{
		"default":        termbox.ColorDefault,
		"black":          termbox.ColorBlack,
		"red":            termbox.ColorRed,
		"green":          termbox.ColorGreen,
		"yellow":         termbox.ColorYellow,
		"blue":           termbox.ColorBlue,
		"magenta":        termbox.ColorMagenta,
		"cyan":           termbox.ColorCyan,
		"white":          termbox.ColorWhite,
		"bright_black":   termbox.ColorBlack | termbox.AttrBold,
		"bright_red":     termbox.ColorRed | termbox.AttrBold,
		"bright_green":   termbox.ColorGreen | termbox.AttrBold,
		"bright_yellow":  termbox.ColorYellow | termbox.AttrBold,
		"bright_blue":    termbox.ColorBlue | termbox.AttrBold,
		"bright_magenta": termbox.ColorMagenta | termbox.AttrBold,
		"bright_cyan":    termbox.ColorCyan | termbox.AttrBold,
		"bright_white":   termbox.ColorWhite | termbox.AttrBold,
	}
	
	if color, ok := colorMap[strings.ToLower(colorName)]; ok {
		return color
	}
	return termbox.ColorDefault
}

// getThemeConfigFile returns the path to the theme config file
func (m *Manager) getThemeConfigFile() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".xp_theme")
}

// saveThemeName saves the theme name to disk
func (m *Manager) saveThemeName(name string) {
	if name != "" {
		_ = os.WriteFile(m.getThemeConfigFile(), []byte(strings.TrimSpace(name)), 0644)
	}
}

// loadThemeName loads the theme name from disk
func (m *Manager) loadThemeName() string {
	path := m.getThemeConfigFile()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// getDefaultTheme returns a fallback default theme
func getDefaultTheme() Theme {
	return Theme{
		Name:               "Default",
		ColorText:          termbox.ColorWhite,
		ColorBackground:    termbox.ColorBlack,
		ColorHighlight:     termbox.ColorMagenta,
		ColorHighlightText: termbox.ColorWhite,
		ColorFooter:        termbox.ColorCyan,
		ColorFooterBg:      termbox.ColorBlack,
		ColorAddressBar:    termbox.ColorMagenta,
		ColorAddressBarBg:  termbox.ColorBlack,
		ColorSeparator:     termbox.ColorMagenta,
		ColorDim:           termbox.ColorWhite,
		ColorFilter:        termbox.ColorWhite,
		ColorFilterBg:      termbox.ColorMagenta,
		DirColor:           termbox.ColorCyan,
	}
}

// getDefaultFileColors returns the default file color mapping
func getDefaultFileColors() map[string]termbox.Attribute {
	return map[string]termbox.Attribute{
		".go":   termbox.ColorYellow,
		".py":   termbox.ColorBlue,
		".js":   termbox.ColorGreen,
		".ts":   termbox.ColorGreen,
		".sh":   termbox.ColorMagenta,
		".bash": termbox.ColorMagenta,
		".zsh":  termbox.ColorMagenta,
		".json": termbox.ColorWhite,
		".html": termbox.ColorRed,
		".css":  termbox.ColorRed,
		".md":   termbox.ColorCyan,
		".txt":  termbox.ColorWhite,
		".log":  termbox.ColorCyan,
		".zip":  termbox.ColorRed,
		".tar":  termbox.ColorRed,
		".gz":   termbox.ColorRed,
		".rar":  termbox.ColorRed,
		".png":  termbox.ColorMagenta,
		".jpg":  termbox.ColorMagenta,
		".jpeg": termbox.ColorMagenta,
		".svg":  termbox.ColorMagenta,
		".gif":  termbox.ColorMagenta,
		".mp3":  termbox.ColorRed,
		".wav":  termbox.ColorRed,
		".flac": termbox.ColorRed,
		".mp4":  termbox.ColorRed,
		".mkv":  termbox.ColorRed,
		".webm": termbox.ColorRed,
		".pdf":  termbox.ColorRed,
		".exe":  termbox.ColorRed,
		".dll":  termbox.ColorRed,
		".bin":  termbox.ColorWhite,
		".dat":  termbox.ColorWhite,
		".yml":  termbox.ColorWhite,
		".yaml": termbox.ColorWhite,
		".toml": termbox.ColorWhite,
		".c":    termbox.ColorRed,
		".h":    termbox.ColorRed,
		".cpp":  termbox.ColorRed,
		".java": termbox.ColorRed,
	}
}

// SaveTheme saves a theme to a JSON file
func (m *Manager) SaveTheme(theme *Theme) error {
	themesDir := "themes"
	
	// Ensure themes directory exists
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return err
	}
	
	// Create theme JSON
	themeJSON := ThemeJSON{
		Name:       theme.Name,
		Colors:     make(map[string]string),
		FileColors: make(map[string]string),
	}
	
	// Convert colors to strings
	themeJSON.Colors["text"] = colorToString(theme.ColorText)
	themeJSON.Colors["background"] = colorToString(theme.ColorBackground)
	themeJSON.Colors["highlight"] = colorToString(theme.ColorHighlight)
	themeJSON.Colors["highlight_text"] = colorToString(theme.ColorHighlightText)
	themeJSON.Colors["footer"] = colorToString(theme.ColorFooter)
	themeJSON.Colors["footer_bg"] = colorToString(theme.ColorFooterBg)
	themeJSON.Colors["address_bar"] = colorToString(theme.ColorAddressBar)
	themeJSON.Colors["address_bar_bg"] = colorToString(theme.ColorAddressBarBg)
	themeJSON.Colors["separator"] = colorToString(theme.ColorSeparator)
	themeJSON.Colors["dim"] = colorToString(theme.ColorDim)
	themeJSON.Colors["filter"] = colorToString(theme.ColorFilter)
	themeJSON.Colors["filter_bg"] = colorToString(theme.ColorFilterBg)
	themeJSON.Colors["dir"] = colorToString(theme.DirColor)
	
	// Convert file colors
	for ext, color := range theme.FileColors {
		themeJSON.FileColors[ext] = colorToString(color)
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(themeJSON, "", "  ")
	if err != nil {
		return err
	}
	
	// Save to file
	filename := strings.ToLower(strings.ReplaceAll(theme.Name, " ", "-")) + ".json"
	filepath := filepath.Join(themesDir, filename)
	
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return err
	}
	
	// Reload themes to include the new one
	m.themes = m.loadThemesFromJSON()
	
	// Set the new theme as current
	m.SetThemeByName(theme.Name)
	
	return nil
}

// UpdateThemeColor updates a specific color in the current theme and saves it
func (m *Manager) UpdateThemeColor(element, colorName string) {
	if m.current == nil {
		return
	}
	
	m.UpdateThemeColorPreview(element, colorName)
	
	// Save the modified theme
	m.SaveTheme(m.current)
}

// UpdateThemeColorPreview updates a specific color in the current theme without saving (for preview)
func (m *Manager) UpdateThemeColorPreview(element, colorName string) {
	if m.current == nil {
		return
	}
	
	color := parseColor(colorName)
	
	switch element {
	case "Text Color":
		m.current.ColorText = color
	case "Background Color":
		m.current.ColorBackground = color
	case "Highlight Color":
		m.current.ColorHighlight = color
	case "Highlight Text Color":
		m.current.ColorHighlightText = color
	case "Footer Color":
		m.current.ColorFooter = color
	case "Footer Background":
		m.current.ColorFooterBg = color
	case "Address Bar Color":
		m.current.ColorAddressBar = color
	case "Address Bar Background":
		m.current.ColorAddressBarBg = color
	case "Separator Color":
		m.current.ColorSeparator = color
	case "Dim Color":
		m.current.ColorDim = color
	case "Filter Color":
		m.current.ColorFilter = color
	case "Filter Background":
		m.current.ColorFilterBg = color
	case "Directory Color":
		m.current.DirColor = color
	}
}

// RestoreDefaultTheme restores the default theme
func (m *Manager) RestoreDefaultTheme() {
	defaultTheme := getDefaultTheme()
	m.current = &defaultTheme
	m.saveThemeName(defaultTheme.Name)
}

// DeleteTheme deletes a theme file
func (m *Manager) DeleteTheme(themeName string) error {
	// Don't allow deleting the current theme or default theme
	if m.current != nil && m.current.Name == themeName {
		return fmt.Errorf("cannot delete the currently active theme")
	}
	
	if themeName == "Default" {
		return fmt.Errorf("cannot delete the default theme")
	}
	
	// Find and delete the theme file
	themesDir := "themes"
	filename := strings.ToLower(strings.ReplaceAll(themeName, " ", "-")) + ".json"
	filepath := filepath.Join(themesDir, filename)
	
	if err := os.Remove(filepath); err != nil {
		return err
	}
	
	// Reload themes
	m.themes = m.loadThemesFromJSON()
	
	return nil
}

// RenameTheme renames a theme
func (m *Manager) RenameTheme(oldName, newName string) error {
	if oldName == "Default" {
		return fmt.Errorf("cannot rename the default theme")
	}
	
	if newName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	
	// Check if new name already exists
	for _, t := range m.themes {
		if t.Name == newName {
			return fmt.Errorf("theme '%s' already exists", newName)
		}
	}
	
	// Find the theme
	var themeToRename *Theme
	for i := range m.themes {
		if m.themes[i].Name == oldName {
			themeToRename = &m.themes[i]
			break
		}
	}
	
	if themeToRename == nil {
		return fmt.Errorf("theme '%s' not found", oldName)
	}
	
	// Delete old file
	themesDir := "themes"
	oldFilename := strings.ToLower(strings.ReplaceAll(oldName, " ", "-")) + ".json"
	oldFilepath := filepath.Join(themesDir, oldFilename)
	
	// Update theme name
	themeToRename.Name = newName
	
	// Save with new name
	if err := m.SaveTheme(themeToRename); err != nil {
		return err
	}
	
	// Delete old file
	os.Remove(oldFilepath)
	
	// If this was the current theme, update the saved theme name
	if m.current != nil && m.current.Name == newName {
		m.saveThemeName(newName)
	}
	
	return nil
}

// colorToString converts a termbox.Attribute to a color name string
func colorToString(color termbox.Attribute) string {
	// Check for bright colors (with bold attribute)
	if color&termbox.AttrBold != 0 {
		baseColor := color &^ termbox.AttrBold
		brightColorMap := map[termbox.Attribute]string{
			termbox.ColorBlack:   "bright_black",
			termbox.ColorRed:     "bright_red",
			termbox.ColorGreen:   "bright_green",
			termbox.ColorYellow:  "bright_yellow",
			termbox.ColorBlue:    "bright_blue",
			termbox.ColorMagenta: "bright_magenta",
			termbox.ColorCyan:    "bright_cyan",
			termbox.ColorWhite:   "bright_white",
		}
		if name, ok := brightColorMap[baseColor]; ok {
			return name
		}
	}
	
	// Regular colors
	colorMap := map[termbox.Attribute]string{
		termbox.ColorDefault: "default",
		termbox.ColorBlack:   "black",
		termbox.ColorRed:     "red",
		termbox.ColorGreen:   "green",
		termbox.ColorYellow:  "yellow",
		termbox.ColorBlue:    "blue",
		termbox.ColorMagenta: "magenta",
		termbox.ColorCyan:    "cyan",
		termbox.ColorWhite:   "white",
	}
	
	if name, ok := colorMap[color]; ok {
		return name
	}
	return "default"
}

// Made with Bob
