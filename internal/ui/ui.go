package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexcostache/Xplorer/internal/bookmark"
	"github.com/alexcostache/Xplorer/internal/config"
	"github.com/alexcostache/Xplorer/internal/fileops"
	"github.com/alexcostache/Xplorer/internal/filesystem"
	"github.com/alexcostache/Xplorer/internal/preview"
	"github.com/alexcostache/Xplorer/internal/theme"

	"github.com/nsf/termbox-go"
	"golang.org/x/text/width"
)

// IconSpacing defines the space between icon and filename
// Adjust this value to change spacing globally (e.g., " ", "  ", or "")
const IconSpacing = " "

// debugLog writes debug messages to /tmp/xp_debug.log
func debugLog(format string, args ...interface{}) {
	f, err := os.OpenFile("/tmp/xp_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf(format, args...)
}

// Renderer handles all UI rendering
type Renderer struct {
	themeManager    *theme.Manager
	bookmarkManager *bookmark.Manager
	previewManager  *preview.Manager
	config          *config.Config
	fileOpsManager  *fileops.Manager
}

// NewRenderer creates a new UI renderer
func NewRenderer(tm *theme.Manager, bm *bookmark.Manager, pm *preview.Manager, cfg *config.Config, fom *fileops.Manager) *Renderer {
	return &Renderer{
		themeManager:    tm,
		bookmarkManager: bm,
		previewManager:  pm,
		config:          cfg,
		fileOpsManager:  fom,
	}
}

// Draw renders the entire UI
func (r *Renderer) Draw(nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) {
	termbox.Clear(r.theme().ColorBackground, r.theme().ColorBackground)
	w, h := termbox.Size()

	// Define panel widths and positions with consistent spacing
	// Layout: [Parent Panel] | [Middle Panel] | [Preview Panel]
	parentPanelWidth := w / 5                    // 20% for parent
	middlePanelWidth := (w * 2) / 5              // 40% for middle
	
	// Calculate positions
	parentPanelStart := 0
	separator1Pos := parentPanelWidth
	middlePanelStart := separator1Pos + 1
	separator2Pos := middlePanelStart + middlePanelWidth
	previewPanelStart := separator2Pos + 1

	// Draw address bar
	r.drawAddressBar(nav.GetCurrentDir(), inPathEditMode, pathEditBuffer)

	// Draw left panel (parent directory)
	r.drawParentPanel(nav, parentPanelStart, parentPanelWidth, h)

	// Draw middle panel (current directory)
	r.drawCurrentPanel(nav, middlePanelStart, middlePanelWidth, h)

	// Draw right panel (preview)
	r.drawPreviewPanel(nav, previewPanelStart, w, h)

	// Draw vertical separators
	for y := 1; y < h-1; y++ {
		termbox.SetCell(separator1Pos, y, '│', r.theme().ColorSeparator, r.theme().ColorBackground)
		termbox.SetCell(separator2Pos, y, '│', r.theme().ColorSeparator, r.theme().ColorBackground)
	}

	// Draw filter bar
	if filter := nav.GetFilter(); filter != "" {
		r.drawFilterBar(filter, w, h)
	}

	// Draw metadata bar
	r.drawMetadataBar(nav, w, h)

	// Draw help panel if active
	if showHelp {
		r.drawHelpPanel()
	}

	// NOTE: Don't flush here - let caller decide when to flush
	// This allows progress bar to be drawn as an overlay
}

// DrawAndFlush renders the UI and flushes to screen
func (r *Renderer) DrawAndFlush(nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) {
	r.Draw(nav, inPathEditMode, pathEditBuffer, showHelp)
	termbox.Flush()
}
// drawTextInBox draws text in a box with proper Unicode support
func drawTextInBox(startX, y, maxWidth int, text string, fg, bg termbox.Attribute) {
	runes := []rune(text)
	if len(runes) > maxWidth {
		runes = runes[:maxWidth]
	}
	
	x := 0
	for _, r := range runes {
		termbox.SetCell(startX+x, y, r, fg, bg)
		x++
	}
	// Fill remaining space
	for x < maxWidth {
		termbox.SetCell(startX+x, y, ' ', fg, bg)
		x++
	}
}


// drawAddressBar draws the address/path bar at the top
func (r *Renderer) drawAddressBar(path string, inPathEditMode bool, pathEditBuffer string) {
	w, _ := termbox.Size()

	if inPathEditMode {
		text := "Path: " + pathEditBuffer
		for i := 0; i < w; i++ {
			termbox.SetCell(i, 0, ' ', r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		for i, rn := range text {
			if i >= w {
				break
			}
			termbox.SetCell(i, 0, rn, r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		return
	}

	// Raw path view
	if r.config.ShowRawPath {
		text := path
		// Replace home directory with ~ for cleaner display
		usr, _ := user.Current()
		if usr != nil && usr.HomeDir != "" {
			if strings.HasPrefix(text, usr.HomeDir) {
				text = strings.Replace(text, usr.HomeDir, "~", 1)
			}
		}
		for i := 0; i < w; i++ {
			termbox.SetCell(i, 0, ' ', r.theme().ColorAddressBar, r.theme().ColorAddressBarBg)
		}
		for i, rn := range text {
			if i >= w {
				break
			}
			termbox.SetCell(i, 0, rn, r.theme().ColorAddressBar, r.theme().ColorAddressBarBg)
		}
		return
	}

	// Breadcrumb view
	usr, _ := user.Current()
	home := usr.HomeDir
	if strings.HasPrefix(path, home) {
		path = strings.Replace(path, home, "~", 1)
	}

	parts := strings.Split(filepath.Clean(path), string(os.PathSeparator))
	if parts[0] == "" {
		parts[0] = string(os.PathSeparator)
	}

	x := 0
	for i, part := range parts {
		bg := r.theme().ColorAddressBarBg
		fg := r.theme().ColorAddressBar
		if i == len(parts)-1 {
			bg = r.theme().ColorHighlight
			fg = r.theme().ColorHighlightText
		}

		text := part
		if i > 0 {
			text = " › " + part
		}
		for _, rn := range text {
			if x >= w {
				break
			}
			termbox.SetCell(x, 0, rn, fg, bg)
			x += runeWidth(rn)
		}
	}
	for ; x < w; x++ {
		termbox.SetCell(x, 0, ' ', r.theme().ColorAddressBar, r.theme().ColorAddressBarBg)
	}
}

// drawParentPanel draws the left panel showing parent directory
func (r *Renderer) drawParentPanel(nav *filesystem.Navigator, startX, width, height int) {
	parentEntries := nav.GetParentEntries()
	currentBase := filepath.Base(nav.GetCurrentDir())

	y := 2
	for _, f := range parentEntries {
		name := f.Name()
		icon := config.FileIcon(name, f.IsDir(), r.config.UseAsciiIcons)
		color := r.themeManager.GetFileColor(name, f.IsDir())
		fullPath := filepath.Join(nav.GetParentDir(), name)
		
		displayName := name
		if r.bookmarkManager.IsBookmarked(fullPath) {
			displayName += " ★"
		}
		line := formatFileLine(icon, displayName)

		isActiveFolder := (name == currentBase)
		bgColor := r.theme().ColorBackground
		textColor := color
		if isActiveFolder {
			bgColor = r.theme().ColorHighlight
			textColor = r.theme().ColorHighlightText
		}

		// Fill background
		for i := 0; i < width; i++ {
			termbox.SetCell(startX+i, y, ' ', r.theme().ColorText, bgColor)
		}
		
		// Add padding when icons are disabled
		x := startX
		if !r.config.UseAsciiIcons {
			x = startX + 1
		}
		for _, rn := range line {
			if x >= startX+width {
				break
			}
			termbox.SetCell(x, y, rn, textColor, bgColor)
			x += runeWidth(rn)
		}
		
		y++
		if y >= height-2 {
			break
		}
	}
}

// drawCurrentPanel draws the middle panel showing current directory
func (r *Renderer) drawCurrentPanel(nav *filesystem.Navigator, startX, width, height int) {
	fileList := nav.GetFileList()
	cursor := nav.GetCursor()
	scrollOffset := nav.GetScrollOffset()
	visibleHeight := height - 4
	sizeColumnWidth := 12 // Width for size column (e.g., "1.23 MB")

	for i := scrollOffset; i < len(fileList) && i < scrollOffset+visibleHeight; i++ {
		y := (i - scrollOffset) + 2
		file := fileList[i]
		icon := config.FileIcon(file.Name(), file.IsDir(), r.config.UseAsciiIcons)
		color := r.themeManager.GetFileColor(file.Name(), file.IsDir())
		fullPath := filepath.Join(nav.GetCurrentDir(), file.Name())
		
		displayName := file.Name()
		if r.bookmarkManager.IsBookmarked(fullPath) {
			displayName += " ★"
		}
		
		line := formatFileLine(icon, displayName)
		
		// Get file size
		var sizeStr string
		if file.IsDir() {
			sizeStr = "<DIR>"
		} else {
			sizeStr = formatSize(file.Size())
		}

		// Determine if file is selected
		isSelected := r.fileOpsManager.IsSelected(fullPath)
		
		// Add selection marker to line if selected
		if isSelected {
			line = "✓ " + line
		}
		
		// Draw background
		for x := 0; x < width; x++ {
			bg := r.theme().ColorBackground
			if i == cursor {
				bg = r.theme().ColorHighlight
			}
			termbox.SetCell(startX+x, y, ' ', r.theme().ColorText, bg)
		}

		// Draw filename
		fg := color
		bg := r.theme().ColorBackground
		if i == cursor {
			fg = r.theme().ColorHighlightText
			bg = r.theme().ColorHighlight
		} else if isSelected {
			// Selected files use highlight color for text (no background change)
			fg = r.theme().ColorHighlight
		}
		
		// Add padding when icons are disabled
		x := startX
		if !r.config.UseAsciiIcons {
			x = startX + 1
		}
		maxNameWidth := width - sizeColumnWidth - 1
		if !r.config.UseAsciiIcons {
			maxNameWidth--
		}
		charCount := 0
		for _, rn := range line {
			if charCount >= maxNameWidth {
				break
			}
			termbox.SetCell(x, y, rn, fg, bg)
			w := runeWidth(rn)
			x += w
			charCount += w
		}
		
		// Draw size column (right-aligned) - same color as filename
		sizeX := startX + width - len(sizeStr)
		for j, rn := range sizeStr {
			termbox.SetCell(sizeX+j, y, rn, fg, bg)
		}
	}
}

// drawPreviewPanel draws the right panel showing file/directory preview
func (r *Renderer) drawPreviewPanel(nav *filesystem.Navigator, startX, width, height int) {
	fileList := nav.GetFileList()
	if len(fileList) == 0 {
		return
	}

	cursor := nav.GetCursor()
	selected := filepath.Join(nav.GetCurrentDir(), fileList[cursor].Name())
	info, err := os.Stat(selected)
	if err != nil {
		return
	}

	if info.IsDir() {
		// Directory preview
		entries, _ := os.ReadDir(selected)
		lineNum := 0
		for _, entry := range entries {
			if !nav.GetShowHidden() && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			icon := config.FileIcon(entry.Name(), entry.IsDir(), r.config.UseAsciiIcons)
			color := r.themeManager.GetFileColor(entry.Name(), entry.IsDir())
			text := formatFileLine(icon, entry.Name())
			
			// Add padding when icons are disabled
			x := startX
			if !r.config.UseAsciiIcons {
				x = startX + 1
			}
			for _, rn := range text {
				if x >= width {
					break
				}
				termbox.SetCell(x, lineNum+2, rn, color, r.theme().ColorBackground)
				x += runeWidth(rn)
			}
			lineNum++
			if lineNum >= height-4 {
				break
			}
		}
	} else {
		// File preview with syntax highlighting
		lines := r.previewManager.GetLines()
		if lines != nil {
			visibleHeight := height - 4
			scrollOffset := r.previewManager.GetScrollOffset()
			start := scrollOffset
			end := start + visibleHeight
			if end > len(lines) {
				end = len(lines)
			}
			
			lang := preview.DetectLanguage(fileList[cursor].Name())
			for i := start; i < end; i++ {
				y := (i - start) + 2
				preview.DrawText(startX+1, y, lines[i], lang, r.theme().ColorText, r.theme().ColorBackground, r.theme().ColorDim)
			}
		}
	}
}

// drawFilterBar draws the filter input bar
func (r *Renderer) drawFilterBar(filter string, width, height int) {
	filterText := "Filter: " + filter
	for i := 0; i < width; i++ {
		termbox.SetCell(i, height-2, ' ', r.theme().ColorFilter, r.theme().ColorFilterBg)
	}
	for i, rn := range filterText {
		if i >= width {
			break
		}
		termbox.SetCell(i, height-2, rn, r.theme().ColorFilter, r.theme().ColorFilterBg)
	}
}

// drawMetadataBar draws the bottom status bar
func (r *Renderer) drawMetadataBar(nav *filesystem.Navigator, width, height int) {
	fileList := nav.GetFileList()
	if len(fileList) == 0 {
		return
	}

	cursor := nav.GetCursor()
	info := fileList[cursor]
	name := info.Name()
	size := formatSize(info.Size())
	mode := info.Mode()
	modTime := info.ModTime().Format("2006-01-02 15:04:05")

	// Count items
	parentCount := len(nav.GetParentEntries())
	currentCount := len(fileList)
	
	previewCount := 0
	selected := filepath.Join(nav.GetCurrentDir(), fileList[cursor].Name())
	if info.IsDir() {
		entries, _ := os.ReadDir(selected)
		for _, e := range entries {
			if !nav.GetShowHidden() && strings.HasPrefix(e.Name(), ".") {
				continue
			}
			previewCount++
		}
	} else {
		previewCount = len(r.previewManager.GetLines())
	}

	selectedCount := r.fileOpsManager.GetSelectedCount()
	selectionInfo := ""
	if selectedCount > 0 {
		selectionInfo = fmt.Sprintf(" | Selected: %d", selectedCount)
	}
	left := fmt.Sprintf(" %s | %s | %s | %s%s", name, size, mode, modTime, selectionInfo)
	right := fmt.Sprintf("▲ %d ◀ %d ▶ %d | Hidden: %s | Sort: %s", parentCount, currentCount, previewCount, boolStr(nav.GetShowHidden()), nav.GetSortModeName())

	for i := 0; i < width; i++ {
		termbox.SetCell(i, height-1, ' ', r.theme().ColorFooter, r.theme().ColorFooterBg)
	}
	for i, rn := range left {
		if i >= width {
			break
		}
		termbox.SetCell(i, height-1, rn, r.theme().ColorFooter, r.theme().ColorFooterBg)
	}
	startX := width - len(right)
	if startX > len(left)+2 {
		for i, rn := range right {
			if startX+i >= width {
				break
			}
			termbox.SetCell(startX+i, height-1, rn, r.theme().ColorFooter, r.theme().ColorFooterBg)
		}
	}
}

// drawHelpPanel draws the help overlay
func (r *Renderer) drawHelpPanel() {
	w, h := termbox.Size()
	keys := r.config.Keys

	help := []string{
		"↑↓       Navigate",
		"PgUp/Dn  Navigate fast (5 lines)",
		"←→       Enter/Back Dir",
		"Enter    Open with... (select editor)",
		"Space    Select/Deselect file",
		"Ctrl+O   File operations menu",
		"Ctrl+S   Change sorting mode",
		fmt.Sprintf("%c        Filter", keys.Filter),
		fmt.Sprintf("%c        Themes", keys.OpenThemePopup),
		fmt.Sprintf("%c        Configuration Menu", keys.ConfigMenu),
		fmt.Sprintf("%c        Toggle Hidden", keys.ToggleHidden),
		fmt.Sprintf("%c        Open in Terminal", keys.OpenTerminal),
		fmt.Sprintf("%c        Quit", keys.Quit),
		fmt.Sprintf("%c        Toggle Help", keys.Help),
		fmt.Sprintf("%c        Bookmark current folder", keys.BookmarkToggle),
		fmt.Sprintf("%c        Jump to a bookmark", keys.BookmarkPopup),
		fmt.Sprintf("%c        Edit path", keys.EditPath),
		fmt.Sprintf("%c        Scroll preview ↓", keys.ScrollDown),
		fmt.Sprintf("%c        Scroll preview ↑", keys.ScrollUp),
		fmt.Sprintf("%c        Scroll preview ↓ (fast)", keys.ScrollDownFast),
		fmt.Sprintf("%c        Scroll preview ↑ (fast)", keys.ScrollUpFast),
		fmt.Sprintf("%c        Toggle path display", keys.TogglePath),
	}

	boxWidth := 50
	boxHeight := len(help) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2

	DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Help", r.theme().ColorFooter, r.theme().ColorFooterBg)

	for i, line := range help {
		for j, ch := range line {
			if startX+2+j >= startX+boxWidth-2 {
				break
			}
			termbox.SetCell(startX+2+j, startY+2+i, ch, r.theme().ColorFooter, r.theme().ColorFooterBg)
		}
	}
}

// ShowThemeSelector shows the theme selection with full window preview
func (r *Renderer) ShowThemeSelector(nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) {
	w, h := termbox.Size()
	themes := r.themeManager.GetThemes()
	boxWidth := 40
	boxHeight := len(themes) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2

	selectedIndex := -1
	currentTheme := r.themeManager.GetCurrent()
	for i, t := range themes {
		if t.Name == currentTheme.Name {
			selectedIndex = i
			break
		}
	}
	if selectedIndex == -1 {
		selectedIndex = 0
	}
	originalThemeName := currentTheme.Name

	r.themeManager.SetThemeByName(themes[selectedIndex].Name)

	for {
		// Draw the full UI with the current theme
		r.Draw(nav, inPathEditMode, pathEditBuffer, showHelp)
		
		// Draw the theme selector box on top
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Themes", r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		for j, ch := range "[Themes] ↑↓, Enter to confirm, Esc to cancel" {
			if startX+2+j < startX+boxWidth-2 {
				termbox.SetCell(startX+2+j, startY, ch, r.theme().ColorFooter, r.theme().ColorFooterBg)
			}
		}

		for i, t := range themes {
			name := t.Name
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			if i == selectedIndex {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			for j, ch := range name {
				if startX+2+j < startX+boxWidth-2 {
					termbox.SetCell(startX+2+j, startY+2+i, ch, fg, bg)
				}
			}
		}

		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				if selectedIndex > 0 {
					selectedIndex--
				} else {
					selectedIndex = len(themes) - 1
				}
				r.themeManager.SetThemeByName(themes[selectedIndex].Name)
			case termbox.KeyArrowDown:
				if selectedIndex < len(themes)-1 {
					selectedIndex++
				} else {
					selectedIndex = 0
				}
				r.themeManager.SetThemeByName(themes[selectedIndex].Name)
			case termbox.KeyEnter:
				return
			case termbox.KeyEsc:
				r.themeManager.SetThemeByName(originalThemeName)
				return
			}
		}
	}
}

// ShowBookmarkPopup shows the bookmark selection popup
func (r *Renderer) ShowBookmarkPopup() string {
	w, h := termbox.Size()
	bookmarks := r.bookmarkManager.GetAll()
	boxWidth := 50
	boxHeight := len(bookmarks) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2

	index := 0
	for {
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Bookmarks", r.theme().ColorFooter, r.theme().ColorFooterBg)

		for i, b := range bookmarks {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == index {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			text := " " + b.Name
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}

		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				index--
				if index < 0 {
					index = len(bookmarks) - 1
				}
			case termbox.KeyArrowDown:
				index++
				if index >= len(bookmarks) {
					index = 0
				}
			case termbox.KeyEnter:
				return r.bookmarkManager.GetPath(index)
			case termbox.KeyEsc:
				return ""
			}
		}
	}
}

// Prompt shows an input prompt (for filter - updates file list)
func (r *Renderer) Prompt(label string, nav *filesystem.Navigator) string {
	w, h := termbox.Size()
	input := ""

	for {
		nav.SetFilter(input)
		nav.MoveCursorToBestMatch(h - 4)
		r.Draw(nav, false, "", false)

		full := label + input
		for i := 0; i < w; i++ {
			termbox.SetCell(i, h-2, ' ', r.theme().ColorFilter, r.theme().ColorFilterBg)
		}
		for i, rn := range full {
			if i >= w {
				break
			}
			termbox.SetCell(i, h-2, rn, r.theme().ColorFilter, r.theme().ColorFilterBg)
		}
		termbox.Flush()

		e := termbox.PollEvent()
		if e.Type == termbox.EventKey {
			switch e.Key {
			case termbox.KeyEnter:
				return input
			case termbox.KeyEsc:
				nav.SetFilter("")
				r.Draw(nav, false, "", false)
				return ""
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			default:
				if e.Ch != 0 {
					input += string(e.Ch)
				}
			}
		}
	}
}

// SimplePrompt shows a simple input prompt without filtering (allows spaces)
func (r *Renderer) SimplePrompt(label string, nav *filesystem.Navigator) string {
	w, h := termbox.Size()
	input := ""

	for {
		// Draw current UI without modifying it
		r.Draw(nav, false, "", false)

		full := label + input
		for i := 0; i < w; i++ {
			termbox.SetCell(i, h-2, ' ', r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		for i, rn := range full {
			if i >= w {
				break
			}
			termbox.SetCell(i, h-2, rn, r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		termbox.Flush()

		e := termbox.PollEvent()
		if e.Type == termbox.EventKey {
			switch e.Key {
			case termbox.KeyEnter:
				return input
			case termbox.KeyEsc:
				return ""
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			case termbox.KeySpace:
				input += " "
			default:
				if e.Ch != 0 {
					input += string(e.Ch)
				}
			}
		}
	}
}

// ConfirmPrompt shows a yes/no confirmation prompt
func (r *Renderer) ConfirmPrompt(message string) bool {
	w, h := termbox.Size()
	prompt := message + " (y/n)"
	
	for {
		for i := 0; i < w; i++ {
			termbox.SetCell(i, h-2, ' ', r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		for i, rn := range prompt {
			if i >= w {
				break
			}
			termbox.SetCell(i, h-2, rn, r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Ch {
			case 'y', 'Y':
				return true
			case 'n', 'N':
				return false
			}
			if ev.Key == termbox.KeyEsc {
				return false
			}
		}
	}
}

// OpenTerminal opens a terminal in the given directory
func OpenTerminal(path, terminalApp string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("cmd", "/C", "start", "cmd", "/K", "cd", "/d", path).Start()
	case "darwin":
		exec.Command("open", "-a", terminalApp, path).Start()
	default:
		exec.Command(terminalApp, "--working-directory="+path).Start()
	}
}

// Helper functions

func (r *Renderer) theme() *theme.Theme {
	return r.themeManager.GetCurrent()
}

func formatFileLine(icon, name string) string {
	if icon == "" {
		return name
	}
	return IconSpacing + icon + IconSpacing + name
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func boolStr(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

func runeWidth(r rune) int {
	prop := width.LookupRune(r)
	switch prop.Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	default:
		return 1
	}
}

// DrawBoxWithTitle draws a box with a centered title
func DrawBoxWithTitle(startX, startY, width, height int, title string, fg, bg termbox.Attribute) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ch := ' '
			switch {
			case y == 0 && x == 0:
				ch = '╔'
			case y == 0 && x == width-1:
				ch = '╗'
			case y == height-1 && x == 0:
				ch = '╚'
			case y == height-1 && x == width-1:
				ch = '╝'
			case y == 0 || y == height-1:
				ch = '═'
			case x == 0 || x == width-1:
				ch = '║'
			}
			termbox.SetCell(startX+x, startY+y, ch, fg, bg)
		}
	}

	// Draw centered title
	title = " " + title + " "
	titleStartX := startX + (width-len(title))/2
	for i, r := range title {
		if titleStartX+i >= startX && titleStartX+i < startX+width {
			termbox.SetCell(titleStartX+i, startY, r, fg, bg)
		}
	}
}

// ShowEditorSelectionPopup displays a popup to select an editor
func (r *Renderer) ShowEditorSelectionPopup(editors []config.EditorOption, nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) int {
	w, h := termbox.Size()
	popupWidth := 60
	popupHeight := len(editors) + 4
	startX := (w - popupWidth) / 2
	startY := (h - popupHeight) / 2

	selected := 0

	for {
		termbox.Clear(r.theme().ColorText, r.theme().ColorBackground)
		r.Draw(nav, inPathEditMode, pathEditBuffer, showHelp)

		// Draw popup box
		DrawBoxWithTitle(startX, startY, popupWidth, popupHeight, "Open With", r.theme().ColorText, r.theme().ColorBackground)

		// Draw editor options
		for i, editor := range editors {
			y := startY + 2 + i
			fg := r.theme().ColorText
			bg := r.theme().ColorBackground

			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}

			// Format: "Name - Description"
			text := fmt.Sprintf(" %s - %s", editor.Name, editor.Description)
			drawTextInBox(startX+1, y, popupWidth-2, text, fg, bg)
		}

		termbox.Flush()

		ev := termbox.PollEvent()
		
		// Handle window focus events - redraw on any event type
		if ev.Type == termbox.EventResize || ev.Type == termbox.EventInterrupt {
			continue // Redraw and continue
		}
		
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(editors) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(editors) {
					selected = 0
				}
			case termbox.KeyEnter:
				return selected
			case termbox.KeyEsc:
				return -1
			}
		}
	}
}

// Made with Bob

// ShowContextMenu displays a context menu for file operations
func (r *Renderer) ShowContextMenu(options []string, nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) int {
	w, h := termbox.Size()
	popupWidth := 40
	popupHeight := len(options) + 4
	startX := (w - popupWidth) / 2
	startY := (h - popupHeight) / 2

	selected := 0

	for {
		termbox.Clear(r.theme().ColorText, r.theme().ColorBackground)
		r.Draw(nav, inPathEditMode, pathEditBuffer, showHelp)

		// Draw popup box
		DrawBoxWithTitle(startX, startY, popupWidth, popupHeight, "File Operations", r.theme().ColorText, r.theme().ColorBackground)

		// Draw menu options
		for i, option := range options {
			y := startY + 2 + i
			fg := r.theme().ColorText
			bg := r.theme().ColorBackground

			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}

			text := " " + option
			drawTextInBox(startX+1, y, popupWidth-2, text, fg, bg)
		}

		termbox.Flush()
		debugLog("ShowSortingPopup: Waiting for event...")

		ev := termbox.PollEvent()
		debugLog("ShowSortingPopup: Got event type=%d key=%d ch=%c", ev.Type, ev.Key, ev.Ch)
		
		// Handle window focus events - redraw on any event type
		if ev.Type == termbox.EventResize || ev.Type == termbox.EventInterrupt {
			debugLog("ShowSortingPopup: Resize/Interrupt event, continuing")
			continue // Redraw and continue
		}
		
		if ev.Type == termbox.EventKey {
			debugLog("ShowSortingPopup: Key event")
			switch ev.Key {
			case termbox.KeyArrowUp:
				debugLog("ShowSortingPopup: Arrow Up")
				selected--
				if selected < 0 {
					selected = len(options) - 1 // Wrap to bottom
				}
			case termbox.KeyArrowDown:
				debugLog("ShowSortingPopup: Arrow Down")
				selected++
				if selected >= len(options) {
					selected = 0 // Wrap to top
				}
			case termbox.KeyEnter:
				debugLog("ShowSortingPopup: Enter pressed, returning %d", selected)
				return selected
			case termbox.KeyEsc:
				debugLog("ShowSortingPopup: ESC pressed, returning -1")
				return -1
			}
		}
	}
}
// ShowSortingPopup displays a popup to select sorting mode
func (r *Renderer) ShowSortingPopup(nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) int {
	debugLog("ShowSortingPopup: ENTER")
	w, h := termbox.Size()
	
	// Build sorting options
	options := []string{
		"Alphabetical",
		"Size",
		"Modified Time",
		"Type",
	}
	
	popupWidth := 40
	popupHeight := len(options) + 4
	startX := (w - popupWidth) / 2
	startY := (h - popupHeight) / 2

	// Start with current sort mode selected
	selected := int(nav.GetSortMode())
	debugLog("ShowSortingPopup: Starting event loop")

	for {
		termbox.Clear(r.theme().ColorText, r.theme().ColorBackground)
		r.Draw(nav, inPathEditMode, pathEditBuffer, showHelp)

		// Draw popup box
		DrawBoxWithTitle(startX, startY, popupWidth, popupHeight, "Sort Files By", r.theme().ColorText, r.theme().ColorBackground)

		// Draw menu options
		for i, option := range options {
			y := startY + 2 + i
			fg := r.theme().ColorText
			bg := r.theme().ColorBackground

			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}

			// Add checkmark for current sort mode and reverse indicator
			prefix := "  "
			suffix := ""
			if i == int(nav.GetSortMode()) {
				prefix = "✓ "
				if nav.GetSortReverse() {
					suffix = " ↓"
				}
			}
			text := prefix + option + suffix
			
			// Convert to runes for proper Unicode handling
			runes := []rune(text)
			maxRunes := popupWidth - 4
			if len(runes) > maxRunes {
				runes = runes[:maxRunes]
			}

			// Draw the text with proper Unicode support
			x := 0
			for _, r := range runes {
				termbox.SetCell(startX+1+x, y, r, fg, bg)
				x++
			}
			// Fill remaining space
			for x < popupWidth-2 {
				termbox.SetCell(startX+1+x, y, ' ', fg, bg)
				x++
			}
		}

		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(options) - 1 // Wrap to bottom
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(options) {
					selected = 0 // Wrap to top
				}
			case termbox.KeyEnter:
				return selected
			case termbox.KeyEsc:
				return -1
			}
		}
	}
}


