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
echo "Updating tmux plugins..."
"$TPM_DIR/bin/update_plugins" all
echo "Plugins up to date."

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "Applying tmux plugin patches..."
bash "$SCRIPT_DIR/patch-tmux-notify.sh"

echo "Reloading tmux config (if server is running)..."
tmux source-file ~/.config/tmux/tmux.conf 2>/dev/null && echo "Config reloaded." || echo "No tmux server running — reload manually with prefix+r."
