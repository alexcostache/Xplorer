# Xplorer - Architecture Documentation

## Overview

Xplorer (command: `xp`) is a terminal-based file explorer written in Go with a modular architecture designed for maintainability, extensibility, and testability. The application follows clean architecture principles with clear separation of concerns.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│                    (Entry Point - 11 lines)                  │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/app/                             │
│              Application Orchestration Layer                 │
│  • Coordinates all modules                                   │
│  • Manages application lifecycle                             │
│  • Handles event loop and user input                         │
└──┬────────┬────────┬────────┬────────┬────────┬────────┬───┘
   │        │        │        │        │        │        │
   ▼        ▼        ▼        ▼        ▼        ▼        ▼
┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
│Config│ │Theme │ │Book- │ │File- │ │Pre-  │ │ UI   │ │File  │
│      │ │      │ │marks │ │system│ │view  │ │      │ │Ops   │
└──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘
```

## Module Structure

### 1. **internal/app/** - Application Layer
**Purpose**: Orchestrates all modules and manages the application lifecycle.

**Key Components**:
- `App`: Main application struct that holds references to all managers
- Event loop handling
- User input routing
- State management (help mode, path edit mode, etc.)

**Responsibilities**:
- Initialize all subsystems
- Coordinate interactions between modules
- Handle keyboard events
- Manage application state transitions

**Key Files**:
- `app.go`: Main application logic (238 lines)

---

### 2. **internal/config/** - Configuration Layer
**Purpose**: Manages application configuration and platform-specific settings.

**Key Components**:
- `Config`: Configuration struct with editor, terminal, and display settings
- `KeyBindings`: Keyboard shortcut mappings
- File icon and description utilities

**Responsibilities**:
- Load platform-specific defaults (Windows, macOS, Linux)
- Provide file icons based on extensions
- Describe file types
- Manage keybindings

**Key Files**:
- `config.go`: Configuration management (165 lines)
- `config_test.go`: Unit tests (96 lines)

**Key Functions**:
- `New()`: Creates configuration with platform defaults
- `FileIcon(name, isDir)`: Returns icon for file type
- `DescribeFileByExt(name)`: Returns human-readable file description

---

### 3. **internal/theme/** - Theme Management Layer
**Purpose**: Handles color schemes and visual theming.

**Key Components**:
- `Theme`: Color scheme definition
- `Manager`: Theme management and persistence

**Responsibilities**:
- Store and manage multiple themes
- Apply themes to UI
- Persist theme selection
- Provide file coloring based on extensions

**Key Files**:
- `theme.go`: Theme definitions and management (408 lines)

**Built-in Themes**:
- Nightfall, Dusk, Sandstorm, LightMode, Monochrome
- Midnight Breeze, Sunlight, Earthtone, Iceberg
- Willow, Soft Shell, Ember, Pastel

**Key Functions**:
- `NewManager()`: Creates theme manager
- `SetThemeByName(name)`: Switches to named theme
- `GetFileColor(name, isDir)`: Returns color for file type
- `LoadSavedTheme()`: Loads previously saved theme

---

### 4. **internal/bookmark/** - Bookmark Management Layer
**Purpose**: Manages directory bookmarks for quick navigation.

**Key Components**:
- `Bookmark`: Bookmark data structure
- `Manager`: Bookmark CRUD operations

**Responsibilities**:
- Add/remove bookmarks
- Check if path is bookmarked
- Persist bookmarks to disk (~/.xp_bookmarks.json)
- Provide bookmark list for UI

**Key Files**:
- `bookmark.go`: Bookmark management (108 lines)
- `bookmark_test.go`: Unit tests (82 lines)

**Key Functions**:
- `Toggle(path)`: Add or remove bookmark
- `IsBookmarked(path)`: Check bookmark status
- `GetAll()`: Retrieve all bookmarks
- `Save()/Load()`: Persist to/from disk

---

### 5. **internal/filesystem/** - File System Navigation Layer
**Purpose**: Handles file system operations and navigation state.

**Key Components**:
- `Navigator`: File system navigation state and operations

**Responsibilities**:
- Directory traversal (up/down, enter/back)
- File list management and filtering
- Cursor position and scrolling
- Hidden file visibility toggle
- Navigation history

**Key Files**:
- `filesystem.go`: Navigation logic (276 lines)

**Key Functions**:
- `RefreshFileList()`: Updates file list based on filters
- `MoveUp()/MoveDown()`: Cursor navigation
- `GoToParent()/EnterDirectory()`: Directory navigation
- `SetFilter(filter)`: Apply search filter
- `ToggleHidden()`: Show/hide hidden files

---

### 6. **internal/preview/** - File Preview Layer
**Purpose**: Handles file content preview and syntax highlighting.

**Key Components**:
- `Manager`: Preview state and operations
- Syntax highlighting using Chroma library

**Responsibilities**:
- Load file/directory previews
- Syntax highlighting for code files
- Scroll management for preview pane
- Binary file detection
- Language detection from file extensions

**Key Files**:
- `preview.go`: Preview logic (304 lines)
- `preview_test.go`: Unit tests (63 lines)

**Key Functions**:
- `LoadPreview(path, showHidden, maxLines)`: Load file/dir preview
- `DrawText(x, y, line, lang, ...)`: Render syntax-highlighted text
- `DetectLanguage(filename)`: Detect programming language
- `ScrollUp()/ScrollDown()`: Scroll preview content

**Supported Languages**:
- Go, Python, JavaScript, TypeScript, Java, C/C++
- Ruby, Rust, PHP, HTML, CSS, JSON, Shell

---

### 7. **internal/ui/** - User Interface Layer
**Purpose**: Renders all UI components and handles user interactions.

**Key Components**:
- `Renderer`: Main UI rendering engine
- Panel drawing functions
- Popup dialogs (help, themes, bookmarks, context menu)
- Input prompts

**Responsibilities**:
- Draw three-panel layout (parent, current, preview)
- Render address bar, filter bar, metadata bar
- Show help overlay
- Display theme/bookmark selection popups
- Show context menu for file operations
- Handle user prompts and confirmations
- Display file selection indicators

**Key Files**:
- `ui.go`: UI rendering (897 lines)
- `ui_test.go`: Unit tests (93 lines)

**Key Functions**:
- `Draw(nav, inPathEditMode, pathEditBuffer, showHelp)`: Main render function
- `ShowThemeSelector()`: Theme selection popup
- `ShowBookmarkPopup()`: Bookmark navigation popup
- `ShowContextMenu(options, ...)`: File operations context menu
- `ShowError(message)`: Error message display
- `Prompt(label, nav)`: Input prompt
- `ConfirmPrompt(message)`: Yes/no confirmation

**UI Layout**:
```
┌─────────────────────────────────────────────────────────┐
│ Address Bar (breadcrumb or raw path)                    │
├──────────┬──────────────────┬──────────────────────────┤
│ Parent   │ Current Dir      │ Preview                  │
│ Dir      │ (with cursor)    │ (file content or         │
│          │ ✓ selected file  │  directory listing)      │
│          │   normal file    │                          │
├──────────┴──────────────────┴──────────────────────────┤
│ Filter: [search text]                                   │
├─────────────────────────────────────────────────────────┤
│ Metadata: name | size | mode | time | counts | hidden  │
└─────────────────────────────────────────────────────────┘
```

---

### 8. **internal/fileops/** - File Operations Layer
**Purpose**: Handles file system operations (copy, cut, paste, rename, delete).

**Key Components**:
- `Manager`: File operations state and execution
- `Operation`: Operation type enum (Copy, Cut)

**Responsibilities**:
- Manage file selection state
- Copy/cut files to clipboard
- Paste files with conflict resolution
- Rename files with validation
- Delete files and directories
- Handle recursive directory operations

**Key Files**:
- `fileops.go`: File operations logic (240 lines)

**Key Functions**:
- `ToggleSelection(path)`: Select/deselect file
- `IsSelected(path)`: Check selection status
- `Copy(files)`: Copy files to clipboard
- `Cut(files)`: Cut files to clipboard
- `Paste(destDir)`: Paste clipboard contents
- `Rename(oldPath, newName)`: Rename file
- `Delete(files)`: Delete files/directories
- `GetSelectedFiles()`: Get list of selected files

**Features**:
- Multi-file selection with Space key
- Context menu with Ctrl+O
- Automatic name conflict resolution (adds _copy1, _copy2, etc.)
- Recursive directory copy/delete
- Permission preservation on copy

---

## Data Flow

### 1. Application Startup
```
main() → app.New() → Initialize all managers → termbox.Init() → app.Run()
```

### 2. User Input Flow
```
Event Loop → Handle Key Event → Update State → Reload Preview → Render UI
```

### 3. Navigation Flow
```
User Input → Navigator.MoveUp/Down/Left/Right() → RefreshFileList() 
→ PreviewManager.LoadPreview() → Renderer.Draw()
```

### 4. Theme Change Flow
```
User presses 'O' → Renderer.ShowThemeSelector() → ThemeManager.SetThemeByName()
→ Save to disk → Renderer.Draw()
```

### 5. Bookmark Flow
```
User presses 'B' → BookmarkManager.Toggle() → Save to disk → Renderer.Draw()
User presses 'b' → Renderer.ShowBookmarkPopup() → Navigator.SetCurrentDir()
→ RefreshFileList() → Renderer.Draw()
```

---

## Design Patterns

### 1. **Manager Pattern**
Each domain (theme, bookmark, preview) has a Manager that encapsulates all operations and state for that domain.

### 2. **Dependency Injection**
The App struct receives all managers through its constructor, making testing and swapping implementations easy.

### 3. **Single Responsibility Principle**
Each module has a clear, focused responsibility:
- Config: Configuration only
- Theme: Visual theming only
- Bookmark: Bookmark management only
- Filesystem: Navigation only
- Preview: File preview only
- UI: Rendering only
- App: Orchestration only

### 4. **Separation of Concerns**
- Business logic is separate from UI rendering
- State management is separate from presentation
- File operations are separate from display logic

---

## Testing Strategy

### Unit Tests
Each module has its own test file testing core functionality:
- `config_test.go`: Configuration and file utilities
- `bookmark_test.go`: Bookmark CRUD operations
- `preview_test.go`: Language detection and utilities
- `ui_test.go`: Formatting and utility functions

### Test Coverage
- Configuration defaults and platform detection
- File icon and description mapping
- Bookmark add/remove/toggle operations
- Language detection for syntax highlighting
- Size formatting and display utilities

---

## Extension Points

### Adding New Features

#### 1. **New File Type Support**
- Add extension to `config.FileIcon()` map
- Add color to `theme.getDefaultFileColors()` map
- Add language to `preview.DetectLanguage()` map

#### 2. **New Theme**
- Add theme definition to `theme.getBuiltInThemes()` array
- Theme will automatically appear in theme selector

#### 3. **New Keyboard Shortcut**
- Add key to `config.KeyBindings` struct
- Set default in `config.defaultKeyBindings()`
- Handle in `app.handleKeyEvent()`

#### 4. **New UI Panel**
- Add drawing function to `ui.Renderer`
- Call from `ui.Draw()` method
- Update layout calculations as needed

#### 5. **New Navigation Feature**
- Add method to `filesystem.Navigator`
- Call from `app.handleKeyEvent()`
- Update UI in `ui.Renderer.Draw()`

---

## File Organization

```
Xplorer/
├── main.go                    # Entry point
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── README.md                  # Project documentation
├── LICENSE                    # MIT License
├── CONTRIBUTING.md            # Contribution guidelines
├── ARCHITECTURE.md            # This file
├── FEATURES.md                # Feature documentation
├── CONFIG.md                  # Configuration guide
├── TESTING.md                 # Testing documentation
├── internal/
│   ├── app/
│   │   └── app.go            # Application orchestration
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── theme/
│   │   └── theme.go          # Theme management
│   ├── bookmark/
│   │   └── bookmark.go       # Bookmark management
│   ├── filesystem/
│   │   └── filesystem.go     # File navigation
│   ├── preview/
│   │   └── preview.go        # File preview & syntax highlighting
│   ├── fileops/
│   │   └── fileops.go        # File operations
│   └── ui/
│       └── ui.go             # UI rendering
├── themes/                    # JSON theme files
│   ├── nightfall.json
│   ├── forest.json
│   ├── ocean.json
│   └── [16 more themes...]
├── tests/                     # Test files
│   ├── bookmark_test.go
│   ├── config_test.go
│   ├── preview_test.go
│   ├── progress_test.go
│   └── ui_test.go
└── .github/
    └── workflows/             # CI/CD workflows
        ├── ci.yml
        └── release.yml