// ShowError displays an error message
func (r *Renderer) ShowError(message string) {
	w, h := termbox.Size()
	
	for i := 0; i < w; i++ {
		termbox.SetCell(i, h-2, ' ', r.theme().ColorHighlightText, r.theme().ColorHighlight)
	}
	
	errorMsg := "Error: " + message
	for i, rn := range errorMsg {
		if i >= w {
			break
		}
		termbox.SetCell(i, h-2, rn, r.theme().ColorHighlightText, r.theme().ColorHighlight)
	}
	termbox.Flush()
	
	// Wait for any key press
	termbox.PollEvent()
}

// ShowConfigMenu displays the main configuration menu
func (r *Renderer) ShowConfigMenu() string {
	w, h := termbox.Size()
	
	// Build options with current state
	mouseStatus := "disabled"
	if r.config.MouseEnabled {
		mouseStatus = "enabled"
	}
	
	iconStatus := "ASCII"
	if !r.config.UseAsciiIcons {
		iconStatus = "Unicode"
	}
	
	options := []string{
		"Select Theme",
		"Create New Theme",
		"Modify Theme Colors",
		"Rename Theme",
		"Delete Theme",
		"Set Default Editor",
		"Toggle Mouse Support [" + mouseStatus + "]",
		"Toggle Icon Style [" + iconStatus + "]",
		"Restore to Default",
		"Cancel",
	}
	
	boxWidth := 50
	boxHeight := len(options) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2
	
	selected := 0
	
	for {
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Configuration Menu", r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		// Draw menu options
		for i, option := range options {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			text := " " + option
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}
		
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(options) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(options) {
					selected = 0
				}
			case termbox.KeyEnter:
				return options[selected]
			case termbox.KeyEsc:
				return "Cancel"
			}
		}
	}
}

