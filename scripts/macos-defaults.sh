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

# ── Desktop Services ────────────────────────────────────────────────────────
defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true
defaults write com.apple.desktopservices DSDontWriteUSBStores -bool true

# ── Trackpad ────────────────────────────────────────────────────────────────
defaults write NSGlobalDomain com.apple.trackpad.scaling -float 3

# ── Dock ────────────────────────────────────────────────────────────────────
defaults write com.apple.dock tilesize -int 24
defaults write com.apple.dock show-recents -bool false
defaults write com.apple.dock orientation -string "left"

# ── Appearance ──────────────────────────────────────────────────────────────
defaults write NSGlobalDomain AppleInterfaceStyle -string "Dark"
defaults write NSGlobalDomain AppleICUForce24HourTime -bool true

# ── Keyboard ────────────────────────────────────────────────────────────────
# Key repeat: lower = faster. System minimum exposed in UI is 2.
defaults write NSGlobalDomain KeyRepeat -int 2
defaults write NSGlobalDomain InitialKeyRepeat -int 15
# Remap Caps Lock → Escape for current session; LaunchAgent handles persistence on login.
hidutil property --set '{"UserKeyMapping":[{"HIDKeyboardModifierMappingSrc":0x700000039,"HIDKeyboardModifierMappingDst":0x700000029}]}'

# ── Power Management ────────────────────────────────────────────────────────
# pmset requires sudo — will prompt if not already elevated.
sudo pmset -b displaysleep 60  # battery: sleep display after 60 min
sudo pmset -c displaysleep 0   # AC adapter: never sleep display

# ── Restart affected processes ───────────────────────────────────────────────
killall Finder          2>/dev/null || true
killall Dock            2>/dev/null || true
killall SystemUIServer  2>/dev/null || true

echo "Done. Keyboard repeat and some appearance changes require re-login to take effect."
