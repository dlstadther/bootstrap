#!/usr/bin/zsh

# Get operating system
platform='unknown'
unamestr=$(uname)
if [[ $unamestr == 'Linux' ]]; then
  platform='linux'
elif [[ $unamestr == 'Darwin' ]]; then
  platform='darwin'
fi
echo "OS: $platform"

ABSOLUTESCRIPTPATH=${0:a:h}
echo "Executing with Absolute Script Path: $ABSOLUTESCRIPTPATH ..."

# install brew if not installed
# update brew if installed
brew -v
if [[ $? != 0 ]]; then
    echo "Brew not installed; Installing ..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"
else
    echo "Brew is already installed; Doctoring ..."
    brew doctor
fi

# symbolic link brewfile
# rename old one if exists
if [[ $platform == 'linux' ]]; then
    NAME=".Brewfile-linux"
elif [[ $platform == 'darwin' ]]; then
    NAME=".Brewfile-mac"
else
    echo ""
fi
SRC="$ABSOLUTESCRIPTPATH/$NAME"
DST="$HOME/.Brewfile"
MODIFIEDDATE=$(date +"%Y%m%dT%H%M%S")
if [[ -f $DST ]]; then
    RENAMED="$DST.bak.$MODIFIEDDATE"
    echo "Backing up $DST to $RENAMED ..."
    mv "$DST" "$RENAMED"
else
    echo "$DST did not exist ..."
fi

echo "Symbolic linking $SRC to $DST"
ln -s $SRC $DST

# brew install
#brew bundle install --global --verbose --no-upgrade