// ShowThemeCreator shows the theme creation interface
func (r *Renderer) ShowThemeCreator() bool {
	themeName := r.promptForInput("Enter theme name: ")
	if themeName == "" {
		return false
	}
	
	// Create new theme based on current theme
	newTheme := *r.themeManager.GetCurrent()
	newTheme.Name = themeName
	
	// Save the new theme
	if err := r.themeManager.SaveTheme(&newTheme); err != nil {
		r.ShowError("Failed to create theme: " + err.Error())
		return false
	}
	
	r.ShowMessage("Theme '" + themeName + "' created successfully!")
	return true
}

// ShowThemeColorModifier shows the color modification interface
func (r *Renderer) ShowThemeColorModifier(nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) {
	w, h := termbox.Size()
	
	colorOptions := []string{
		"Text Color",
		"Background Color",
		"Highlight Color",
		"Highlight Text Color",
		"Footer Color",
		"Footer Background",
		"Address Bar Color",
		"Address Bar Background",
		"Separator Color",
		"Dim Color",
		"Filter Color",
		"Filter Background",
		"Directory Color",
		"Done",
	}
	
	boxWidth := 50
	boxHeight := len(colorOptions) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2
	
	selected := 0
	
	for {
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Modify Colors", r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		for i, option := range colorOptions {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			text := " " + option
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}
		
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(colorOptions) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(colorOptions) {
					selected = 0
				}
			case termbox.KeyEnter:
				if colorOptions[selected] == "Done" {
					return
				}
				r.modifyColor(colorOptions[selected], nav, inPathEditMode, pathEditBuffer, showHelp)
			case termbox.KeyEsc:
				return
			}
		}
	}
}

