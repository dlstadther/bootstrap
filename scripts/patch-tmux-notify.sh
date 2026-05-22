#!/usr/bin/env bash
# Patch tmux-notify to use terminal-notifier instead of osascript.
# Must be re-run after TPM plugin updates (prefix+U).
set -euo pipefail

HELPERS="$HOME/.config/tmux/plugins/tmux-notify/scripts/helpers.sh"

if [[ ! -f "$HELPERS" ]]; then
  echo "tmux-notify not installed — skipping patch"
  exit 0
fi

if grep -q "terminal-notifier" "$HELPERS"; then
  echo "tmux-notify already patched — skipping"
  exit 0
fi

python3 - "$HELPERS" <<'PYEOF'
import sys

path = sys.argv[1]
with open(path) as f:
    content = f.read()

OLD = (
    "  if [[ \"$OSTYPE\" =~ ^darwin ]]; then # If macOS\n"
    "    if [ -n \"$2\" ]; then\n"
    "      osascript -e 'display notification \"'\"$1\"'\" with title \"'\"$2\"'\"'\n"
    "    else\n"
    "      osascript -e 'display notification \"'\"$1\"'\" with title \"tmux-notify\"'\n"
    "    fi"
)

NEW = (
    "  if [[ \"$OSTYPE\" =~ ^darwin ]]; then # If macOS\n"
    "    local _title=\"${2:-tmux-notify}\"\n"
    "    local _exec=\"open -a Ghostty; tmux switch-client -t '\\$${SESSION_ID}'; tmux select-window -t '@${WINDOW_ID}'; tmux select-pane -t '%${PANE_ID}'\"\n"
    "    terminal-notifier -title \"$_title\" -message \"$1\" -sender com.mitchellh.ghostty -execute \"$_exec\" 2>/dev/null \\\n"
    "      || osascript -e 'display notification \"'\"$1\"'\" with title \"'\"$_title\"'\"'"
)

if OLD in content:
    content = content.replace(OLD, NEW)
    with open(path, "w") as f:
        f.write(content)
    print("tmux-notify patched successfully")
else:
    print("WARNING: patch target not found — plugin may have changed upstream")
    sys.exit(1)
PYEOF
