#!/usr/bin/env bash

input=$(cat)

# Extract values
cwd=$(echo "$input" | jq -r '.workspace.current_dir // .cwd // ""')
rel_path="${cwd/#$HOME/~}"

# Git branch
branch=""
if git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  branch=$(git -C "$cwd" --no-optional-locks symbolic-ref --short HEAD 2>/dev/null)
fi

# Date / time
dow=$(date "+%a")
datetime=$(date "+%Y-%m-%d %H:%M")

# Line 1: path (branch) 📁 | DOW YYYY-MM-DD HH:MM 🕐
if [ -n "$branch" ]; then
  line1="📁 $rel_path ($branch) | 🕐 $dow $datetime"
else
  line1="📁 $rel_path | 🕐 $dow $datetime"
fi

# --- Line 2 components ---

# Model
model=$(echo "$input" | jq -r '.model.display_name // "unknown"')

# Colored context bar
used_pct=$(echo "$input" | jq -r '.context_window.used_percentage // empty')
if [ -n "$used_pct" ]; then
  pct_int=${used_pct%.*}
  if [ "$pct_int" -ge 90 ]; then
    color="\033[31m"   # red
  elif [ "$pct_int" -ge 75 ]; then
    color="\033[33m"   # yellow
  else
    color="\033[32m"   # green
  fi
  reset="\033[0m"

  filled=$(( pct_int / 10 ))
  bar=""
  for i in $(seq 1 10); do
    if [ "$i" -le "$filled" ]; then
      bar="${bar}█"
    else
      bar="${bar}░"
    fi
  done
  ctx_display="${color}${bar}${reset} ${used_pct}%"
else
  ctx_display="░░░░░░░░░░ -"
fi

# Session duration from cost.total_duration_ms
total_ms=$(echo "$input" | jq -r '.cost.total_duration_ms // empty')
if [ -n "$total_ms" ]; then
  total_sec=$(( total_ms / 1000 ))
  hours=$(( total_sec / 3600 ))
  minutes=$(( (total_sec % 3600) / 60 ))
  if [ "$hours" -gt 0 ]; then
    duration_display="⏳ ${hours}h ${minutes}m"
  else
    duration_display="⏳ ${minutes}m"
  fi
else
  duration_display="⏳ -"
fi

# Line 3: Session cost
session_cost=$(echo "$input" | jq -r '.cost.total_cost_usd // 0')
cost_display=$(printf '$%.2f' "$session_cost")

printf "%s\n%b\n%s" "$line1" "🤖 $model | $ctx_display | $duration_display" "💰 $cost_display (session)"
