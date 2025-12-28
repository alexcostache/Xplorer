#!/bin/bash

set -e

XP_NAME="xp"

# Detect platform
OS="$(uname)"
case "$OS" in
    Linux*)     PLATFORM="linux"; EXT=""; INSTALL_DIR="$HOME/.local/bin";;
    Darwin*)    PLATFORM="darwin"; EXT=""; INSTALL_DIR="$HOME/.local/bin";;
    CYGWIN*|MINGW*|MSYS*) PLATFORM="windows"; EXT=".exe"; INSTALL_DIR="$HOME/.local/bin";;
    *)          echo "❌ Unsupported OS: $OS"; exit 1;;
esac

# Ensure install directory exists
mkdir -p "$INSTALL_DIR"

echo "🔧 Building '$XP_NAME' for $PLATFORM..."
GOOS="$PLATFORM" GOARCH=amd64 go build -o "$XP_NAME$EXT"

echo ""
read -p "🌍 Install globally (requires sudo, not recommended on Windows)? [y/N]: " choice
if [[ "$choice" =~ ^[Yy]$ ]]; then
    if [[ "$PLATFORM" == "windows" ]]; then
        echo "⚠️ Global install not supported via script on Windows. Installing locally to $INSTALL_DIR..."
        mv "$XP_NAME$EXT" "$INSTALL_DIR/"
    else
        echo "📦 Installing globally to /usr/local/bin..."
        sudo mv "$XP_NAME$EXT" /usr/local/bin/
        echo "✅ Installed globally. You can now run '$XP_NAME' from anywhere."
    fi
else
    echo "📦 Installing to $INSTALL_DIR..."
    mv "$XP_NAME$EXT" "$INSTALL_DIR/"

    # Add to PATH if not already there (Linux/macOS only)
    if [[ "$PLATFORM" != "windows" ]]; then
        SHELL_CONFIG=""
        if [ -n "$ZSH_VERSION" ]; then SHELL_CONFIG="$HOME/.zshrc"; fi
        if [ -n "$BASH_VERSION" ]; then SHELL_CONFIG="$HOME/.bashrc"; fi
        if [ -z "$SHELL_CONFIG" ]; then SHELL_CONFIG="$HOME/.profile"; fi

        if ! grep -q "$INSTALL_DIR" "$SHELL_CONFIG"; then
            echo "🔧 Adding $INSTALL_DIR to PATH in $SHELL_CONFIG..."
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_CONFIG"
            echo "✅ Added to PATH. Restart your terminal or run: source $SHELL_CONFIG"
        fi
    else
        echo "⚠️ On Windows, ensure '$INSTALL_DIR' is in your PATH manually."
    fi

    echo "✅ Installed to $INSTALL_DIR. You can now run '$XP_NAME'."
fi
