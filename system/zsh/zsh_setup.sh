#!/usr/bin/zsh

# configure zsh

ABSOLUTESCRIPTPATH=${0:a:h}
echo "Executing with Absolute Script Path: $ABSOLUTESCRIPTPATH ..."

# install configs (back up locations if already exists)
back_up_if_exists() {
    SRCDIR=$1
    NAME=$2

    SRC="$SRCDIR/$NAME"
    DST="$HOME/$NAME"
    MODIFIEDDATE=$(date +"%Y%m%dT%H%M%S")

    # if DST exists (either as a dir or a file), suffix with .bak.yyyymmddthhmmss
    echo "Checking existence of '$DST' ..."
    if [[ -d "$DST" ]] || [[ -f "$DST" ]]; then
        RENAMED="$DST.bak.$MODIFIEDDATE"
        echo "Backing up $DST to $RENAMED ..."
        mv "$DST" "$RENAMED"
    else
        echo "$DST did not exist ..."
    fi

    # sym link
    echo "Symbolic linking $SRC to $DST"
    ln -s "$SRC" "$DST"
}

ZSHDIR="$ABSOLUTESCRIPTPATH"
back_up_if_exists $ZSHDIR .zsh
back_up_if_exists $ZSHDIR .zshrc.after
back_up_if_exists $ZSHDIR .zshrc.before
back_up_if_exists $ZSHDIR .zshrc

source "$HOME/.zshrc"

