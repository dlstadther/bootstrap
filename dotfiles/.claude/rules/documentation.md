---
paths:
  - "**/*.md"
---

Rules for writing and editing Markdown documentation.

## Location

Prefer a repo-root `docs/` directory. Not strict — defer to:

- explicit user direction
- an existing structure already in place in the repo

## Formatting

- Wrap lines at 120 characters max.
- End every file with a trailing newline.

## Language

Follow the Communication Style rules in `~/.claude/CLAUDE.md` — full sentences only
(the fragments exception there is chat-only).

## Visuals

Use the simplest representation that fits:

1. ASCII art — for simple flowcharts, sequences, gantt-style timelines, trees.
2. Mermaid — only once size or complexity outgrows ASCII, or the diagram needs a style ASCII
   can't express.
3. Anything beyond Mermaid — check with the user first.
