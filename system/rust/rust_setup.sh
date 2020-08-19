# install Rust
echo "Installing Rust ..."
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
  | sh -s -- -y --default-toolchain nightly
# cargo path already added to zsh config

# Load Rust env
echo "Loading Rust Environment ..."
source $HOME/.cargo/env

echo "Installing pinned Rust toolchain ..."
rustup toolchain install "nightly-2019-11-04"

echo "Installing source for pinned Rust toolchain ..."
rustup component add --toolchain "nightly-2019-11-04" rust-src

