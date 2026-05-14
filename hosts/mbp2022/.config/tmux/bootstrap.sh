#!/usr/bin/env bash
# mbp2022 tmux session bootstrap.
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
ensure_window "admin" "agentsview" "$HOME"               "agentsview update && agentsview serve"
ensure_window "admin" "middleman"  "$HOME/code/middleman" "git pull origin main && make install && middleman"
