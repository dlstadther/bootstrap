

# Get this scripts relative directory path
SCRIPTPATH=$(dirname "$0")

# Get operating system
platform='unknown'
unamestr=$(uname)
if [[ $unamestr == 'Linux' ]]; then
  platform='linux'
elif [[ $unamestr == 'Darwin' ]]; then
  platform='darwin'
fi
echo "OS: $platform"

# if darwin, xcode-select --install
if [[ $platform == 'darwin' ]]; then
    echo "Ensuring xcode dev tools are installed ..."
    # xcode-select --install
fi

# if brew, run brew/setup.sh

# add option to also install zsh
# TODO: ONLY INSTALL IF DNE - check existing shell and/or if zsh is already in /etc/shells
echo "Installing zsh ..."
if [[ $platform == 'darwin' ]]; then
    brew install zsh
elif [[ $platform == 'linux' ]]; then
    # if flavor == debian
    sudo apt install zsh -y
else
    echo "OS zsh installation process not configured ..."
fi

# switch to zsh
# chsh -s `which zsh`

# install configs (back up locations if already exists)
back_up_if_exists() {
    SRCDIR=$1
    NAME=$2

    SRC="$SRCDIR/$NAME"
    DST="$HOME/$NAME"
    MODIFIEDDATE=$(date +"%Y%m%dT%H%M%S")

    # if DST exists (either as a dir or a file), suffix with .bak.yyyymmddthhmmss
    if [[ -d "$DST" -o -f "$DST" ]]; then
        RENAMED="$DST.bak.$MODIFIEDDATE"
        echo "Backing up $DST to $RENAMED ..."
        # mv "$DST" "$RENAMED"
    fi

    # sym link
    echo "Symbolic linking $SRC to $DST"
    # ln -s "$SRC" "$DST"
}

ZSHDIR="$SCRIPTPATH/zsh"
back_up_if_exists $ZSHDIR .zsh
back_up_if_exists $ZSHDIR .zshrc.after
back_up_if_exists $ZSHDIR .zshrc.before
back_up_if_exists $ZSHDIR .zshrc

