# Flyb Specs Suggestions

## `show-labels` should support label selection

Current `show-labels=true` is useful, but too broad when notes carry both:

- structural labels such as `design`, `flow`, `usecase`
- semantic/status labels such as `status.completed`, `status.partial`

In practice, users often want to display only a subset of labels.

## Suggested capability

Allow `show-labels` to accept an optional label filter list at section level.

Examples:

- `show-labels=true`
  Current behavior: show all labels.
- `show-labels=status.completed,status.partial,status.missing`
  Show only those labels when present.
- `show-label-prefix=status.`
  Show only labels under a prefix.

## Why this matters

This keeps labels useful for both:

- machine filtering/querying in the model
- clean human-readable rendering in generated markdown

Without filtering, enabling `show-labels` can make output noisy because
taxonomy labels and status labels are mixed together.

## Minimal spec direction

If flyb wants to keep the current boolean form, it could add one extra
argument:

- `show-labels=true`
- `show-labels-include=status.completed,status.partial,status.missing`

Or, more generally:

- `show-labels=true`
- `show-labels-prefix=status.`

## Recommended behavior

- preserve current `show-labels=true`
- add optional filtering, not a breaking change
- keep rendered label order deterministic
- ignore unknown requested labels silently
- allow both exact-label and prefix-based filtering
