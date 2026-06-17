export GITHUB_TOKEN="$(gh auth token 2>/dev/null)"
eval "$(command mise activate zsh)"
# activate only updates PATH on interactive cd; non-interactive subshells (make, scripts)
# and dirs it didn't re-eval get no tool. Shims read .tool-versions/.python-version at exec
# time, so they resolve the right version everywhere — incl. subprojects that pin a different one.
export PATH="$HOME/.local/share/mise/shims:$PATH"