// modifyColor shows color selection for a specific element with live preview
func (r *Renderer) modifyColor(element string, nav *filesystem.Navigator, inPathEditMode bool, pathEditBuffer string, showHelp bool) {
	w, h := termbox.Size()
	
	colors := []string{
		"default",
		"black", "red", "green", "yellow",
		"blue", "magenta", "cyan", "white",
		"bright_black", "bright_red", "bright_green", "bright_yellow",
		"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
	}
	
	boxWidth := 45
	boxHeight := len(colors) + 4
	if boxHeight > h-4 {
		boxHeight = h - 4
	}
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2
	
	selected := 0
	
	// Store original color value to restore on cancel
	originalTheme := *r.themeManager.GetCurrent()
	
	for {
		// Apply the selected color temporarily for preview
		r.themeManager.UpdateThemeColorPreview(element, colors[selected])
		
		// Draw the full UI with the preview
		r.Draw(nav, inPathEditMode, pathEditBuffer, showHelp)
		
		// Draw the color selector box on top
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Select Color for "+element, r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		for i, color := range colors {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			// Show color preview box
			text := " " + color
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}
		
		// Add instruction text
		instruction := "↑↓ Navigate, Enter to confirm, Esc to cancel"
		for i, ch := range instruction {
			if startX+2+i < startX+boxWidth-2 {
				termbox.SetCell(startX+2+i, startY+boxHeight-1, ch, r.theme().ColorFooter, r.theme().ColorFooterBg)
			}
		}
		
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(colors) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(colors) {
					selected = 0
				}
			case termbox.KeyEnter:
				// Save the selected color permanently
				r.themeManager.UpdateThemeColor(element, colors[selected])
				return
			case termbox.KeyEsc:
				// Restore original theme
				*r.themeManager.GetCurrent() = originalTheme
				return
			}
		}
	}
}

