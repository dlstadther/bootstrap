---
name: adding-plugin
description: Use when the user asks to add, enable, or troubleshoot a Claude Code plugin or marketplace for a machine managed by this dotfiles repo, or when a plugin added to settings.json does not appear active after a reload. Covers where to declare it, how to sync the live settings.json, how to register the marketplace with the CLI, and how to confirm the plugin is actually installed.
---

# Adding a Claude Code plugin

This repo manages `~/.claude/settings.json` through symlinks. A plugin or
marketplace only becomes active after four steps: declare it in the repo,
sync the symlink, register the marketplace with the CLI, and install the
plugin. Skipping a step is the most common cause of "I added it but it does
not show up."

## Step 1: Declare the marketplace and plugin in the repo

Do not edit `~/.claude/settings.json` directly. Edit the repo file instead,
so the change survives a fresh machine setup.

Pick the file:

- `dotfiles/.claude/settings.json` for a plugin that every host should get.
- `hosts/<machine>/.claude/settings.json` for one host only. `<machine>` is
  the output of `hostname -s`. The host file applies after the shared file
  and wins when both set the same key.

Add the marketplace under `extraKnownMarketplaces`:

```json
"extraKnownMarketplaces": {
  "my-marketplace": {
    "source": { "source": "github", "repo": "owner/repo" }
  }
}
```

Add the plugin under `enabledPlugins`:

```json
"enabledPlugins": {
  "my-plugin@my-marketplace": true
}
```

## Step 2: Sync the live settings.json

Run `bs audit`. Look at the `.claude/settings.json` line for the shared and
host sections.

- `FOREIGN` means the file is a symlink that points somewhere other than
  its own directory. For the host section this is correct, because the host
  file overrides the shared one at the same target path.
- `REAL FILE` means the symlink is broken. Something replaced it with a
  plain file, so repo edits stop reaching `~/.claude/settings.json`. This
  happens when an app (for example Orca) writes hooks straight into the
  live file, or when the file gets a manual edit.

If you see `REAL FILE`, do not overwrite it blindly. First diff it against
the repo file you plan to symlink to:

```bash
diff <(jq -S . ~/.claude/settings.json) <(jq -S . hosts/<machine>/.claude/settings.json)
```

Anything present only in the live file is a live-only change (for example,
hooks an app added directly). Merge it into the repo file before you
symlink, or you will silently lose it. Then run:

```bash
make install
```

This backs up the current `~/.claude/settings.json` to a timestamped
`.bak` file and replaces it with a symlink into the repo. Run `bs audit`
again to confirm the line now reads `FOREIGN` (or a plain symlink note),
not `REAL FILE`.

## Step 3: Register the marketplace with the CLI

Declaring `extraKnownMarketplaces` in settings.json does not register the
marketplace with the Claude Code CLI. Registration is a separate, one-time
step:

```bash
claude plugin marketplace add owner/repo
```

Use the same `repo` value you put in `extraKnownMarketplaces.source.repo`
(or the git URL, if the source type is `git`). Skip this step and the next
step fails with an error like:

```
Failed to install plugin "my-plugin@my-marketplace": Plugin "my-plugin"
not found in marketplace "my-marketplace". Your local copy may be out of
date — try `claude plugin marketplace update my-marketplace`.
```

That error message suggests running `marketplace update`, but if the
marketplace was never registered, `update` fails too, with
`Marketplace 'my-marketplace' not found`. The real fix is `marketplace add`,
not `update`.

## Step 4: Install the plugin and confirm

```bash
make install-plugins
```

This reads `enabledPlugins` from the now-synced `~/.claude/settings.json`
and installs each one. You can also install a single plugin directly:

```bash
claude plugin install my-plugin@my-marketplace
```

Confirm the result:

```bash
claude plugin marketplace list | grep -i my-marketplace
claude plugin list | grep -i my-plugin
```

Both commands must show the plugin. If a skill inside the plugin should
run automatically, restart Claude Code (or run `/reload-plugins`) and
check the skill list in the system reminder at session start.

## Related: a plugin's output style does not take effect

If a plugin sets `outputStyle` in settings.json but the session still
reports the old style after a restart, check
`.claude/settings.local.json` in the current project. That file is local
to the checkout, is usually untracked, and its `outputStyle` value wins
over the user-level `~/.claude/settings.json`. Remove the stale key there,
or set it to match.
