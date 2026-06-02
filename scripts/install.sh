#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOTFILES_DIR="$REPO_DIR/dotfiles"
HOSTS_DIR="$REPO_DIR/hosts"

machine="$(hostname -s)"

link_file() {
  local src="$1"
  local target="$2"

  # Already a symlink pointing into this repo — skip
  if [[ -L "$target" && "$(readlink "$target")" == "$src" ]]; then
    return
  fi

  # Existing file or foreign symlink — back it up
  if [[ -e "$target" || -L "$target" ]]; then
    local backup="${target}.bak.$(date +%Y%m%d%H%M%S)"
    mv "$target" "$backup"
    echo "backed up $target → $backup"
    # Prune old backups, keep newest 3
    local bak_dir bak_base bak_count
    bak_dir="$(dirname "$target")"
    bak_base="$(basename "$target")"
    bak_count=$(find "$bak_dir" -maxdepth 1 -name "${bak_base}.bak.*" 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$bak_count" -gt 3 ]]; then
      while IFS= read -r old_backup; do
        rm -- "$old_backup"
        echo "pruned old backup $old_backup"
      done < <(find "$bak_dir" -maxdepth 1 -name "${bak_base}.bak.*" 2>/dev/null | sort | head -n $(( bak_count - 3 )))
    fi
  fi

  mkdir -p "$(dirname "$target")"
  ln -s "$src" "$target"
  echo "linked $target → $src"
}

install_dir() {
  local dir="$1"
  local prefix="$2"

  while IFS= read -r -d '' src; do
    local rel="${src#$prefix/}"
    local target="$HOME/$rel"
    link_file "$src" "$target"
  done < <(find "$dir" -not -name '.gitkeep' -type f -print0)
}

# Shared dotfiles
install_dir "$DOTFILES_DIR" "$DOTFILES_DIR"

# Host-specific overrides (silent no-op if directory absent)
if [[ -d "$HOSTS_DIR/$machine" ]]; then
  install_dir "$HOSTS_DIR/$machine" "$HOSTS_DIR/$machine"
fi
