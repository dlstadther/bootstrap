#!/usr/bin/env bash
# MB-HYYXP4CRFJ tmux session bootstrap.
# Run via: tmux-bootstrap  or  tmux-start

ensure_window() {
  local session="$1" window="$2" dir="$3" cmd="$4"
  if ! tmux has-session -t "$session" 2>/dev/null; then
    tmux new-session -d -s "$session" -n "$window" -c "$dir"
    [[ -n "$cmd" ]] && tmux send-keys -t "$session:$window" "$cmd" Enter
    return
  fi
  if ! tmux list-windows -t "$session" -F "#{window_name}" 2>/dev/null | grep -qx "$window"; then
    tmux new-window -t "$session" -n "$window" -c "$dir"
    [[ -n "$cmd" ]] && tmux send-keys -t "$session:$window" "$cmd" Enter
  fi
}

# admin: always-on services and tools
ensure_window "admin" "agentsview" "$HOME"                "agentsview update && agentsview serve"
ensure_window "admin" "middleman"  "$HOME/code/middleman" "git pull origin main && make install && middleman"

# bootstrap: repo workspace — left pane (60%) + top-right + bottom-right
BOOTSTRAP_DIR="$HOME/code/bootstrap"
if ! tmux has-session -t "bootstrap" 2>/dev/null; then
  tmux new-session  -d  -s "bootstrap" -n "main" -c "$BOOTSTRAP_DIR"
  tmux send-keys    -t  "bootstrap:main" "claude agents --cwd $BOOTSTRAP_DIR"
  tmux split-window -h  -p 40 -t "bootstrap:main" -c "$BOOTSTRAP_DIR"
  tmux send-keys    -t  "bootstrap:main" "git pull origin main && ls -al && bd ready" Enter
  tmux split-window -v  -t "bootstrap:main" -c "$BOOTSTRAP_DIR"
  tmux send-keys    -t  "bootstrap:main" "lazygit" Enter
  tmux select-pane  -t  "bootstrap:main" -L
fi
