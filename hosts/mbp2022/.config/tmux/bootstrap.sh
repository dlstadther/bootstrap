#!/usr/bin/env bash
# mbp2022 tmux session bootstrap.
# Run via: tmux-bootstrap  or  tmux-start

# Example: monitoring session
if ! tmux has-session -t "monitor" 2>/dev/null; then
  tmux new-session -d -s "monitor" -c "$HOME"
  tmux rename-window -t "monitor:1" "sys"
fi
