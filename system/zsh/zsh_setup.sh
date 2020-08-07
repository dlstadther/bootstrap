# install zsh, change to zsh, configure zsh

# add option to also install zsh
# if darwin, brew install zsh
# if ubuntu, apt install zsh -y

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

# source "$HOME/.zshrc"
