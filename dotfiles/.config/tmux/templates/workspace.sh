#!/usr/bin/env bash
# Usage: workspace.sh --cwd <path> [--name <window-name>] [--agent <agent>]
#
# --cwd     Required. Working directory for all panes.
# --name    Optional. Window name override (session name derives from basename of --cwd).
# --agent   Optional. Default: claude. One of: claude, codex, gemini, opencode, pi.
#
# Behaviour:
#   New session     → create session + window named from basename(cwd) (or --name override)
#   Existing session → add new window; auto-increment name if no --name given
#   Always focuses the newly created window.

set -euo pipefail

ALLOWED_AGENTS=(claude codex gemini opencode pi)

usage() {
  cat >&2 <<EOF
Usage: workspace.sh --cwd <path> [--name <window-name>] [--agent <agent>]

  --cwd     Required. Working directory for all panes.
  --name    Optional. Window name override.
  --agent   Optional. Agent to stage in left pane. Default: claude.
            Allowed: ${ALLOWED_AGENTS[*]}
EOF
  exit 1
}

CWD=""
NAME_OVERRIDE=""
AGENT="claude"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --cwd)    CWD="$2";           shift 2 ;;
    --name)   NAME_OVERRIDE="$2"; shift 2 ;;
    --agent)  AGENT="$2";         shift 2 ;;
    -h|--help) usage ;;
    *) echo "error: unknown argument: $1" >&2; usage ;;
  esac
done

if [[ -z "$CWD" ]]; then
  echo "error: --cwd is required" >&2
  usage
fi

TARGET_DIR="$(realpath "$CWD")"

if [[ ! -d "$TARGET_DIR" ]]; then
  echo "error: directory does not exist: $TARGET_DIR" >&2
  exit 1
fi

valid_agent=0
for a in "${ALLOWED_AGENTS[@]}"; do
  [[ "$AGENT" == "$a" ]] && valid_agent=1 && break
done
if [[ $valid_agent -eq 0 ]]; then
  echo "error: invalid agent '${AGENT}'. Allowed: ${ALLOWED_AGENTS[*]}" >&2
  exit 1
fi

if ! tmux info &>/dev/null; then
  echo "error: tmux is not running" >&2
  exit 1
fi

DIRNAME="$(basename "$TARGET_DIR")"
SESSION_NAME="$DIRNAME"

next_window_name() {
  local session="$1"
  local base="$2"
  local candidate="$base"
  local n=2
  while tmux list-windows -t "$session" -F "#{window_name}" 2>/dev/null \
      | grep -qx "$candidate"; do
    candidate="${base}-${n}"
    (( n++ ))
  done
  echo "$candidate"
}

create_panes() {
  local session="$1"
  local window="$2"
  local dir="$3"

  # Split right column (~40%)
  tmux split-window -h -p 40 -t "${session}:${window}" -c "$dir"

  # Split right column: top for git/bd, bottom for lazygit
  tmux split-window -v -t "${session}:${window}" -c "$dir"
  tmux send-keys -t "${session}:${window}" "git pull origin main && ls -al && bd ready" Enter

  tmux split-window -v -t "${session}:${window}" -c "$dir"
  tmux send-keys -t "${session}:${window}" "lazygit" Enter

  # Left pane: stage agent command (no Enter — leaves user in control)
  tmux select-pane -t "${session}:${window}.0"
  tmux send-keys -t "${session}:${window}.0" "$AGENT"
}

if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
  WINDOW_NAME="${NAME_OVERRIDE:-$DIRNAME}"
  tmux new-session -d -s "$SESSION_NAME" -n "$WINDOW_NAME" -c "$TARGET_DIR"
  create_panes "$SESSION_NAME" "$WINDOW_NAME" "$TARGET_DIR"
  tmux switch-client -t "$SESSION_NAME"
else
  WINDOW_NAME="${NAME_OVERRIDE:-$(next_window_name "$SESSION_NAME" "$DIRNAME")}"
  tmux new-window -t "$SESSION_NAME" -n "$WINDOW_NAME" -c "$TARGET_DIR"
  create_panes "$SESSION_NAME" "$WINDOW_NAME" "$TARGET_DIR"
  tmux select-window -t "${SESSION_NAME}:${WINDOW_NAME}"
fi
