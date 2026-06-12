#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOST="$(hostname -s)"

register_marketplaces() {
  local f="$1"
  [ -f "$f" ] || return 0
  while IFS= read -r src; do
    echo "registering marketplace: $src"
    claude plugin marketplace add "$src" 2>/dev/null || true
  done < <(jq -r 'to_entries[] | .value' "$f")
}

register_marketplaces "$REPO_DIR/dotfiles/.claude/plugin-marketplaces.json"
register_marketplaces "$REPO_DIR/hosts/$HOST/.claude/plugin-marketplaces.json"

echo "installing enabled plugins..."
jq -r '.enabledPlugins | to_entries[] | select(.value == true) | .key' \
  "$HOME/.claude/settings.json" | while IFS= read -r plugin; do
  echo "installing plugin: $plugin"
  claude plugin install "$plugin" || true
done
