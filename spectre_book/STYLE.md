# Spectre Book — Style Guide

Internal reference for all chapter authors.

---

## Voice

- **Second person throughout.** "You define a variable" not "Variables are defined."
- **Conversational, not academic.** Write like you're explaining to a colleague, not submitting a paper.
- **Active voice.** "Spectre checks the invariant" not "the invariant is checked."
- **Section intros: one sentence max**, then jump to code or a bullet list.

**Banned phrases:** "it should be noted", "furthermore", "in this chapter we will explore",
"as we have seen", "in summary", "to summarize", "as mentioned above", "it is important to".

---

## Structure

- **Code first, explanation second.** Show the example, then say what it does.
- **Paragraphs: 2–3 sentences max.** If you're writing a 5-sentence paragraph, split it.
- **Lists over prose** for any group of 3+ items.
- **Numbered lists** for sequential steps. Bullet lists for non-ordered items.
- **Every new concept gets a runnable example within 5 lines** of being introduced.

---

## Diagrams

Use ASCII diagrams for any multi-step workflow:

```
.spec file
    │
    ▼
spectre verify
    │
    ├── safe ───────────────► ✓ done
    │
    └── violation ──────────► trace + --suggest-repairs
                                         │
                                         ▼
                                   fix spec, re-verify
```

Use tables for comparisons (backends, flags, omission kinds, etc.).

---

## Headers

- **Action-oriented.** "Verify your first spec" not "Verification of specifications."
- **Sentence case.** "Reading a counterexample trace" not "Reading A Counterexample Trace."
- **No gerund-heavy titles.** "Fix the bug" not "Fixing the Bug Using the Repair Suggestions Feature."

---

## Code blocks

| Content | Fence |
|---------|-------|
| Spectre spec | ` ```spectre ` |
| Shell commands | ` ```bash ` |
| Terminal output | ` ```text ` |
| TLA+ | ` ```tla ` |
| SMT-LIB2 | ` ```smt2 ` |
| Rust | ` ```rust ` |

Annotate output blocks with a comment when the source isn't obvious:
```bash
spectre verify counter.spec
```
```text
✓ safe — 6 states explored (exhaustive)
```

---

## What a good section looks like

```
## Deposit with an overflow guard

Add a require guard to stop deposits beyond the limit:

    action deposit {
      require balance + amount <= LIMIT
      balance' = balance + amount
    }

- `require` blocks the action when the condition is false — it simply won't fire.
- Spectre checks all reachable states, so no trace ever reaches `balance > LIMIT`.
```

Notice: example first, then bullets explaining the key points. No intro sentence like
"In this section, we will look at how to add a require guard to prevent overflow."
