package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexcostache/Xplorer/internal/bookmark"
	"github.com/alexcostache/Xplorer/internal/config"
	"github.com/alexcostache/Xplorer/internal/fileops"
	"github.com/alexcostache/Xplorer/internal/filesystem"
	"github.com/alexcostache/Xplorer/internal/preview"
	"github.com/alexcostache/Xplorer/internal/theme"
	"github.com/alexcostache/Xplorer/internal/ui"

	"github.com/nsf/termbox-go"
)

// getDebugLogPath returns the path to the debug log file in the executable's directory
func getDebugLogPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "xp_debug.log" // Fallback to current directory
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "xp_debug.log")
}

// debugLog writes debug messages to xp_debug.log in the app directory (only if debug is enabled)
func (a *App) debugLog(format string, args ...interface{}) {
	if !a.debugEnabled {
		return
	}
	logPath := getDebugLogPath()
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf(format, args...)
}

// EnableDebug enables debug logging
func (a *App) EnableDebug() {
	a.debugEnabled = true
	// Clear previous log file
	logPath := getDebugLogPath()
	os.Remove(logPath)
	a.debugLog("=== Debug mode enabled ===")
	a.debugLog("Log file: %s", logPath)
}

// App represents the main application
type App struct {
	config          *config.Config
	themeManager    *theme.Manager
	bookmarkManager *bookmark.Manager
	previewManager  *preview.Manager
	navigator       *filesystem.Navigator
	renderer        *ui.Renderer
	fileOpsManager  *fileops.Manager
	
	// UI state
	showHelp        bool
	inPathEditMode  bool
	pathEditBuffer  string
	showContextMenu bool
	debugEnabled    bool
	
	// Mouse state
	lastClickTime   int64
	lastClickX      int
	lastClickY      int
	ctrlPressed     bool
	
	// Progress bar state
	progressHideTime  time.Time
	showProgress      bool
	lastOperationWasActive bool
}

// New creates a new application instance
func New() *App {
	cfg := config.New()
	tm := theme.NewManager()
	bm := bookmark.NewManager()
	pm := preview.NewManager()
	nav := filesystem.NewNavigator()
	fom := fileops.NewManager()
	
	// Load saved theme
	tm.LoadSavedTheme()
	
	renderer := ui.NewRenderer(tm, bm, pm, cfg, fom)
	
	return &App{
		config:          cfg,
		themeManager:    tm,
		bookmarkManager: bm,
		previewManager:  pm,
		navigator:       nav,
		renderer:        renderer,
		fileOpsManager:  fom,
		showHelp:        false,
		inPathEditMode:  false,
		pathEditBuffer:  "",
		showContextMenu: false,
	}
}

// Run starts the application
func (a *App) Run() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()
	
	// Enable mouse support if configured
	if a.config.MouseEnabled {
		termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	} else {
		termbox.SetInputMode(termbox.InputEsc)
	}
	
	// Load initial preview
	a.reloadPreview()
	a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	
	return a.eventLoop()
}

// pauseProgressUpdates is now a no-op (kept for compatibility)
func (a *App) pauseProgressUpdates() {
	// No longer needed - no background goroutine
}

// resumeProgressUpdates is now a no-op (kept for compatibility)
func (a *App) resumeProgressUpdates() {
	// No longer needed - no background goroutine
}