// promptForInput shows a simple input prompt
func (r *Renderer) promptForInput(label string) string {
	w, h := termbox.Size()
	input := ""
	
	for {
		for i := 0; i < w; i++ {
			termbox.SetCell(i, h-2, ' ', r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		
		full := label + input
		for i, rn := range full {
			if i >= w {
				break
			}
			termbox.SetCell(i, h-2, rn, r.theme().ColorHighlightText, r.theme().ColorHighlight)
		}
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyEnter:
				return input
			case termbox.KeyEsc:
				return ""
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			case termbox.KeySpace:
				input += " "
			default:
				if ev.Ch != 0 {
					input += string(ev.Ch)
				}
			}
		}
	}
}

// ShowMessage displays a message to the user
func (r *Renderer) ShowMessage(message string) {
	w, h := termbox.Size()
	
	for i := 0; i < w; i++ {
		termbox.SetCell(i, h-2, ' ', r.theme().ColorHighlightText, r.theme().ColorHighlight)
	}
	
	for i, rn := range message {
		if i >= w {
			break
		}
		termbox.SetCell(i, h-2, rn, r.theme().ColorHighlightText, r.theme().ColorHighlight)
	}
	termbox.Flush()
	
	// Wait for any key press
	termbox.PollEvent()
}

// ShowThemeDeleter shows theme deletion interface
func (r *Renderer) ShowThemeDeleter() bool {
	w, h := termbox.Size()
	themes := r.themeManager.GetThemes()
	
	// Filter out default theme and current theme
	var deletableThemes []string
	currentTheme := r.themeManager.GetCurrent()
	for _, t := range themes {
		if t.Name != "Default" && t.Name != currentTheme.Name {
			deletableThemes = append(deletableThemes, t.Name)
		}
	}
	
	if len(deletableThemes) == 0 {
		r.ShowMessage("No themes available to delete")
		return false
	}
	
	boxWidth := 50
	boxHeight := len(deletableThemes) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2
	
	selected := 0
	
	for {
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Delete Theme", r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		for i, themeName := range deletableThemes {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			text := " " + themeName
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}
		
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(deletableThemes) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(deletableThemes) {
					selected = 0
				}
			case termbox.KeyEnter:
				if r.ConfirmPrompt("Delete theme '" + deletableThemes[selected] + "'?") {
					if err := r.themeManager.DeleteTheme(deletableThemes[selected]); err != nil {
						r.ShowError(err.Error())
					} else {
						r.ShowMessage("Theme deleted successfully!")
						return true
					}
				}
				return false
			case termbox.KeyEsc:
				return false
			}
		}
	}
}

