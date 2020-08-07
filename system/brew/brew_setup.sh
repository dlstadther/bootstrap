# install brew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"

# symbolic link brewfile
ln -s $(dirname $0)/.Brewfile $HOME/.Brewfile

# brew install
brew bundle install --global --verbose --no-upgrade
