#!/usr/bin/env bash
# mbp2022 tmux session bootstrap.
# Run via: tmux-bootstrap  or  tmux-start

window_exists() {
  tmux list-windows -t "$1" -F "#{window_name}" 2>/dev/null | grep -qx "$2"
}

ensure_window() {
  local session="$1" window="$2" dir="$3" cmd="$4"
  if ! window_exists "$session" "$window"; then
    tmux new-window -t "$session" -n "$window" -c "$dir"
    [[ -n "$cmd" ]] && tmux send-keys -t "$session:$window" "$cmd" Enter
  fi
}

# admin: always-on services and tools
if ! tmux has-session -t "admin" 2>/dev/null; then
  tmux new-session -d -s "admin" -n "agentsview" -c "$HOME"
  tmux send-keys -t "admin:agentsview" "agentsview update && agentsview serve" Enter
fi

ensure_window "admin" "agentsview" "$HOME" "agentsview update && agentsview serve"
ensure_window "admin" "middleman" "$HOME/code/middleman" "git pull origin main && make install && middleman"
