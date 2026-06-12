#!/usr/bin/env bash
set -euo pipefail

jq -r '.enabledPlugins | to_entries[] | select(.value == true) | .key' \
  "$HOME/.claude/settings.json" | while IFS= read -r plugin; do
  echo "installing plugin: $plugin"
  claude plugin install "$plugin" || true
done
