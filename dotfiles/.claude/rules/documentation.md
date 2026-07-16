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

- Punch-line first; supporting detail after, and only when the topic needs it.
- Simple words, simple sentences. Reusing the same word is fine if it's the clearest one.
- Be concise over thorough. More detail is not automatically better — don't bury the reader
  (human or agent) in things they don't need.
- KISS applies to docs, not just code.

## Visuals

Use the simplest representation that fits:

1. ASCII art — for simple flowcharts, sequences, gantt-style timelines, trees.
2. Mermaid — only once size or complexity outgrows ASCII, or the diagram needs a style ASCII
   can't express.
3. Anything beyond Mermaid — check with the user first.
