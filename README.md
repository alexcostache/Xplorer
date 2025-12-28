# Xplorer

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/alexcostache/Xplorer/workflows/CI/badge.svg)](https://github.com/alexcostache/Xplorer/actions)

A modern, fast, and feature-rich terminal-based file explorer written in Go. Navigate your filesystem with ease using an intuitive three-panel interface, syntax highlighting, customizable themes, and powerful file operations.

![Xplorer Demo](https://via.placeholder.com/800x400.png?text=Xplorer+Screenshot)

## ✨ Features

### 🎯 Core Navigation
- **Three-panel layout** - Parent directory, current directory, and preview pane
- **Intuitive keyboard navigation** - Arrow keys for seamless browsing
- **Smart scrolling** - Automatic scroll offset management
- **Directory history** - Navigate back and forth through your path

### 📁 File Operations
- **Multi-file selection** - Select multiple files with Space key
- **Copy/Cut/Paste** - Full clipboard support with conflict resolution
- **Rename & Delete** - Safe file operations with confirmations
- **Context menu** - Quick access to operations with Ctrl+O
- **External editor integration** - Open files in your favorite editor

### 🎨 Customization
- **19 built-in themes** - From dark to light, monochrome to colorful
- **Custom themes** - Create your own JSON theme files
- **File type icons** - 40+ file extensions with unique icons
- **Syntax highlighting** - Code preview with Chroma lexer support
- **Color-coded files** - Visual distinction by file type

### 🔍 Advanced Features
- **Real-time filtering** - Search files as you type
- **Hidden files toggle** - Show/hide dotfiles instantly
- **Bookmarks** - Quick navigation to favorite directories
- **Preview pane** - View file contents or directory listings
- **Breadcrumb navigation** - Clear path visualization
- **Unicode support** - Full East Asian character support

### 🖥️ Supported Languages for Syntax Highlighting
Go • Python • JavaScript • TypeScript • Java • C/C++ • Rust • Ruby • PHP • HTML • CSS • JSON • Shell • and more

## 📦 Installation

### Prerequisites
- Go 1.23 or higher
- A terminal emulator

### Quick Install

#### Using the install script (Linux/macOS)
```bash
git clone https://github.com/alexcostache/Xplorer.git
cd Xplorer
chmod +x install-xp.sh
./install-xp.sh
```

#### Manual installation
```bash
# Clone the repository
git clone https://github.com/alexcostache/Xplorer.git
cd Xplorer

# Build the binary
go build -o xp

# Move to a directory in your PATH
sudo mv xp /usr/local/bin/

# Or for local installation
mkdir -p ~/.local/bin
mv xp ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"
```

#### Using Go install
```bash
go install github.com/alexcostache/Xplorer@latest
```

### Windows Installation
```bash
git clone https://github.com/alexcostache/Xplorer.git
cd Xplorer
go build -o xp.exe
# Move xp.exe to a directory in your PATH
```

## 🚀 Quick Start

Simply run:
```bash
xp
```

Or start in a specific directory:
```bash
cd /path/to/directory
xp
```

Enable debug mode:
```bash
xp --debug
```

## ⌨️ Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| `↑` / `↓` | Move cursor up/down |
| `←` / `→` | Navigate to parent/child directory |
| `Enter` | Open file in editor or enter directory |
| `Backspace` | Go to parent directory |

### File Operations
| Key | Action |
|-----|--------|
| `Space` | Select/deselect file |
| `Ctrl+O` | Open context menu (copy, cut, paste, rename, delete) |
| `Ctrl+C` | Copy selected files |
| `Ctrl+X` | Cut selected files |
| `Ctrl+V` | Paste files |

### View & Search
| Key | Action |
|-----|--------|
| `/` | Filter files (search) |
| `.` | Toggle hidden files |
| `p` | Toggle path display (breadcrumb/raw) |
| `[` / `]` | Scroll preview down/up |
| `{` / `}` | Fast scroll preview (10 lines) |

### Bookmarks & Themes
| Key | Action |
|-----|--------|
| `B` | Toggle bookmark for current directory |
| `b` | Open bookmark selector |
| `O` | Open theme selector |

### Other
| Key | Action |
|-----|--------|
| `?` | Toggle help |
| `t` | Open terminal at current directory |
| `e` | Edit path directly |
| `q` / `Esc` | Quit |

## ⚙️ Configuration

### Editor Configuration

Create `~/.xp_config.json`:
```json
{
  "editor_cmd": "code",
  "terminal_app": "iTerm"
}
```

Or use environment variables:
```bash
export EDITOR_CMD="vim"
export TERMINAL_APP="gnome-terminal"
```

### Supported Editors
- **Terminal**: vim, nvim, nano, emacs, micro, helix
- **GUI**: VS Code, Sublime Text, Atom, gedit

See [CONFIG.md](CONFIG.md) for detailed configuration options.

## 🎨 Themes

Xplorer comes with 19 beautiful built-in themes:

**Dark Themes**: Nightfall, Dusk, Midnight Breeze, Monochrome, Ember  
**Light Themes**: LightMode, Sunlight, Pastel, Soft Shell  
**Nature Themes**: Forest, Ocean, Earthtone, Willow, Autumn  
**Colorful Themes**: Sandstorm, Iceberg, Lavender, Sunset, Alex

Press `O` to open the theme selector and preview themes in real-time.

### Creating Custom Themes

Create a JSON file in the `themes/` directory:
```json
{
  "name": "MyTheme",
  "colors": {
    "text": "white",
    "background": "black",
    "highlight": "cyan",
    "highlight_text": "black",
    "footer": "green",
    "footer_bg": "black",
    "address_bar": "cyan",
    "address_bar_bg": "black",
    "separator": "cyan",
    "dim": "white",
    "filter": "black",
    "filter_bg": "cyan",
    "dir": "cyan"
  },
  "file_colors": {
    ".go": "cyan",
    ".py": "yellow",
    ".js": "green"
  }
}
```

See [themes/README.md](themes/README.md) for more details.

## 📚 Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - Project architecture and design patterns
- [FEATURES.md](FEATURES.md) - Comprehensive feature list
- [CONFIG.md](CONFIG.md) - Configuration guide
- [TESTING.md](TESTING.md) - Testing documentation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines

## 🧪 Testing

Run the test suite:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 🏗️ Architecture

Xplorer follows clean architecture principles with clear separation of concerns:

```
xplorer/
├── main.go              # Entry point
├── internal/
│   ├── app/            # Application orchestration
│   ├── config/         # Configuration management
│   ├── theme/          # Theme system
│   ├── bookmark/       # Bookmark management
│   ├── filesystem/     # File navigation
│   ├── preview/        # File preview & syntax highlighting
│   ├── fileops/        # File operations (copy, cut, paste, etc.)
│   └── ui/             # User interface rendering
└── themes/             # Theme JSON files
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture documentation.

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup
```bash
# Clone the repository
git clone https://github.com/alexcostache/Xplorer.git
cd Xplorer

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o xp
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [termbox-go](https://github.com/nsf/termbox-go) - Terminal UI library
- [chroma](https://github.com/alecthomas/chroma) - Syntax highlighting
- All contributors and users of Xplorer

## 📧 Contact

- GitHub: [@alexcostache](https://github.com/alexcostache)
- Issues: [GitHub Issues](https://github.com/alexcostache/Xplorer/issues)

## ⭐ Star History

If you find Xplorer useful, please consider giving it a star on GitHub!

---

**Made with ❤️ and Go**