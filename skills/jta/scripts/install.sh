#!/bin/bash
# Jta Installation Script
# Automatically detects OS and architecture, installs jta

set -e  # Exit on error

echo "🚀 Jta Installation Script"
echo "=========================="
echo ""

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

echo "Detected system: $OS $ARCH"
echo ""

# Check if jta is already installed
if command -v jta &> /dev/null; then
    CURRENT_VERSION=$(jta --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    echo "✓ Jta is already installed (version: $CURRENT_VERSION)"
    read -p "Do you want to reinstall/update? [y/N] " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
fi

# Install based on OS
case "$OS" in
    Darwin*)
        echo "📦 Installing on macOS..."

        # Try Homebrew first
        if command -v brew &> /dev/null; then
            echo "Using Homebrew..."
            brew tap hikanner/jta 2>/dev/null || true
            brew install jta || brew upgrade jta
        else
            echo "Homebrew not found, downloading binary..."

            # Determine architecture
            if [[ "$ARCH" == "arm64" ]]; then
                BINARY="jta-darwin-arm64"
            else
                BINARY="jta-darwin-amd64"
            fi

            # Download binary
            echo "Downloading $BINARY..."
            curl -L "https://github.com/hikanner/jta/releases/latest/download/$BINARY" -o jta

            # Make executable
            chmod +x jta

            # Move to PATH
            echo "Installing to /usr/local/bin/ (may require sudo password)..."
            sudo mv jta /usr/local/bin/
        fi
        ;;

    Linux*)
        echo "📦 Installing on Linux..."

        # Determine architecture
        if [[ "$ARCH" == "x86_64" ]]; then
            BINARY="jta-linux-amd64"
        elif [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
            BINARY="jta-linux-arm64"
        else
            echo "❌ Unsupported architecture: $ARCH"
            exit 1
        fi

        # Download binary
        echo "Downloading $BINARY..."
        curl -L "https://github.com/hikanner/jta/releases/latest/download/$BINARY" -o jta

        # Make executable
        chmod +x jta

        # Move to PATH
        echo "Installing to /usr/local/bin/ (may require sudo password)..."
        sudo mv jta /usr/local/bin/
        ;;

    *)
        echo "❌ Unsupported operating system: $OS"
        echo "Please visit https://github.com/hikanner/jta for manual installation instructions."
        exit 1
        ;;
esac

echo ""
echo "✅ Installation complete!"
echo ""

# Verify installation
if command -v jta &> /dev/null; then
    VERSION=$(jta --version 2>&1 | head -1)
    echo "Installed version: $VERSION"
    echo ""
    echo "🎉 You can now use jta!"
    echo ""
    echo "Quick start:"
    echo "  jta en.json --to zh,ja,ko"
    echo ""
    echo "For help:"
    echo "  jta --help"
    echo ""
    echo "⚠️  Don't forget to set your API key:"
    echo "  export OPENAI_API_KEY=sk-..."
else
    echo "❌ Installation verification failed"
    echo "Please check the error messages above and try again."
    exit 1
fi
