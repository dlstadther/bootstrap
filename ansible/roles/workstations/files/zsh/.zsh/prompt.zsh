# vcs information
autoload -Uz vcs_info
precmd_vcs_info() { vcs_info }
precmd_functions+=( precmd_vcs_info )
setopt PROMPT_SUBST
zstyle ':vcs_info:git:*' formats '(%F{green}%r@%b%f) '
zstyle ':vcs_info:*' enable git

# [%*] - [hh:mm:ss] using 24 hr time
# %(?.√.?%?) - if return code ? is 0, show √, else show ?%?
# %? - exit code of previous command
# %4~ - cwd, shortening HOME to ~, show only last 4 elements
# %(!.#.>) - # with root privileges, > otherwise
# %B%b - start/stop bold
# %F{...} - text (foreground) color
# %f - reset to default textcolor
# ${vcs_info_msg_0_} - VCS info formatted as specified by zstyle above
PROMPT='%F{yellow}[%*] %(?.%F{green}√.%F{red}?%?)%f %B%F{240}%4~%f%b ${vcs_info_msg_0_}%(!.#.>) '

# sys time
#RPROMPT='%*'
#RPROMPT=''
