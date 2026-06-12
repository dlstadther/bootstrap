.PHONY: install bootstrap macos-defaults diff diff-host brew-install brew-sync brew-dump init-tmux build-bs test-bs install-plugins sync

HOST ?= $(shell hostname -s)

BS_LDFLAGS = \
	-X github.com/dlstadther/bootstrap/cli/internal/version.CommitHash=$(shell git rev-parse HEAD) \
	-X github.com/dlstadther/bootstrap/cli/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
	-X github.com/dlstadther/bootstrap/cli/internal/version.RepoPath=$(HOME)/code/bootstrap

# Build the bs binary and install it to ~/.local/bin/bs
build-bs:
	cd cli && go build -ldflags "$(BS_LDFLAGS)" -o ~/.local/bin/bs .

# Run bs CLI unit tests
test-bs:
	cd cli && go test ./...

# Symlink shared dotfiles and host-specific overrides into ~/
install: build-bs
	./scripts/install.sh

# Full machine setup: dotfiles + macOS defaults
bootstrap: install macos-defaults

# Apply preferred macOS system defaults (Finder, Dock, keyboard)
macos-defaults:
	./scripts/macos-defaults.sh

diff:
	git diff

# Compare host-specific files against their dotfiles counterparts
# Usage: make diff-host [HOST=<hostname>]
diff-host:
	@found=0; \
	for f in $$(find hosts/$(HOST) -type f 2>/dev/null); do \
	  rel=$${f#hosts/$(HOST)/}; \
	  base="dotfiles/$$rel"; \
	  if [[ -f "$$base" ]]; then \
	    found=1; \
	    echo "=== $$rel ==="; \
	    diff -u "$$base" "$$f" || true; \
	  fi; \
	done; \
	if [[ $$found -eq 0 ]]; then echo "No overlapping files found for host: $(HOST)"; fi

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

install-plugins:
	./scripts/install-plugins.sh

sync: install
	bs sync