// eventLoop handles all user input events
func (a *App) eventLoop() error {
	for {
		// Poll for event (this blocks until an event occurs)
		a.debugLog("Main eventLoop: Waiting for event...")
		ev := termbox.PollEvent()
		a.debugLog("Main eventLoop: Got event type=%d key=%d", ev.Type, ev.Key)
		
		switch ev.Type {
		case termbox.EventResize:
			a.debugLog("Main eventLoop: Resize event")
			a.drawWithProgress()
			
		case termbox.EventKey:
			a.debugLog("Main eventLoop: Key event")
			// Track Ctrl key state
			if ev.Key == termbox.KeyCtrlC {
				a.ctrlPressed = true
			} else if ev.Ch != 0 || ev.Key == termbox.KeyEnter || ev.Key == termbox.KeyEsc {
				a.ctrlPressed = false
			}
			
			if a.inPathEditMode {
				if a.handlePathEditMode(ev) {
					a.debugLog("Main eventLoop: Path edit mode returned true, exiting")
					return nil
				}
				continue
			}
			
			if a.handleKeyEvent(ev) {
				a.debugLog("Main eventLoop: handleKeyEvent returned true, exiting")
				return nil
			}
			
			a.drawWithProgress()
			
		case termbox.EventMouse:
			if a.handleMouseEvent(ev) {
				a.debugLog("Main eventLoop: handleMouseEvent returned true, exiting")
				return nil
			}
			a.drawWithProgress()
		}
		
		// Update progress display after each event
		a.updateProgressDisplay()
	}
}

// updateProgressDisplay checks and updates progress bar display
func (a *App) updateProgressDisplay() {
	progress := a.fileOpsManager.GetProgress()
	if progress == nil {
		return
	}
	
	progress.Mu.RLock()
	isActive := progress.Active
	hasData := progress.TotalFiles > 0
	progress.Mu.RUnlock()
	
	wasOperationActive := a.lastOperationWasActive
	a.lastOperationWasActive = isActive
	
	// If operation just started, reset hide timer
	if !wasOperationActive && isActive {
		a.progressHideTime = time.Time{}
	}
	
	// If operation just finished, set hide timer
	if wasOperationActive && !isActive && hasData && a.progressHideTime.IsZero() {
		a.progressHideTime = time.Now().Add(2 * time.Second)
	}
	
	// Check if we should hide the progress bar (clear the data)
	if !a.progressHideTime.IsZero() && time.Now().After(a.progressHideTime) {
		a.progressHideTime = time.Time{}
		// Clear progress data so it won't show anymore
		progress.Mu.Lock()
		progress.TotalFiles = 0
		progress.ProcessedFiles = 0
		progress.TotalBytes = 0
		progress.ProcessedBytes = 0
		progress.CurrentFile = ""
		progress.Mu.Unlock()
		a.drawWithProgress()
		return
	}
	
	// Always redraw to show/update progress
	a.drawWithProgress()
}

