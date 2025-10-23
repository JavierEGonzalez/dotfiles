<!--
Sync Impact Report
Version: (none) -> 1.0.0
Modified Principles: N/A (initial baseline established)
Added Sections: Operational Constraints & Tooling Standards; Workflow & Review Process
Removed Sections: None
Templates Requiring Updates:
- .specify/templates/plan-template.md: ✅ updated (added Constitution Gates)
- .specify/templates/tasks-template.md: ✅ updated (test/verification guidance aligned)
- .specify/templates/spec-template.md: ✅ aligned (already enforces independent, testable stories)
- .specify/templates/checklist-template.md: ✅ aligned (no conflicting guidance; categories map to principles when generated)
- .specify/templates/agent-file-template.md: ✅ aligned (will reflect principles when populated)
Deferred TODOs:
- RATIFICATION_DATE (original adoption date unknown; baseline established today)
-->

# Dotfiles & Configs Constitution

## Core Principles

### I. Single Source of Truth & Reproducibility
- All configuration files (shell, editor, terminal, multiplexer, scripts) MUST reside in this repository.
- `create_hardlinks.sh` MUST remain idempotent; re-running MUST NOT duplicate or corrupt files.

### II. Minimalism & Purposeful Tooling
- Each new plugin/dependency MUST provide measurable benefit (startup time, productivity, reliability) stated at addition.
- Avoid overlapping tools providing identical function unless an experiment branch is used.
- Preferences (tmux over zellij, Alacritty terminal, Neovim editor) MUST be explicit; experimental alternates live in clearly named branches.
Rationale: Keeps environment fast (<300ms shell startup target) and cognitively manageable.

### III. Safe Change & Fast Feedback
- Script changes MUST support a `-h` or `--help` usage output and be executed once post-change to confirm no errors.
- Risky changes (altering PATH, shell init order) SHOULD be done on a branch `config/<area>-<desc>` and merged only after validation.
- Rollback MUST be trivial (single `git revert` or re-run of `create_hardlinks.sh`).
Rationale: Prevents productivity-blocking breakages and enables rapid iteration.

### IV. Script Reliability & Idempotence
- All scripts in `scripts/` MUST: (1) set `set -euo pipefail` (or equivalent safety) when feasible, (2) emit non‑zero exit codes on failure, (3) avoid destructive operations without explicit user confirmation.
- Actions SHOULD be idempotent (re-running causes no adverse side effects). If not possible, document in script header comment.
- Echo/log key actions; silent failures are forbidden.
- External dependencies MUST be checked (command existence) with graceful error messages.
Rationale: Ensures automation is dependable and debuggable.

### V. Declared Toolchain Preferences & Consistency
- Shell: zsh with oh-my-zsh; auto updates stay disabled unless governance approves change.
- AI/Automation: `workers_ai.sh` + aliases provide assistant functions; modifications MUST retain non-interactive and interactive modes.
Rationale: Clear defaults prevent fragmentation and reduce onboarding friction.

## Operational Constraints & Tooling Standards
- Shell startup goal: <300ms (baseline measured occasionally with `time zsh -i -c exit`).
- No credentials or secrets stored in tracked files.
- Environment managers (e.g., `mise`, `jenv`) MUST be initialized without blocking interactive shell usage.
- Plugin/extension count SHOULD remain minimal; justify additions in commit message.
- Path modifications MUST avoid duplicates and MUST NOT prepend insecure directories.
- Logging/echo from scripts SHOULD be concise (<120 chars per line) and actionable.

## Workflow & Review Process
1. Branch naming: `config/<area>-<short-desc>` for non-trivial changes.
2. Implement change; run validation (fresh shell, Neovim launch, script help output).
3. Apply hardlinks with `create_hardlinks.sh` (ensure idempotent behavior) when necessary.
4. Commit with structured message including type (`add`, `change`, `remove`, `refactor`, `script`) + rationale.
5. Review gating (see Constitution Gates below) prior to merge.
6. Monthly (or pre-release) review ensures principles still fit current workflow; propose amendments via PR.

## Governance
The Constitution supersedes ad hoc practices. Compliance is required for all merges affecting configs or scripts.

Amendment Procedure:
- Proposal PR includes: change summary, principle impact, migration steps (if any), and version bump justification.
- At least one reviewer (self or designated peer) validates gates alignment and absence of regressions.
- On approval, update this file and adjust related templates if gates change.

Versioning Policy:
- MAJOR: Removing or redefining a principle in a backward-incompatible way, or changing a primary tool preference (e.g., replacing tmux with zellij as default).
- MINOR: Adding a new principle, expanding governance rules materially, or introducing new mandatory constraints.
- PATCH: Clarifications, wording improvements, typo fixes, non-semantic refinements.

Compliance & Review:
- Pre-merge checklist must verify: (a) Reproducibility maintained, (b) Minimalism respected (no unused additions), (c) Validation performed, (d) Idempotence preserved in scripts, (e) Preferences unchanged or version bump justified.
- Quarterly or on-demand audits may prune stale configs and confirm performance targets.

Runtime Guidance:
- For day-to-day development commands and structure, refer to `AGENTS.md` and generated development guidelines file (from `agent-file-template.md`) once populated.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): Original adoption date unknown; baseline established | **Last Amended**: 2025-10-23
