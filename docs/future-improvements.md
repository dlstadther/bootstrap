# Future Improvements

Ideas that aren't worth implementing now but are worth revisiting.

## Claude Code notifications in tmux

When Claude Code finishes a task or needs attention, surface a notification inside the active tmux session rather than relying on a separate notification channel.

**Rough approach:**

- Wire up Claude Code's `Stop` and `Notification` hooks to a small shell script.
- The script detects the originating tmux session/pane (e.g. via `$TMUX_PANE` or by querying `tmux display-message`).
- Write a flag file or update a tmux variable that the status line polls.
- The status line renders a visual indicator (e.g. a bell icon) on the pane that fired the event.
- Optionally also send a macOS `terminal-notifier` notification whose click action focuses that pane.

**Why it's deferred:** only useful inside a tmux workflow; adds a dependency on `terminal-notifier`; the flag/status-line polling approach needs care to avoid stale state when Claude Code exits uncleanly.
