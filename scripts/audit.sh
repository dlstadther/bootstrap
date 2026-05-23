#!/usr/bin/env bash
# Audits the gap between this repo and the current machine:
#   1. Dotfile symlink status (shared + host-specific)
#   2. Brew package drift (repo Brewfile vs live machine state)

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOTFILES_DIR="$REPO_DIR/dotfiles"
HOSTS_DIR="$REPO_DIR/hosts"
MACHINE="$(hostname -s)"

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

section() { echo -e "\n${BOLD}${CYAN}=== $* ===${RESET}"; }
ok()      { echo -e "  ${GREEN}OK${RESET}        $1"; }
warn()    { echo -e "  ${YELLOW}DIFF${RESET}      $1"; }
missing() { echo -e "  ${RED}MISSING${RESET}   $1"; }
realfile(){ echo -e "  ${YELLOW}REAL FILE${RESET} $1"; }
foreign() { echo -e "  ${RED}FOREIGN${RESET}   $1  ->  $2"; }

check_dir() {
  local dir="$1"
  local prefix="$2"
  local label="$3"
  local has_output=0

  while IFS= read -r -d '' src; do
    local rel="${src#$prefix/}"
    local target="$HOME/$rel"

    if [[ -L "$target" ]]; then
      local link_dest
      link_dest="$(readlink "$target")"
      if [[ "$link_dest" == "$src" ]]; then
        ok "$rel"
      else
        foreign "$rel" "$link_dest"
      fi
    elif [[ -f "$target" ]]; then
      if diff -q "$src" "$target" &>/dev/null; then
        realfile "$rel (identical content, not symlinked)"
      else
        realfile "$rel"
        { diff --unified=2 "$src" "$target" 2>/dev/null || true; } | tail -n +4 | head -20 | sed 's/^/    /' || true
        warn "^ differs from repo"
      fi
    else
      missing "$rel"
    fi
    has_output=1
  done < <(find "$dir" -not -name '.gitkeep' -type f -print0 | sort -z)

  if [[ $has_output -eq 0 ]]; then
    echo "  (no files)"
  fi
}

# ── Dotfiles ────────────────────────────────────────────────────────────────
section "Shared dotfiles  ($DOTFILES_DIR)"
check_dir "$DOTFILES_DIR" "$DOTFILES_DIR" "shared"

if [[ -d "$HOSTS_DIR/$MACHINE" ]]; then
  section "Host-specific dotfiles  (hosts/$MACHINE)"
  check_dir "$HOSTS_DIR/$MACHINE" "$HOSTS_DIR/$MACHINE" "host"
else
  section "Host-specific dotfiles"
  echo -e "  ${YELLOW}No host directory for '$MACHINE' (hosts/$MACHINE not found)${RESET}"
fi

# ── Brew ────────────────────────────────────────────────────────────────────
section "Brew package drift"

# Prefer host Brewfile, fall back to shared
brewfile_src=""
if [[ -f "$HOSTS_DIR/$MACHINE/.Brewfile" ]]; then
  brewfile_src="$HOSTS_DIR/$MACHINE/.Brewfile"
  echo "  Using host Brewfile: hosts/$MACHINE/.Brewfile"
elif [[ -f "$DOTFILES_DIR/.Brewfile" ]]; then
  brewfile_src="$DOTFILES_DIR/.Brewfile"
  echo "  Using shared Brewfile: dotfiles/.Brewfile"
else
  echo -e "  ${RED}No Brewfile found in repo${RESET}"
  exit 1
fi

tmp_machine="$(mktemp)"
trap 'rm -f "$tmp_machine"' EXIT
HOMEBREW_NO_AUTO_UPDATE=1 brew bundle dump --force --file="$tmp_machine" 2>/dev/null

repo_lines="$(grep -E '^(brew|cask|tap|mas|vscode|npm|go|uv) ' "$brewfile_src" | sort)"
machine_lines="$(grep -E '^(brew|cask|tap|mas|vscode|npm|go|uv) ' "$tmp_machine" | sort)"

echo ""
echo -e "  ${BOLD}In repo but NOT installed on machine:${RESET}"
in_repo_only="$(comm -23 <(echo "$repo_lines") <(echo "$machine_lines"))"
if [[ -n "$in_repo_only" ]]; then
  echo "$in_repo_only" | sed 's/^/    /'
else
  echo "    (none)"
fi

echo ""
echo -e "  ${BOLD}Installed on machine but NOT in repo:${RESET}"
on_machine_only="$(comm -13 <(echo "$repo_lines") <(echo "$machine_lines"))"
if [[ -n "$on_machine_only" ]]; then
  echo "$on_machine_only" | sed 's/^/    /'
else
  echo "    (none)"
fi

echo ""