```

---

## Dependencies

### External Libraries
- **github.com/nsf/termbox-go**: Terminal UI library
- **github.com/alecthomas/chroma**: Syntax highlighting
- **golang.org/x/text**: Unicode text processing

### Standard Library
- `os`, `path/filepath`: File system operations
- `encoding/json`: Bookmark persistence
- `bufio`: File reading
- `sort`: File list sorting
- `strings`: String manipulation
- `runtime`: Platform detection

---

## Performance Considerations

### 1. **Lazy Loading**
- Preview content is loaded only when needed
- File lists are filtered on-demand

### 2. **Efficient Rendering**
- Only visible items are rendered
- Scroll offsets prevent unnecessary drawing

### 3. **Caching**
- Preview content is cached until navigation changes
- Theme colors are pre-computed

### 4. **Minimal Allocations**
- Reuse of buffers where possible
- Efficient string operations

---

## Implemented Features

### File Operations (✅ Completed)
The file operations module provides comprehensive file management:
- **Selection**: Space key to select/deselect files (✓ indicator shown)
- **Context Menu**: Ctrl+O to open operations menu
- **Copy**: Copy selected files to clipboard
- **Cut**: Cut selected files for moving
- **Paste**: Paste with automatic conflict resolution
- **Rename**: Rename single file with validation
- **Delete**: Delete files/directories with confirmation

### Future Enhancements

### Potential Additions
1. **Search**: Full-text search across files
2. **Tabs**: Multiple directory tabs
3. **Git Integration**: Show git status in file list
4. **Custom Commands**: User-defined file operations
5. **Plugins**: Plugin system for extensions
6. **Remote Files**: SSH/FTP support
7. **Archive Support**: Browse inside zip/tar files
8. **Batch Operations**: Apply operations to multiple files
9. **Undo/Redo**: Undo file operations

### How to Add
Each enhancement would be a new module in `internal/`:
- `internal/search/`: Search functionality
- `internal/tabs/`: Tab management
- `internal/git/`: Git integration
- etc.

The modular architecture makes these additions straightforward without disrupting existing functionality.

---

## Maintenance Guidelines

### Code Style
- Follow Go conventions and idioms
- Keep functions focused and small
- Use meaningful variable names
- Add comments for complex logic

### Testing
- Write unit tests for new features
- Maintain test coverage above 70%
- Test edge cases and error conditions

### Documentation
- Update ARCHITECTURE.md for structural changes
- Update FEATURES.md for new features
- Keep inline comments current

### Refactoring
- Extract common patterns into utilities
- Keep modules independent
- Avoid circular dependencies
- Maintain clear interfaces between modules

---

## Troubleshooting

### Common Issues

**Issue**: Tests failing after refactor
- **Solution**: Update test imports and function calls

**Issue**: Theme not persisting
- **Solution**: Check ~/.xp_theme file permissions

**Issue**: Bookmarks not saving
- **Solution**: Check ~/.xp_bookmarks.json file permissions

**Issue**: Syntax highlighting not working
- **Solution**: Verify Chroma library is installed

---

## Contact & Contributing

For questions about the architecture or to propose changes:
1. Review this document first
2. Check existing code for patterns
3. Follow the established module structure
4. Write tests for new functionality
5. Update documentation

The modular architecture is designed to make contributions easy and safe!