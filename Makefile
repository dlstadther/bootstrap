.PHONY: install bootstrap macos-defaults diff brew-install brew-sync brew-dump init-tmux

# Symlink shared dotfiles and host-specific overrides into ~/
install:
	./install.sh

# Full machine setup: dotfiles + macOS defaults
bootstrap: install macos-defaults

# Apply preferred macOS system defaults (Finder, Dock, keyboard)
macos-defaults:
	./scripts/macos-defaults.sh

diff:
	git diff

brew-install:
	brew update
	brew bundle install --global --verbose

brew-sync:
	brew bundle dump --force --file=/tmp/.Brewfile.current
	diff dotfiles/.Brewfile /tmp/.Brewfile.current || true

brew-dump:
	brew bundle dump --force --file=dotfiles/.Brewfile

init-tmux:
	./scripts/init-tmux.sh
