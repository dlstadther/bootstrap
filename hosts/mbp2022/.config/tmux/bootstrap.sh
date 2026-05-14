#!/usr/bin/env bash
# mbp2022 tmux session bootstrap.
# Run via: tmux-bootstrap  or  tmux-start

# Example: monitoring session
if ! tmux has-session -t "monitor" 2>/dev/null; then
  tmux new-session -d -s "monitor" -c "$HOME"
  tmux rename-window -t "monitor:1" "sys"
fi

if ! tmux has-session -t "agentsview" 2>/dev/null; then
  tmux new-session -d -s "agentsview" -c "$HOME"
  tmux rename-window -t "agentsview:1" "serve"
  tmux send-keys -t "agentsview:serve" "agentsview update && agentsview serve" Enter
fi

if ! tmux has-session -t "middleman" 2>/dev/null; then
  tmux new-session -d -s "middleman" -c "$HOME/code/middleman"
  tmux rename-window -t "middleman:1" "serve"
  tmux send-keys -t "middleman:serve" "git pull origin main && make install && middleman" Enter
fi
