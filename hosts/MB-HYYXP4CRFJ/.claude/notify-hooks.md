# Claude Code Notify Hooks (currently disabled)

The `Stop` and `Notification` hooks that fired `~/.local/bin/claude-notify`
have been **removed** from `settings.json`.

They were built for a tmux-based workflow: `claude-notify` queries tmux for the
session/pane that fired the event, writes a pending-notification flag for the
tmux status line, and sends a macOS `terminal-notifier` notification whose click
action switches the right tmux client to the right pane. Without tmux, most of
that logic is inert, so the hooks were disabled while running tmux-free.

The `claude-notify` script itself is left in place at
`dotfiles/.local/bin/claude-notify` — nothing references it anymore, but it's
kept so the hooks can be restored without rewriting it.

## To re-enable

Add this `"hooks"` block back into `hosts/MB-HYYXP4CRFJ/.claude/settings.json`
(e.g. right after the `"model"` key, where it used to live):

```json
"hooks": {
  "Stop": [
    {
      "hooks": [
        {
          "type": "command",
          "command": "$HOME/.local/bin/claude-notify"
        }
      ]
    }
  ],
  "Notification": [
    {
      "hooks": [
        {
          "type": "command",
          "command": "$HOME/.local/bin/claude-notify"
        }
      ]
    }
  ]
}
```

Then reload config (open `/hooks` once, or restart Claude Code) so the new hooks
are picked up.
