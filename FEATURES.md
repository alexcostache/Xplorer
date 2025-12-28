# Xplorer - Feature List

A terminal-based file explorer written in Go with a three-panel interface, syntax highlighting, and customizable themes. Launch with the `xp` command.

---

## Core Navigation
- Three-panel layout (parent directory, current directory, preview)
- Arrow key navigation (up/down for files, left/right for directories)
- Cursor wrapping (top/bottom navigation loops)
- Automatic scrolling with scroll offset management
- Directory traversal with history tracking

## File Operations
- Open files in external editor (configurable: VS Code, notepad, nano, etc.)
- Open terminal at current directory
- Platform-specific terminal support (iTerm on macOS, cmd on Windows, x-terminal-emulator on Linux)

## Filtering & Search
- Real-time file filtering with `/` key
- Case-insensitive search
- Auto-cursor positioning to best match
- Toggle hidden files visibility with `.` key

## Preview Panel
- Directory preview (shows contents with icons)
- Text file preview with syntax highlighting (using Chroma lexer)
- Scrollable preview for long files (up/down with `[` and `]`)
- Fast scroll (10 lines at a time with `{` and `}`)
- Binary file detection
- File type descriptions for non-readable files
- Language detection for syntax highlighting:
  - Go, Python, JavaScript, TypeScript
  - C/C++, Java, Rust, Ruby, PHP
  - HTML, CSS, JSON, Shell

## Bookmarks
- Add/remove bookmarks with `B` key
- Jump to bookmarks with `b` key
- Bookmark popup selector with keyboard navigation
- Persistent bookmark storage (`~/.xp_bookmarks.json`)
- Star indicator (★) for bookmarked items

## Themes
- **13 built-in themes:**
  - Nightfall
  - Dusk
  - Sandstorm
  - LightMode
  - Monochrome
  - Midnight Breeze
  - Sunlight
  - Earthtone
  - Iceberg
  - Willow
  - Soft Shell
  - Ember
  - Pastel
- Theme selector popup with `O` key
- Live theme preview
- Persistent theme saving (`~/.xp_theme`)
- Customizable colors for all UI elements

## UI Components
- **Breadcrumb address bar** with hierarchical path display
- **Raw path display mode** - Toggle with `p` key to show full copyable path
- Home directory abbreviation (`~` symbol)
- Editable path mode with `e` key
- **Metadata footer bar** showing:
  - File name, size, permissions, modification time
  - Item counts for all three panels
  - Hidden files toggle status
- **File type icons** (40+ file extensions supported)
- Color-coded file extensions
- Active folder highlighting in parent panel
- Vertical panel separators
- Help panel with `?` key

## Visual Indicators
- Current selection highlighting
- Bookmark star indicators (★)
- **File type icons** for:
  - Folders
  - Programming languages (Go, Python, JS, TS, Java, C/C++, Rust)
  - Web files (HTML, CSS, JSON)
  - Config files (YAML, TOML)
  - Archives (ZIP, TAR, GZ, RAR)
  - Media files (images, audio, video)
  - Documents (PDF, TXT, MD, LOG)
  - Shell scripts
- Color-coded files by extension

## Configuration
- Environment variable support:
  - `EDITOR_CMD` - Custom editor command
  - `TERMINAL_APP` - Custom terminal application
- Platform-specific defaults
- Configurable keybindings
- Home directory-based config storage

## Advanced Features
- Wide character support (East Asian text)
- Proper Unicode handling with rune width calculation
- Syntax highlighting with Chroma and Monokai style
- Smart RGB to termbox color mapping
- Terminal resize handling
- Error handling for inaccessible paths
- Smart file size formatting (B, KB, MB, GB, etc.)

## Keybindings

| Key | Action |
|-----|--------|
| `/` | Filter files |
| `.` | Toggle hidden files |
| `q` | Quit |
| `?` | Toggle help |
| `t` | Open terminal |
| `B` | Toggle bookmark for current directory |
| `b` | Open bookmark selector |
| `e` | Edit path directly |
| `p` | Toggle path display (breadcrumb/raw) |
| `[` | Scroll preview down |
| `]` | Scroll preview up |
| `{` | Scroll preview down fast (10 lines) |
| `}` | Scroll preview up fast (10 lines) |
| `O` | Open theme selector |
| `Enter` | Open file in editor |
| `Esc` | Close popups/quit |
| `↑↓` | Navigate files |
| `←→` | Navigate directories |

---

## Technical Stack
- **Language:** Go 1.24.2
- **Dependencies:**
  - `github.com/nsf/termbox-go` - Terminal UI
  - `github.com/alecthomas/chroma` - Syntax highlighting
  - `golang.org/x/text` - Unicode text processing

---

*This feature list is maintained to track current capabilities and plan future enhancements.*
