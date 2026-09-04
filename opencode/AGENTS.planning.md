# Planning Instructions

1. **SCOPE** — Never include work not explicitly requested in the ticket. Before
   finalizing, audit every phase, step, and section: does it map directly to the
   ticket? Remove any "related bug", "proactive hardening", "while we're at it"
   improvement, or second OPCO not named in the ticket.

2. **ASK BEFORE PLANNING** — Before writing a plan, use the `question` tool to
   surface any ambiguities that only the user can resolve. Do not start drafting
   until requirements are clear.

3. **NO HEDGING** — Never expose research failures or uncertainty in plan text.
   Delete phrases like "the schema is not present", "I got a 403", "this is
   unknown", "the implementation should inspect...". The plan must read as decided
   and authoritative.

4. **BASELINE FIRST** — Open the plan by establishing the current state: what
   files exist, what the functions do today, what the data looks like.

5. **NO EXPOSITION** — Do not explain how an API, auth flow, or library works.
   State what to implement and where (file paths, function names).

6. **CODEBASE FLUENCY** — Verify the approach against the actual codebase before
   proposing it. Search for existing utilities that already do the job, check
   naming conventions in adjacent files, confirm constructor and function
   signatures.

7. **SCOPE TO ONE TARGET FIRST** — One OPCO (stsh first), one app, one
   component. Note where expansion applies; do not plan the expansion.

8. **DEFER NON-BLOCKING** — Analytics, tracking, and behavioral details that
   don't block the main implementation get a one-line "defer to follow-up" note
   and are removed from plan steps.

9. **SELF-CONTAINED** — After approval the orchestrator hands off in a fresh
   session referencing only the plan file on disk. Planning conversation context
   must not carry over. Write the plan to be self-contained.

10. **PERSISTENCE** — Write the plan to `./.scratch/{TICKET}_plan.md` before
    considering it complete. Chat-only plans get denied.

11. **VERIFY, DON'T HEDGE** — If unsure whether an approach works, investigate
    and report a confirmed decision — not "X might work, needs verification."

12. **SPECIFICATION** — Use the correct spec for every artifact. GitHub Copilot
    agent skills live in `.github/skills/`. Review comments are prefixed
    `[standards]: {comment body}`. Check an existing example in the repo before
    guessing.
