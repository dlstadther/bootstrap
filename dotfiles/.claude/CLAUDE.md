# CLAUDE.md

## Response Style

**Summary only.** The final result is what matters. Skip research narration, intermediate findings, and progress updates unless a finding invalidates an assumption the user would otherwise act on.

**No duplication.** Never repeat in a summary what was already stated during research. Say it once, in the right place — the end.

**Punchline first.** Lead with the answer or conclusion. Details follow only if they change what the user should do or understand. If omitting them loses nothing, omit them.

**Fragments are fine.** No mandatory full sentences, proper grammar, or filler. Minimize tokens; the user will ask if they want more.

## Code Comments

Write comments by default to one of two tests; if a comment passes neither, don't write it (and remove it on cleanup):

- **Keep — context not in the code.** The *why*: external facts learned outside the source (data quirks, measured numbers, cadence/schemas), design decisions and the rationale/tradeoff behind them, non-obvious constraints, caveats, links to tickets/docs.
- **Keep — clarifies genuinely complex code.** A non-obvious algorithm, a subtle branch/filter choice, an invariant, a "this looks wrong but is intentional" note.
- **Remove — restates the code.** Comments that paraphrase the line they sit on, label the obvious (`# increment counter`, `// the user id from the request`), or narrate self-evident structure.

Apply this when writing the first draft — don't wait to be told to clean up.

## Git

- Do not commit AI-generated planning artifacts (plans, solution docs, learning summaries) unless explicitly requested.

### Issue Workflow

Each issue (regardless of tracking tool — beads, Jira, Linear, GitHub Issues, etc.) should be:

- Implemented in its **own git worktree and feature branch**
- Merged via its **own PR** — small, focused, independently reviewable
- Kept as small and incremental as possible

Independent issues with no shared unresolved dependencies should be worked **concurrently** in separate worktrees.

### Commit Style

Use **Conventional Commits** format: `type(scope): message`

Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`

#### Scope

- If the current branch name or the user's prompt contains a Jira-style ticket reference (e.g. `ABC-123`), use it as the scope: `feat(abc-123): add login flow`
- Always use lowercase for the scope
- Otherwise, use a short noun describing the affected area (e.g. `feat(auth): ...`)

## Memory

- Proactively save to memory when you learn something non-obvious about my preferences, workflow, or project context.
- Do not save ephemeral task details, git history summaries, or anything derivable from the code.
- Before recalling a memory that names a file or function, verify it still exists.

