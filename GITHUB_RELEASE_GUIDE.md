# GitHub Release Guide for Xplorer

This document provides step-by-step instructions for publishing Xplorer to GitHub.

## Pre-Release Checklist

✅ All files have been prepared:
- [x] README.md - Comprehensive project documentation
- [x] LICENSE - MIT License
- [x] CONTRIBUTING.md - Contribution guidelines
- [x] ARCHITECTURE.md - Architecture documentation
- [x] FEATURES.md - Feature list
- [x] CONFIG.md - Configuration guide
- [x] TESTING.md - Testing documentation
- [x] .gitignore - Proper Go project ignores
- [x] .github/workflows/ci.yml - CI workflow
- [x] .github/workflows/release.yml - Release workflow
- [x] go.mod - Updated with GitHub module path
- [x] All imports updated to use github.com/alexcostache/Xplorer

✅ Code quality:
- [x] All tests passing
- [x] Build successful
- [x] Module path updated
- [x] Documentation consistent

## Step 1: Initialize Git Repository

```bash
cd /Users/alexandru.stefan.costache/projects/xp

# Initialize git if not already done
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Xplorer v1.0.0

- Terminal-based file explorer with three-panel interface
- 19 built-in themes with custom theme support
- Syntax highlighting for 14+ programming languages
- File operations: copy, cut, paste, rename, delete
- Bookmarks for quick navigation
- Real-time file filtering and search
- Multi-file selection
- Preview pane with scrolling
- Platform-specific editor and terminal integration
- Comprehensive test suite
- CI/CD with GitHub Actions
"
```

## Step 2: Add Remote Repository

```bash
# Add the GitHub repository as remote
git remote add origin https://github.com/alexcostache/Xplorer.git

# Verify remote
git remote -v
```

## Step 3: Push to GitHub

```bash
# Push to main branch
git branch -M main
git push -u origin main
```

## Step 4: Create First Release

### Option A: Using GitHub Web Interface

1. Go to https://github.com/alexcostache/Xplorer
2. Click on "Releases" → "Create a new release"
3. Click "Choose a tag" and type `v1.0.0`
4. Set release title: `Xplorer v1.0.0 - Initial Release`
5. Add release notes (see template below)
6. Click "Publish release"

### Option B: Using Git Tags

```bash
# Create and push tag
git tag -a v1.0.0 -m "Xplorer v1.0.0 - Initial Release"
git push origin v1.0.0
```

The GitHub Actions release workflow will automatically:
- Build binaries for Linux, macOS, and Windows (AMD64 and ARM64)
- Create release archives
- Generate checksums
- Publish the release with all artifacts

## Release Notes Template

```markdown
# Xplorer v1.0.0 - Initial Release

A modern, fast, and feature-rich terminal-based file explorer written in Go.

## 🎉 Features

### Core Functionality
- **Three-panel interface** - Parent directory, current directory, and preview pane
- **Intuitive navigation** - Arrow keys for seamless browsing
- **File operations** - Copy, cut, paste, rename, delete with multi-file selection
- **Real-time filtering** - Search files as you type
- **Bookmarks** - Quick navigation to favorite directories

### Customization
- **19 built-in themes** - From dark to light, monochrome to colorful
- **Custom themes** - Create your own JSON theme files
- **Syntax highlighting** - Code preview for 14+ programming languages
- **File type icons** - 40+ file extensions with unique icons

### Advanced Features
- **Preview pane** - View file contents or directory listings with scrolling
- **Hidden files toggle** - Show/hide dotfiles instantly
- **Platform integration** - Works with your favorite editor and terminal
- **Unicode support** - Full East Asian character support

## 📦 Installation

### Quick Install

**Linux/macOS:**
```bash
# Download the appropriate binary
tar -xzf xp-*.tar.gz
sudo mv xp-* /usr/local/bin/xp
xp
```

**Windows:**
```bash
# Extract the zip file and add to PATH
xp.exe
```

### Using Go
```bash
go install github.com/alexcostache/Xplorer@latest
```

## 🚀 Quick Start

```bash
# Launch Xplorer
xp

# Launch in specific directory
cd /path/to/directory && xp

