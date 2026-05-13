#!/usr/bin/env bash
# Apply preferred macOS system defaults.
# Some settings require re-login or restarting the affected process to take effect.
set -euo pipefail

echo "Applying macOS defaults..."

# ── Finder ──────────────────────────────────────────────────────────────────
defaults write com.apple.finder AppleShowAllFiles -bool true
defaults write NSGlobalDomain AppleShowAllExtensions -bool true
defaults write com.apple.finder ShowPathbar -bool true
defaults write com.apple.finder ShowStatusBar -bool true
defaults write com.apple.finder FXPreferredViewStyle -string "Nlsv"

# ── Dock ────────────────────────────────────────────────────────────────────
defaults write com.apple.dock tilesize -int 24
defaults write com.apple.dock show-recents -bool false

# ── Keyboard ────────────────────────────────────────────────────────────────
# Key repeat: lower = faster. System minimum exposed in UI is 2.
defaults write NSGlobalDomain KeyRepeat -int 2
defaults write NSGlobalDomain InitialKeyRepeat -int 15

# ── Restart affected processes ───────────────────────────────────────────────
killall Finder 2>/dev/null || true
killall Dock   2>/dev/null || true

echo "Done. Some settings (e.g. keyboard repeat) require re-login to take effect."
