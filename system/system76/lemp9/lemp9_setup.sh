#!/usr/bin/env zsh

set -e

# be sure sdcc and rust are installed
apt-get update
sudo apt-get install \
  build-essential \
  sdcc

if which rustup > /dev/null; then
  echo "Rust is already installed!"
else
  echo "Calling Rust Installation Script ..."
  # TODO: call script based on relative path
fi

# Update ec.git fork with upstream
CURDIR=$(PWD)
CODEDIR="$HOME/code"
REPO="ec"
MYBRANCH="ds"
USERORIGIN="dlstadther"
USERUPSTREAM="system76"

mkdir -p $CODEDIR
cd $CODEDIR
if [[ ! -d ./$REPO ]]; then
  echo "Cloning repo '$USERORIGIN/$REPO' ..."
  git clone git@github.com:$USERORIGIN/$REPO
else
  echo "Repo has already been cloned"
fi

cd $REPO
git fetch origin
git remote add upstream git@github.com:$USERUPSTREAM/$REPO
git fetch upstream

echo "Updating origin/master with upstream/master ..."
git checkout master
git pull origin master
git rebase upstream/master
git push origin master

echo "Syncing origin branch '$MYBRANCH' with upstream master ..."
git checkout $MYBRANCH
git rebase master
# git push origin $MYBRANCH --force

# ensure user's config.mk is correct
cat <EOF > ./config.mk
BOARD?=system76/lemp9
KEYMAP?=dillon
EOF

# run make to create firmware build based on latest commit
make
# print instructions to apply firmware update
echo "When ready to apply EC firmware update, run 'make flash_internal'"

cd $CURDIR

