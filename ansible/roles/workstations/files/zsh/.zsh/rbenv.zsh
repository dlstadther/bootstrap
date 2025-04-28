# set envs
export RBENV_ROOT="$HOME/.rbenv"
export PATH="$RBENV_ROOT/bin:$PATH"

# add 'rbenv init' to shell to enable shims and autocompletion
if command -v rbenv 1>/dev/null 2>&1; then
  eval "$(rbenv init --path)"
fi
