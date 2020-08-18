

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
#echo "Switching default shell to zsh ..."
chsh -s `which zsh`

