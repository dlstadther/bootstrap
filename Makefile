.PHONY: install diff brew-install brew-sync brew-dump

install:
	./install.sh

diff:
	git diff

brew-install:
	brew bundle install --global

brew-sync:
	brew bundle dump --force --file=/tmp/.Brewfile.current
	diff dotfiles/.Brewfile /tmp/.Brewfile.current || true

brew-dump:
	brew bundle dump --force
