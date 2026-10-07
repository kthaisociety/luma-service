# 0002. Registration answers are lists of strings

Status: Accepted

## Context
Luma gives each registration answer a value whose type depends on the question: a string, a list of
strings (multi-select), a boolean, or an object (`company`, `job_title`).

## Decision
`internal/luma` flattens every answer into `values: []string`: one item for text and single-choice, several
for multi-select, `"true"`/`"false"` for checkboxes, the company name only for company answers, and an empty
list when unanswered.

## Consequences
Callers count and group answers without knowing Luma's question types. A caller that needs a company
answer's job title can't get it yet; add it to `Answer` when one does.