// ShowThemeRenamer shows theme renaming interface
func (r *Renderer) ShowThemeRenamer() bool {
	w, h := termbox.Size()
	themes := r.themeManager.GetThemes()
	
	// Filter out default theme
	var renamableThemes []string
	for _, t := range themes {
		if t.Name != "Default" {
			renamableThemes = append(renamableThemes, t.Name)
		}
	}
	
	if len(renamableThemes) == 0 {
		r.ShowMessage("No themes available to rename")
		return false
	}
	
	boxWidth := 50
	boxHeight := len(renamableThemes) + 4
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2
	
	selected := 0
	
	for {
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Rename Theme", r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		for i, themeName := range renamableThemes {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			text := " " + themeName
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}
		
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(renamableThemes) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(renamableThemes) {
					selected = 0
				}
			case termbox.KeyEnter:
				oldName := renamableThemes[selected]
				newName := r.promptForInput("New name for '" + oldName + "': ")
				if newName != "" {
					if err := r.themeManager.RenameTheme(oldName, newName); err != nil {
						r.ShowError(err.Error())
					} else {
						r.ShowMessage("Theme renamed successfully!")
						return true
					}
				}
				return false
			case termbox.KeyEsc:
				return false
			}
		}
	}
}

// ShowDefaultEditorSelector shows editor selection for setting default editor
func (r *Renderer) ShowDefaultEditorSelector() string {
	w, h := termbox.Size()
	
	// Get available editors
	editors := config.GetAvailableEditors()
	
	if len(editors) == 0 {
		r.ShowMessage("No editors found on system")
		return ""
	}
	
	boxWidth := 60
	boxHeight := len(editors) + 4
	if boxHeight > h-4 {
		boxHeight = h - 4
	}
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2
	
	selected := 0
	
	// Find current editor in list
	currentCmd := r.config.EditorCmd
	for i, editor := range editors {
		if editor.Command == currentCmd {
			selected = i
			break
		}
	}
	
	for {
		DrawBoxWithTitle(startX, startY, boxWidth, boxHeight, "Set Default Editor", r.theme().ColorFooter, r.theme().ColorFooterBg)
		
		for i, editor := range editors {
			y := startY + 2 + i
			fg := r.theme().ColorFooter
			bg := r.theme().ColorFooterBg
			
			if i == selected {
				fg = r.theme().ColorHighlightText
				bg = r.theme().ColorHighlight
			}
			
			// Show current editor marker
			marker := "  "
			if editor.Command == currentCmd {
				marker = "✓ "
			}
			
			text := marker + editor.Name + " - " + editor.Description
			drawTextInBox(startX+1, y, boxWidth-2, text, fg, bg)
		}
		
		termbox.Flush()
		
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected--
				if selected < 0 {
					selected = len(editors) - 1
				}
			case termbox.KeyArrowDown:
				selected++
				if selected >= len(editors) {
					selected = 0
				}
			case termbox.KeyEnter:
				return editors[selected].Command
			case termbox.KeyEsc:
				return ""
			}
		}
	}
}

