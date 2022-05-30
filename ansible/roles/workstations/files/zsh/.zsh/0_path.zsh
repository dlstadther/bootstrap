# path, the 0 in the filename causes this to load first

pathPrefix() {
  # Add to the start of the path
  PATH=$1:$PATH
}

pathAppend() {
  # Only adds to the path if it's not already there
  if ! echo $PATH | egrep -q "(^|:)$1($|:)" ; then
    PATH=$PATH:$1
  fi
}

# Remove duplicate entries from PATH:
PATH=$(echo "$PATH" | awk -v RS=':' -v ORS=":" '!a[$1]++{if (NR > 1) printf ORS; printf $a[$1]}')

pathPrefix "$HOME/.pyenv/bin"  # pyenv
pathPrefix "$HOME/.cargo/bin"  # rust
pathPrefix "$(go env GOPATH)/bin"  # golang
pathPrefix "/home/linuxbrew/.linuxbrew/bin"  # homebrew
pathPrefix "$HOME/.local/bin"  # poetry

