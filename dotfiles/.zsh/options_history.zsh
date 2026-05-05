# Zsh History
HISTFILE=~/.zsh_history
SAVEHIST=100000
HISTSIZE=25000
# Add epoch timestamp and elapsed time of command
setopt EXTENDED_HISTORY
# Share history across multiple zsh sessions
setopt SHARE_HISTORY
# Append to history
setopt APPEND_HISTORY
# Adds commands as they are typed, not at shell exit
setopt INC_APPEND_HISTORY
# Expire duplicates first
setopt HIST_EXPIRE_DUPS_FIRST
# Do not store duplications
setopt HIST_IGNORE_DUPS
# Ignore duplicates when searching
setopt HIST_FIND_NO_DUPS
# Removes blank lines from history
setopt HIST_REDUCE_BLANKS
# Show history substitution command before submitting
setopt HIST_VERIFY
