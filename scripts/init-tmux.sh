#!/usr/bin/env bash
set -euo pipefail

TPM_DIR="$HOME/.tmux/plugins/tpm"

if [[ -d "$TPM_DIR" ]]; then
  echo "TPM already installed — pulling latest"
  git -C "$TPM_DIR" pull --rebase
else
  git clone https://github.com/tmux-plugins/tpm "$TPM_DIR"
  echo "TPM installed."
fi

echo "Installing tmux plugins..."
"$TPM_DIR/bin/install_plugins"
echo "Plugins installed."
