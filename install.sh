#!/usr/bin/env bash
set -e

REPO="amine-khemissi/env-builder"   # update before distributing
BIN_DIR="$HOME/.local/bin"
EB_DIR="$HOME/.eb"

# Detect architecture
case $(uname -m) in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
esac

# Resolve latest release tag
VERSION=$(curl -sf "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | cut -d'"' -f4)

if [[ -z "$VERSION" ]]; then
  echo "Error: could not fetch latest release from $REPO"
  exit 1
fi

echo "Installing eb $VERSION ($ARCH)..."
mkdir -p "$BIN_DIR" "$EB_DIR"

echo "Downloading eb binary..."
curl -fL --progress-bar "https://github.com/$REPO/releases/download/$VERSION/eb-linux-$ARCH" \
  -o "$BIN_DIR/eb" || { echo "Error: could not download eb binary for $ARCH from release $VERSION"; exit 1; }
chmod +x "$BIN_DIR/eb"

echo "Downloading managers.yaml..."
curl -fL --progress-bar "https://github.com/$REPO/releases/download/$VERSION/managers.yaml" \
  -o "$EB_DIR/managers.yaml" || { echo "Error: could not download managers.yaml from release $VERSION"; exit 1; }

# Shell completions
if command -v fish >/dev/null 2>&1; then
  mkdir -p ~/.config/fish/completions
  "$BIN_DIR/eb" completion fish > ~/.config/fish/completions/eb.fish
fi
if command -v bash >/dev/null 2>&1; then
  mkdir -p ~/.local/share/bash-completion/completions
  "$BIN_DIR/eb" completion bash > ~/.local/share/bash-completion/completions/eb
fi
if command -v zsh >/dev/null 2>&1; then
  mkdir -p ~/.local/share/zsh/site-functions
  "$BIN_DIR/eb" completion zsh > ~/.local/share/zsh/site-functions/_eb
fi

echo ""
echo "  eb            -> $BIN_DIR/eb"
echo "  managers.yaml -> $EB_DIR/managers.yaml"
echo ""
echo "Next: run 'eb edit' to create and edit your config."
echo "Then: eb install"