# Enable debug mode
xp --debug
```

## 📚 Documentation

- [README.md](README.md) - Getting started guide
- [FEATURES.md](FEATURES.md) - Complete feature list
- [CONFIG.md](CONFIG.md) - Configuration options
- [ARCHITECTURE.md](ARCHITECTURE.md) - Architecture documentation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines

## 🔑 Key Bindings

- `↑/↓` - Navigate files
- `←/→` - Navigate directories
- `Space` - Select/deselect file
- `Ctrl+O` - Open context menu
- `/` - Filter files
- `.` - Toggle hidden files
- `B` - Toggle bookmark
- `b` - Open bookmark selector
- `O` - Open theme selector
- `?` - Show help
- `q` - Quit

## 🛠️ Technical Details

- **Language:** Go 1.24.2
- **Dependencies:** termbox-go, chroma, golang.org/x/text
- **Platforms:** Linux, macOS, Windows (AMD64 and ARM64)
- **License:** MIT

## 🙏 Acknowledgments

- [termbox-go](https://github.com/nsf/termbox-go) - Terminal UI library
- [chroma](https://github.com/alecthomas/chroma) - Syntax highlighting

## 📝 Changelog

### Added
- Initial release with core file explorer functionality
- Three-panel interface with preview
- 19 built-in themes
- File operations (copy, cut, paste, rename, delete)
- Multi-file selection
- Bookmarks system
- Real-time file filtering
- Syntax highlighting for code files
- Platform-specific editor integration
- Comprehensive test suite
- CI/CD with GitHub Actions

---

**Full Changelog**: https://github.com/alexcostache/Xplorer/commits/v1.0.0
```

## Step 5: Post-Release Tasks

### Update Repository Settings

1. **Add repository description:**
   - Go to repository settings
   - Add: "A modern terminal-based file explorer written in Go"
   - Add topics: `go`, `tui`, `file-explorer`, `terminal`, `file-manager`, `golang`

2. **Enable GitHub Pages (optional):**
   - Settings → Pages
   - Source: Deploy from a branch
   - Branch: main, /docs (if you create documentation)

3. **Configure branch protection:**
   - Settings → Branches
   - Add rule for `main` branch
   - Require pull request reviews
   - Require status checks to pass

### Create Additional Resources

1. **Add repository banner/logo** (optional)
   - Create a banner image
   - Add to README.md

2. **Create GitHub Discussions** (optional)
   - Settings → Features → Enable Discussions
   - Create categories: Announcements, Q&A, Ideas

3. **Add issue templates:**
   - Create `.github/ISSUE_TEMPLATE/bug_report.md`
   - Create `.github/ISSUE_TEMPLATE/feature_request.md`

## Step 6: Announce Release

Consider announcing on:
- Reddit: r/golang, r/commandline, r/linux
- Hacker News
- Twitter/X
- Dev.to
- LinkedIn

## Maintenance

### Regular Tasks
- Monitor issues and pull requests
- Update dependencies: `go get -u ./... && go mod tidy`
- Run tests before merging: `go test ./...`
- Create releases for new versions
- Update documentation as features are added

### Version Numbering
Follow Semantic Versioning (SemVer):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backward compatible)
- **PATCH** version for bug fixes (backward compatible)

Example: v1.0.0 → v1.1.0 (new feature) → v1.1.1 (bug fix) → v2.0.0 (breaking change)

## Troubleshooting

### Common Issues

**Issue:** Push rejected
```bash
# Solution: Pull first, then push
git pull origin main --rebase
git push origin main
```

**Issue:** CI/CD workflow fails
- Check workflow logs in Actions tab
- Verify Go version in workflow matches go.mod
- Ensure all tests pass locally first

**Issue:** Release workflow doesn't trigger
- Verify tag format: `v*.*.*` (e.g., v1.0.0)
- Check workflow file syntax
- Ensure GitHub Actions are enabled

## Support

For questions or issues:
- Open an issue: https://github.com/alexcostache/Xplorer/issues
- Check documentation: https://github.com/alexcostache/Xplorer/tree/main/docs
- Review existing issues and discussions

---

**Ready to publish!** Follow the steps above to release Xplorer to the world. 🚀