// drawWithProgress draws the UI with progress bar if needed
func (a *App) drawWithProgress() {
	// Draw the main UI (without flushing)
	a.renderer.Draw(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	
	// ALWAYS try to draw progress bar - it will handle its own visibility
	progress := a.fileOpsManager.GetProgress()
	if progress != nil {
		a.renderer.DrawProgressBar(progress)
	}
	
	// Now flush everything to screen
	termbox.Flush()
}

// handlePathEditMode handles input when in path edit mode
func (a *App) handlePathEditMode(ev termbox.Event) bool {
	switch ev.Key {
	case termbox.KeyEnter:
		a.inPathEditMode = false
		newPath := filepath.Clean(a.pathEditBuffer)
		if stat, err := os.Stat(newPath); err == nil && stat.IsDir() {
			a.navigator.SetCurrentDir(newPath)
			a.previewManager.ResetScroll()
			a.reloadPreview()
		}
		
	case termbox.KeyEsc:
		a.inPathEditMode = false
		
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(a.pathEditBuffer) > 0 {
			a.pathEditBuffer = a.pathEditBuffer[:len(a.pathEditBuffer)-1]
		}
		
	default:
		if ev.Ch != 0 {
			a.pathEditBuffer += string(ev.Ch)
		}
	}
	
	a.drawWithProgress()
	return false
}

// handleKeyEvent handles keyboard input
func (a *App) handleKeyEvent(ev termbox.Event) bool {
	keys := a.config.Keys
	_, h := termbox.Size()
	visibleLines := h - 4
	
	// Handle special keys
	switch ev.Key {
	case termbox.KeyEsc:
		if a.showHelp {
			a.showHelp = false
			return false
		}
		return true // Quit
		
	case termbox.KeySpace:
		// Handle Space key for file selection
		if selectedPath := a.navigator.GetSelectedPath(); selectedPath != "" {
			a.fileOpsManager.ToggleSelection(selectedPath)
		}
		return false
		
	case termbox.KeyArrowUp:
		a.navigator.MoveUp(visibleLines)
		a.previewManager.ResetScroll()
		a.reloadPreview()
		return false
		
	case termbox.KeyArrowDown:
		a.navigator.MoveDown(visibleLines)
		a.previewManager.ResetScroll()
		a.reloadPreview()
		return false
		
	case termbox.KeyArrowLeft:
		if a.navigator.GoToParent() {
			a.fileOpsManager.ClearSelection() // Clear selections when changing directory
			a.reloadPreview()
		}
		return false
		
	case termbox.KeyArrowRight:
		if a.navigator.EnterDirectory() {
			a.fileOpsManager.ClearSelection() // Clear selections when changing directory
			a.reloadPreview()
		}
		return false
		
	case termbox.KeyPgup:
		a.navigator.MoveUpFast(visibleLines)
		a.previewManager.ResetScroll()
		a.reloadPreview()
		return false
		
	case termbox.KeyPgdn:
		a.navigator.MoveDownFast(visibleLines)
		a.previewManager.ResetScroll()
		a.reloadPreview()
		return false
		
	case termbox.KeyEnter:
		if selectedPath := a.navigator.GetSelectedPath(); selectedPath != "" {
			a.openWithEditorSelection(selectedPath)
		}
		return false
		
	case termbox.KeyCtrlS:
		// Show sorting popup
		a.debugLog("Main: Ctrl+S pressed, calling handleSortingPopup")
		a.handleSortingPopup()
		a.debugLog("Main: handleSortingPopup returned, continuing")
		return false
	}
	
	// Handle character keys
	switch ev.Ch {
	case keys.Quit:
		return true
		
	case keys.OpenTerminal:
		go a.openTerminal()
		return false
		
	case keys.Filter:
		a.pauseProgressUpdates()
		filter := a.renderer.Prompt("Filter: ", a.navigator)
		a.resumeProgressUpdates()
		a.navigator.SetFilter(filter)
		a.reloadPreview()
		return false
		
	case keys.ToggleHidden:
		a.navigator.ToggleHidden()
		a.reloadPreview()
		return false
		
	case keys.OpenThemePopup:
		a.pauseProgressUpdates()
		a.renderer.ShowThemeSelector(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
		a.resumeProgressUpdates()
		a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
		return false
		
	case keys.Help:
		a.showHelp = !a.showHelp
		return false
		
	case keys.BookmarkToggle:
		currentDir := a.navigator.GetCurrentDir()
		if a.bookmarkManager.IsBookmarked(currentDir) {
			a.pauseProgressUpdates()
			confirmed := a.renderer.ConfirmPrompt("Remove bookmark?")
			a.resumeProgressUpdates()
			if confirmed {
				a.bookmarkManager.Toggle(currentDir)
			}
		} else {
			a.bookmarkManager.Toggle(currentDir)
		}
		return false
		
	case keys.BookmarkPopup:
		if a.bookmarkManager.Count() > 0 {
			a.pauseProgressUpdates()
			path := a.renderer.ShowBookmarkPopup()
			a.resumeProgressUpdates()
			if path != "" {
				// Check if the bookmarked path still exists
				if stat, err := os.Stat(path); err == nil && stat.IsDir() {
					a.navigator.SetCurrentDir(path)
					a.navigator.ClearFilter()
					a.previewManager.ResetScroll()
					a.reloadPreview()
				} else {
					// Path doesn't exist anymore, remove the bookmark
					a.bookmarkManager.RemoveByPath(path)
					a.renderer.ShowMessage("Bookmark removed: path no longer exists")
				}
			}
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
		}
		return false
		
	case keys.EditPath:
		a.inPathEditMode = true
		a.pathEditBuffer = a.navigator.GetCurrentDir()
		return false
		
	case keys.ScrollDown:
		a.previewManager.ScrollDown(1, visibleLines)
		return false
		
	case keys.ScrollUp:
		a.previewManager.ScrollUp(1)
		return false
		
	case keys.ScrollDownFast:
		a.previewManager.ScrollDown(10, visibleLines)
		return false
		
	case keys.ScrollUpFast:
		a.previewManager.ScrollUp(10)
		return false
		
	case keys.TogglePath:
		a.config.ShowRawPath = !a.config.ShowRawPath
		return false
		
	case keys.ConfigMenu:
		a.handleConfigMenu()
		return false
		
	case ' ': // Space key for selection
		if selectedPath := a.navigator.GetSelectedPath(); selectedPath != "" {
			a.fileOpsManager.ToggleSelection(selectedPath)
		}
		return false
	}
	
	// Handle Alt/Option key for context menu (using Ctrl+O as alternative since Alt detection is limited)
	if ev.Key == termbox.KeyCtrlO {
		a.showContextMenu = true
		a.handleContextMenu()
		a.showContextMenu = false
		return false
	}
	
	return false
}

// reloadPreview reloads the preview for the currently selected file
func (a *App) reloadPreview() {
	selectedPath := a.navigator.GetSelectedPath()
	if selectedPath != "" {
		_, h := termbox.Size()
		maxLines := h * 10 // Load more lines for scrolling
		a.previewManager.LoadPreview(selectedPath, a.navigator.GetShowHidden(), maxLines)
	}
}

// openTerminal opens a terminal in the current directory
func (a *App) openTerminal() {
	currentDir := a.navigator.GetCurrentDir()
	ui.OpenTerminal(currentDir, a.config.TerminalApp)
}

// isTerminalEditor checks if an editor command is a terminal-based editor
func isTerminalEditor(editorCmd string) bool {
	terminalEditors := []string{"vim", "vi", "nvim", "nano", "emacs", "micro", "helix", "hx"}
	for _, te := range terminalEditors {
		if strings.Contains(strings.ToLower(editorCmd), te) {
			return true
		}
	}
	return false
}

// openEditor opens a file in the configured editor
func (a *App) openEditor(path string) {
	editorCmd := a.config.EditorCmd
	
	if isTerminalEditor(editorCmd) {
		// For terminal editors, we need to:
		// 1. Close termbox
		// 2. Run the editor in foreground
		// 3. Reinitialize termbox when done
		termbox.Close()
		
		cmd := exec.Command(editorCmd, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		_ = cmd.Run()
		
		// Reinitialize termbox
		_ = termbox.Init()
		a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	} else {
		// For GUI editors, run in background
		go exec.Command(editorCmd, path).Start()
	}
}


// openWithEditorSelection shows editor selection popup and opens file with chosen editor
func (a *App) openWithEditorSelection(path string) {
	// Build options list: 1) default editor, 2) terminal, 3) file explorer, 4) other editors
	var allOptions []config.EditorOption
	
	// Find the default editor in available editors to get its proper name
	availableEditors := config.GetAvailableEditors()
	var defaultEditorName string
	var defaultEditorDesc string
	foundDefault := false
	
	for _, editor := range availableEditors {
		if editor.Command == a.config.EditorCmd {
			defaultEditorName = editor.Name
			defaultEditorDesc = editor.Description
			foundDefault = true
			break
		}
	}
	
	// If default editor not found in available list, use command as name
	if !foundDefault {
		defaultEditorName = a.config.EditorCmd
		defaultEditorDesc = "Default editor"
	}
	
	// 1. Add default editor first
	defaultEditor := config.EditorOption{
		Name:        defaultEditorName,
		Command:     a.config.EditorCmd,
		IsTerminal:  isTerminalEditor(a.config.EditorCmd),
		Description: defaultEditorDesc,
	}
	allOptions = append(allOptions, defaultEditor)
	
	// 2. Add system actions (terminal and file explorer) second
	systemActions := config.GetSystemActions()
	allOptions = append(allOptions, systemActions...)
	
	// 3. Add other available editors (excluding the default one) last
	for _, editor := range availableEditors {
		if editor.Command != a.config.EditorCmd {
			allOptions = append(allOptions, editor)
		}
	}
	
	// Show editor selection popup
	a.pauseProgressUpdates()
	selectedIndex := a.renderer.ShowEditorSelectionPopup(allOptions, a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	a.resumeProgressUpdates()
	
	// Redraw the main UI after popup closes
	a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	
	// If user cancelled (pressed Esc), return
	if selectedIndex < 0 {
		return
	}
	
	// Get the selected option
	selectedOption := allOptions[selectedIndex]
	
	// Handle special system actions
	switch selectedOption.Command {
	case "__TERMINAL__":
		go a.openTerminal()
		return
	case "__FINDER__":
		a.revealInFinder(path)
		return
	case "__EXPLORER__":
		a.revealInExplorer(path)
		return
	case "__FILEMANAGER__":
		a.revealInFileManager(path)
		return
	}
	
	// Open file with the selected editor
	if selectedOption.IsTerminal {
		// Terminal editor - suspend UI
		termbox.Close()
		
		// Parse command (might have arguments like "emacs -nw")
		parts := strings.Fields(selectedOption.Command)
		cmd := exec.Command(parts[0], append(parts[1:], path)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		_ = cmd.Run()
		
		// Reinitialize termbox
		termbox.Init()
		a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	} else {
		// GUI editor - run in background
		parts := strings.Fields(selectedOption.Command)
		cmd := exec.Command(parts[0], append(parts[1:], path)...)
		_ = cmd.Start()
	}
}

// revealInFinder opens Finder and selects the file (macOS)
func (a *App) revealInFinder(path string) {
	exec.Command("open", "-R", path).Start()
}

// revealInExplorer opens Explorer and selects the file (Windows)
func (a *App) revealInExplorer(path string) {
	exec.Command("explorer", "/select,", path).Start()
}

// revealInFileManager opens the file manager (Linux)
func (a *App) revealInFileManager(path string) {
	// Try common Linux file managers
	fileManagers := []string{"xdg-open", "nautilus", "dolphin", "thunar", "nemo"}
	dir := filepath.Dir(path)
	
	for _, fm := range fileManagers {
		if _, err := exec.LookPath(fm); err == nil {
			if fm == "xdg-open" {
				exec.Command(fm, dir).Start()
			} else {
				exec.Command(fm, path).Start()
			}
			return
		}
	}
}

// handleContextMenu shows and handles the context menu for file operations
func (a *App) handleContextMenu() {
	selectedPath := a.navigator.GetSelectedPath()
	currentDir := a.navigator.GetCurrentDir()
	
	// Get selected files (or current file if none selected)
	selectedFiles := a.fileOpsManager.GetSelectedFiles()
	if len(selectedFiles) == 0 && selectedPath != "" {
		selectedFiles = []string{selectedPath}
	}
	
	// Build menu options based on context
	var options []string
	
	// If we have files selected or a file under cursor, show all options
	if len(selectedFiles) > 0 {
		options = []string{
			"Copy",
			"Cut",
			"Paste",
			"Rename",
			"Delete",
			"New File",
			"New Folder",
			"Cancel",
		}
	} else {
		// Empty directory - only show creation and paste options
		options = []string{
			"Paste",
			"New File",
			"New Folder",
			"Cancel",
		}
	}
	
	// Show context menu
	a.pauseProgressUpdates()
	selectedIndex := a.renderer.ShowContextMenu(options, a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	a.resumeProgressUpdates()
	
	// Redraw after menu closes
	a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	
	if selectedIndex < 0 || selectedIndex >= len(options) {
		return
	}
	
	// Handle selected operation
	switch options[selectedIndex] {
	case "Copy":
		a.fileOpsManager.Copy(selectedFiles)
		a.fileOpsManager.ClearSelection()
		
	case "Cut":
		a.fileOpsManager.Cut(selectedFiles)
		a.fileOpsManager.ClearSelection()
		
	case "Paste":
		if a.fileOpsManager.HasClipboard() {
			// Run paste operation in goroutine to allow UI updates
			go func() {
				err := a.fileOpsManager.Paste(currentDir)
				
				// Always refresh the view after operation
				a.navigator.Refresh()
				a.reloadPreview()
				a.drawWithProgress()
				
				if err != nil {
					a.renderer.ShowError(err.Error())
				}
			}()
		}
		
	case "Rename":
		if len(selectedFiles) == 1 {
			oldPath := selectedFiles[0]
			oldName := filepath.Base(oldPath)
			a.pauseProgressUpdates()
			newName := a.renderer.SimplePrompt("Rename to: ", a.navigator)
			a.resumeProgressUpdates()
			if newName != "" && newName != oldName {
				if err := a.fileOpsManager.Rename(oldPath, newName); err != nil {
					a.renderer.ShowError(err.Error())
				} else {
					a.navigator.Refresh()
					a.reloadPreview()
				}
			}
		}
		
	case "Delete":
		count := len(selectedFiles)
		confirmMsg := "Delete " + filepath.Base(selectedFiles[0]) + "?"
		if count > 1 {
			confirmMsg = fmt.Sprintf("Delete %d files?", count)
		}
		
		a.pauseProgressUpdates()
		confirmed := a.renderer.ConfirmPrompt(confirmMsg)
		a.resumeProgressUpdates()
		if confirmed {
			// Run delete operation in goroutine to allow UI updates
			go func() {
				err := a.fileOpsManager.Delete(selectedFiles)
				
				// Always refresh the view after operation
				a.fileOpsManager.ClearSelection()
				a.navigator.Refresh()
				a.reloadPreview()
				a.drawWithProgress()
				
				if err != nil {
					a.renderer.ShowError(err.Error())
				}
			}()
		}
		
	case "New File":
		a.pauseProgressUpdates()
		filename := a.renderer.SimplePrompt("New file name: ", a.navigator)
		a.resumeProgressUpdates()
		if filename != "" {
			if err := a.fileOpsManager.CreateFile(currentDir, filename); err != nil {
				a.renderer.ShowError(err.Error())
			} else {
				a.navigator.Refresh()
				a.reloadPreview()
			}
		}
		
	case "New Folder":
		a.pauseProgressUpdates()
		foldername := a.renderer.SimplePrompt("New folder name: ", a.navigator)
		a.resumeProgressUpdates()
		if foldername != "" {
			if err := a.fileOpsManager.CreateFolder(currentDir, foldername); err != nil {
				a.renderer.ShowError(err.Error())
			} else {
				a.navigator.Refresh()
				a.reloadPreview()
			}
		}
	}
	
	a.drawWithProgress()
}

// handleConfigMenu shows and handles the configuration menu
func (a *App) handleConfigMenu() {
	for {
		a.pauseProgressUpdates()
		choice := a.renderer.ShowConfigMenu()
		a.resumeProgressUpdates()
		a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
		
		// Handle choices that may have status indicators
		if strings.HasPrefix(choice, "Toggle Mouse Support") {
			choice = "Toggle Mouse Support"
		}
		if strings.HasPrefix(choice, "Toggle Icon Style") {
			choice = "Toggle Icon Style"
		}
		
		switch choice {
		case "Select Theme":
			a.pauseProgressUpdates()
			a.renderer.ShowThemeSelector(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			a.resumeProgressUpdates()
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			
		case "Create New Theme":
			a.pauseProgressUpdates()
			created := a.renderer.ShowThemeCreator()
			a.resumeProgressUpdates()
			if created {
				// Theme created successfully, reload themes
				a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			}
			
		case "Modify Theme Colors":
			a.pauseProgressUpdates()
			a.renderer.ShowThemeColorModifier(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			a.resumeProgressUpdates()
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			
		case "Rename Theme":
			a.pauseProgressUpdates()
			renamed := a.renderer.ShowThemeRenamer()
			a.resumeProgressUpdates()
			if renamed {
				a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			}
			
		case "Delete Theme":
			a.pauseProgressUpdates()
			deleted := a.renderer.ShowThemeDeleter()
			a.resumeProgressUpdates()
			if deleted {
				a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			}
			
		case "Set Default Editor":
			a.pauseProgressUpdates()
			editorCmd := a.renderer.ShowDefaultEditorSelector()
			a.resumeProgressUpdates()
			if editorCmd != "" {
				a.config.EditorCmd = editorCmd
				if err := config.SaveConfigFile(editorCmd, a.config.TerminalApp, &a.config.MouseEnabled, &a.config.UseAsciiIcons); err != nil {
					a.renderer.ShowError("Failed to save editor: " + err.Error())
				} else {
					a.renderer.ShowMessage("Default editor updated!")
				}
			}
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			
		case "Toggle Mouse Support":
			a.config.MouseEnabled = !a.config.MouseEnabled
			if err := config.SaveConfigFile(a.config.EditorCmd, a.config.TerminalApp, &a.config.MouseEnabled, &a.config.UseAsciiIcons); err != nil {
				a.renderer.ShowError("Failed to save mouse setting: " + err.Error())
			} else {
				status := "disabled"
				if a.config.MouseEnabled {
					status = "enabled"
					termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
				} else {
					termbox.SetInputMode(termbox.InputEsc)
				}
				a.renderer.ShowMessage("Mouse support " + status + "!")
			}
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			
		case "Toggle Icon Style":
			a.config.UseAsciiIcons = !a.config.UseAsciiIcons
			if err := config.SaveConfigFile(a.config.EditorCmd, a.config.TerminalApp, &a.config.MouseEnabled, &a.config.UseAsciiIcons); err != nil {
				a.renderer.ShowError("Failed to save icon setting: " + err.Error())
			} else {
				style := "ASCII"
				if !a.config.UseAsciiIcons {
					style = "Unicode"
				}
				a.renderer.ShowMessage("Icon style set to " + style + "!")
			}
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			
		case "Restore to Default":
			if a.renderer.ConfirmPrompt("Restore default theme?") {
				a.themeManager.RestoreDefaultTheme()
				a.renderer.ShowMessage("Default theme restored!")
			}
			a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
			
		case "Cancel":
			return
		}
	}
}

// handleSortingPopup shows and handles the sorting selection popup
func (a *App) handleSortingPopup() {
	a.pauseProgressUpdates()
	defer a.resumeProgressUpdates()
	
	selectedIndex := a.renderer.ShowSortingPopup(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	a.debugLog("handleSortingPopup: Popup returned with index=%d", selectedIndex)
	
	// Redraw after popup closes BEFORE resetting input mode
	// This ensures the screen is updated before any new events are processed
	a.renderer.DrawAndFlush(a.navigator, a.inPathEditMode, a.pathEditBuffer, a.showHelp)
	a.debugLog("handleSortingPopup: Screen redrawn")
	
	if selectedIndex < 0 {
		a.debugLog("handleSortingPopup: User cancelled, EXIT")
		return // User cancelled
	}
	
	// Map index to SortMode
	var sortMode filesystem.SortMode
	switch selectedIndex {
	case 0:
		sortMode = filesystem.SortByName
	case 1:
		sortMode = filesystem.SortBySize
	case 2:
		sortMode = filesystem.SortByModTime
	case 3:
		sortMode = filesystem.SortByExtension
	default:
		return
	}
	
	// Apply the new sort mode
	a.navigator.SetSortMode(sortMode)
	a.reloadPreview()
}

// handleMouseEvent handles mouse input events
func (a *App) handleMouseEvent(ev termbox.Event) bool {
	w, h := termbox.Size()
	
	// Calculate panel boundaries (same as in ui.Draw)
	parentPanelWidth := w / 5
	middlePanelWidth := (w * 2) / 5
	separator1Pos := parentPanelWidth
	middlePanelStart := separator1Pos + 1
	separator2Pos := middlePanelStart + middlePanelWidth
	
	// Handle mouse button events
	if ev.Key == termbox.MouseLeft {
		// Check if Ctrl is held (for context menu)
		if a.ctrlPressed {
			// Ctrl+Click - show context menu
			if ev.MouseX >= middlePanelStart && ev.MouseX < separator2Pos {
				// Only show context menu if clicking in middle panel
				if fileIndex := a.getFileIndexAtY(ev.MouseY, h); fileIndex >= 0 {
					// Move cursor to clicked item first
					a.navigator.SetCursor(fileIndex)
					a.reloadPreview()
				}
				a.handleContextMenu()
			}
			a.ctrlPressed = false // Reset after use
			return false
		}
		
		// Regular left click
		clickTime := time.Now().UnixMilli()
		isDoubleClick := false
		
		// Check if this is a double-click (within 500ms and same position)
		if clickTime-a.lastClickTime < 500 &&
		   ev.MouseX == a.lastClickX &&
		   ev.MouseY == a.lastClickY {
			isDoubleClick = true
		}
		
		a.lastClickTime = clickTime
		a.lastClickX = ev.MouseX
		a.lastClickY = ev.MouseY
		
		// Determine which panel was clicked
		if ev.MouseX >= middlePanelStart && ev.MouseX < separator2Pos {
			// Middle panel (current directory) clicked
			return a.handleMiddlePanelClick(ev.MouseY, h, isDoubleClick)
		} else if ev.MouseX < separator1Pos {
			// Parent panel clicked
			return a.handleParentPanelClick(ev.MouseY, h, isDoubleClick)
		}
		
	} else if ev.Key == termbox.MouseWheelUp {
		// Scroll up
		_, h := termbox.Size()
		visibleLines := h - 4
		a.navigator.MoveUp(visibleLines)
		a.previewManager.ResetScroll()
		a.reloadPreview()
		
	} else if ev.Key == termbox.MouseWheelDown {
		// Scroll down
		_, h := termbox.Size()
		visibleLines := h - 4
		a.navigator.MoveDown(visibleLines)
		a.previewManager.ResetScroll()
		a.reloadPreview()
	}
	
	return false
}

// handleMiddlePanelClick handles clicks in the middle panel (current directory)
func (a *App) handleMiddlePanelClick(mouseY, height int, isDoubleClick bool) bool {
	fileIndex := a.getFileIndexAtY(mouseY, height)
	if fileIndex < 0 {
		return false
	}
	
	// Move cursor to clicked item
	a.navigator.SetCursor(fileIndex)
	a.reloadPreview()
	
	if isDoubleClick {
		// Double-click: open file or enter directory
		if selectedPath := a.navigator.GetSelectedPath(); selectedPath != "" {
			info, err := os.Stat(selectedPath)
			if err == nil {
				if info.IsDir() {
					// Enter directory
					if a.navigator.EnterDirectory() {
						a.fileOpsManager.ClearSelection()
						a.reloadPreview()
					}
				} else {
					// Open file with default editor
					a.openWithEditorSelection(selectedPath)
				}
			}
		}
	}
	
	return false
}

// handleParentPanelClick handles clicks in the parent panel
func (a *App) handleParentPanelClick(mouseY, height int, isDoubleClick bool) bool {
	if isDoubleClick {
		// Double-click in parent panel: go to parent directory
		if a.navigator.GoToParent() {
			a.fileOpsManager.ClearSelection()
			a.reloadPreview()
		}
	}
	return false
}

// getFileIndexAtY calculates which file index corresponds to a Y coordinate
func (a *App) getFileIndexAtY(mouseY, height int) int {
	// Address bar is at y=0, files start at y=2
	if mouseY < 2 {
		return -1
	}
	
	// Calculate visible area
	visibleHeight := height - 4
	scrollOffset := a.navigator.GetScrollOffset()
	fileList := a.navigator.GetFileList()
	
	// Calculate file index
	relativeY := mouseY - 2
	if relativeY >= visibleHeight {
		return -1
	}
	
	fileIndex := scrollOffset + relativeY
	if fileIndex >= len(fileList) {
		return -1
	}
	
	return fileIndex
}
