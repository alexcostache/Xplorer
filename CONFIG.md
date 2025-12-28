# Xplorer Configuration Guide

## Configuration Options

Xplorer (command: `xp`) supports multiple ways to configure your default editor and terminal application.

### Priority Order

Configuration is loaded in the following priority (highest to lowest):

1. **Config File** (`~/.xp_config.json`)
2. **Environment Variables** (`EDITOR_CMD`, `TERMINAL_APP`)
3. **Platform Defaults**

## Opening Files

XP provides two ways to open files:

### 1. Default Editor (Enter Key)
Press **Enter** on a file to open it with your configured default editor.

### 2. Open With Menu (O Key)
Press **O** (lowercase 'o') to open a selection menu with multiple editor options:
- **Terminal Editors**: vim, nano, emacs (run in foreground, XP UI suspends)
- **GUI Editors**: VS Code, Sublime Text, Atom, etc. (run in background)

The "Open With" menu allows you to choose a different editor for a specific file without changing your default configuration.

### Config File

Create a JSON file at `~/.xplorer_config.json` with your preferences:

```json
{
  "editor_cmd": "code",
  "terminal_app": "iTerm"
}
```

#### Supported Options

- **`editor_cmd`**: Command to open files (e.g., `"code"`, `"vim"`, `"nano"`, `"subl"`)
- **`terminal_app`**: Terminal application to open (e.g., `"iTerm"`, `"Terminal"`, `"gnome-terminal"`)

### Environment Variables

Set environment variables in your shell profile:

```bash
# In ~/.bashrc, ~/.zshrc, etc.
export EDITOR_CMD="vim"
export TERMINAL_APP="gnome-terminal"
```

### Platform Defaults

If no config file or environment variables are set, XP uses these defaults:

#### macOS
- Editor: `vim`
- Terminal: `iTerm`

#### Windows
- Editor: `notepad`
- Terminal: `cmd`

#### Linux/Unix
- Editor: `vim`
- Terminal: `x-terminal-emulator`

## Terminal vs GUI Editors

XP automatically detects whether your editor is terminal-based or GUI-based:

### Terminal Editors (Supported)
These editors run in the terminal and XP will properly suspend/resume:
- **vim**, **vi**, **nvim** (Neovim)
- **nano**
- **emacs** (terminal mode)
- **micro**
- **helix**

When you open a file with a terminal editor, Xplorer will:
1. Suspend the UI
2. Launch the editor in the foreground
3. Resume the UI when you exit the editor

### GUI Editors
These editors run in separate windows:
- **code** (VS Code)
- **subl** (Sublime Text)
- **atom**
- **gedit**
- Any other GUI application

GUI editors are launched in the background, so Xplorer remains running.

## Example Configurations

### VS Code User (macOS)
```json
{
  "editor_cmd": "code",
  "terminal_app": "iTerm"
}
```

### Vim User (Linux) - Default
```json
{
  "editor_cmd": "vim",
  "terminal_app": "gnome-terminal"
}
```

### Sublime Text User (Windows)
```json
{
  "editor_cmd": "subl",
  "terminal_app": "cmd"
}
```

### Emacs User
```json
{
  "editor_cmd": "emacs",
  "terminal_app": "x-terminal-emulator"
}
```

## Creating the Config File

### Quick Setup

```bash
# Create config file with your preferred editor
cat > ~/.xplorer_config.json << EOF
{
  "editor_cmd": "code",
  "terminal_app": "iTerm"
}
EOF
```

### Manual Setup

1. Open your text editor
2. Create a new file at `~/.xplorer_config.json`
3. Add your configuration in JSON format
4. Save the file
5. Restart Xplorer

## Verifying Configuration

To check which editor Xplorer will use:

1. The config file location is: `~/.xplorer_config.json`
2. Check environment variables: `echo $EDITOR_CMD`
3. Platform defaults are used if neither exists

## Troubleshooting

### Editor Not Opening

- Verify the command is in your PATH: `which code`
- Check the config file syntax is valid JSON
- Ensure the editor command is correct for your system

### Terminal Not Opening

- Verify the terminal app is installed
- Check the app name matches your system (case-sensitive on Unix)
- Try using the full path to the terminal application

## Advanced Configuration

### Using Full Paths

If your editor isn't in PATH, use the full path:

```json
{
  "editor_cmd": "/usr/local/bin/code",
  "terminal_app": "/Applications/iTerm.app/Contents/MacOS/iTerm2"
}
```

### Editor with Arguments

You can include command-line arguments:

```json
{
  "editor_cmd": "code --new-window",
  "terminal_app": "gnome-terminal --"
}
```

## Other Configuration Files

Xplorer uses additional config files for other features:

- `~/.xplorer_theme` - Selected theme name
- `~/.xplorer_bookmarks.json` - Saved bookmarks
- `~/.xplorer_config.json` - Editor and terminal settings (this file)

## See Also

- [ARCHITECTURE.md](ARCHITECTURE.md) - Project architecture
- [FEATURES.md](FEATURES.md) - Feature documentation
- [TESTING.md](TESTING.md) - Testing guide