// DrawProgressBar draws a progress bar above the metadata bar
func (r *Renderer) DrawProgressBar(progress *fileops.ProgressInfo) {
	w, h := termbox.Size()
	y := h - 2 // One line above the metadata bar
	
	if progress == nil {
		return
	}
	
	progress.Mu.RLock()
	isActive := progress.Active
	opType := progress.Operation
	currentFile := progress.CurrentFile
	processedFiles := progress.ProcessedFiles
	totalFiles := progress.TotalFiles
	processedBytes := progress.ProcessedBytes
	totalBytes := progress.TotalBytes
	progress.Mu.RUnlock()
	
	// Show progress bar if:
	// 1. Operation is currently active, OR
	// 2. Operation completed and we have data to show
	shouldShow := isActive || (totalFiles > 0)
	
	if !shouldShow {
		return
	}
	
	// Get operation name
	opName := "Processing"
	switch opType {
	case fileops.OpCopy:
		opName = "Copying"
	case fileops.OpCut:
		opName = "Moving"
	case fileops.OpDelete:
		opName = "Deleting"
	}
	
	// If not active, show completion message
	if !isActive && totalFiles > 0 {
		opName = opName[:len(opName)-3] // Remove "ing" suffix
		statusText := fmt.Sprintf("%s completed! (%d files)", opName, totalFiles)
		
		// Draw completion message across the bottom
		for x := 0; x < w; x++ {
			ch := ' '
			if x < len(statusText) {
				ch = rune(statusText[x])
			}
			termbox.SetCell(x, y, ch, r.theme().ColorHighlight, r.theme().ColorHighlightText)
		}
		return
	}
	
	// Calculate progress percentage
	percent := 0
	if totalBytes > 0 {
		percent = int((processedBytes * 100) / totalBytes)
	}
	
	// Format speed
	progress.Mu.RLock()
	speed := progress.GetSpeed()
	progress.Mu.RUnlock()
	speedStr := formatBytes(int64(speed)) + "/s"
	
	// Format current file (truncate if too long)
	maxFileLen := 30
	if len(currentFile) > maxFileLen {
		currentFile = "..." + currentFile[len(currentFile)-maxFileLen+3:]
	}
	
	// Build status text
	statusText := fmt.Sprintf("%s: %s (%d/%d files) %d%% - %s",
		opName, currentFile, processedFiles, totalFiles, percent, speedStr)
	
	// Calculate progress bar width (leave space for text)
	barWidth := w - len(statusText) - 4
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 40 {
		barWidth = 40
	}
	
	// Draw status text
	x := 1
	for _, ch := range statusText {
		if x >= w-barWidth-3 {
			break
		}
		termbox.SetCell(x, y, ch, r.theme().ColorFooter, r.theme().ColorFooterBg)
		x++
	}
	
	// Draw progress bar
	barStart := w - barWidth - 2
	filledWidth := (barWidth * percent) / 100
	
	// Draw bar border
	termbox.SetCell(barStart, y, '[', r.theme().ColorFooter, r.theme().ColorFooterBg)
	termbox.SetCell(barStart+barWidth+1, y, ']', r.theme().ColorFooter, r.theme().ColorFooterBg)
	
	// Draw filled portion
	for i := 0; i < barWidth; i++ {
		ch := ' '
		fg := r.theme().ColorFooter
		bg := r.theme().ColorFooterBg
		
		if i < filledWidth {
			ch = '█'
			fg = r.theme().ColorHighlight
		}
		
		termbox.SetCell(barStart+1+i, y, ch, fg, bg)
	}
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
