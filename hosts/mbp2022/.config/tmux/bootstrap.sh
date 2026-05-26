#!/usr/bin/env bash
# mbp2022 tmux session bootstrap.
# Run via: tmux-bootstrap  or  tmux-start

ensure_window() {
  local session="$1" window="$2" dir="$3" cmd="$4"

  if ! tmux has-session -t "$session" 2>/dev/null; then
    tmux new-session -d -s "$session" -n "$window" -c "$dir"
    sleep 1
  elif ! tmux list-windows -t "$session" -F "#{window_name}" 2>/dev/null | grep -qx "$window"; then
    tmux new-window -t "$session" -n "$window" -c "$dir"
    sleep 1
  fi

  # Always (re-)send the command — handles both fresh windows and resurrect-restored ones
  if [[ -n "$cmd" ]]; then
    tmux send-keys -t "$session:$window" C-c
    sleep 0.3
    tmux send-keys -t "$session:$window" "$cmd" Enter
  fi
}

# admin: always-on services and tools
ensure_window "admin" "agentsview" "$HOME"               "agentsview update --yes && agentsview serve --no-browser"
ensure_window "admin" "middleman"  "$HOME/code/middleman" "git pull origin main && make install && middleman"

# bootstrap: repo workspace — left pane (60%) + top-right + bottom-right
BOOTSTRAP_DIR="$HOME/code/bootstrap"
if ! tmux has-session -t "bootstrap" 2>/dev/null; then
  tmux new-session  -d  -s "bootstrap" -n "main" -c "$BOOTSTRAP_DIR"
  sleep 1
  tmux split-window -h  -p 40 -t "bootstrap:main" -c "$BOOTSTRAP_DIR"
  sleep 1
  tmux send-keys    -t  "bootstrap:main" "git pull origin main && ls -al && bd ready" Enter
  tmux split-window -v  -t "bootstrap:main" -c "$BOOTSTRAP_DIR"
  sleep 1
  tmux send-keys    -t  "bootstrap:main" "lazygit" Enter
  tmux select-pane  -t  "bootstrap:main" -L
fi

# Always re-seed the pre-typed claude agents command in the left pane (pane 0),
# whether the session was just created or restored by resurrect.
tmux send-keys -t "bootstrap:main.0" C-c
sleep 0.3
tmux send-keys -t "bootstrap:main.0" "claude agents --cwd $BOOTSTRAP_DIR"
