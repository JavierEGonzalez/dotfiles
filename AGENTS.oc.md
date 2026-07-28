# OpenCode Configuration (`opencode/`)

This directory holds the source-of-truth for the user's OpenCode setup. It is
distributed into place by `create_hardlinks.sh` (see `link_agents`), which:

- Hard-links everything in `opencode/` **except** `external-agents/` into
  `~/.config/opencode/`.
- Hard-links `opencode/external-agents/` into `~/.agents/`.

Because these are hard links, editing a file here is equivalent to editing the
live config. If the user asks to change OpenCode behavior, edit files here first.

## Layout

| Path | Links to | Purpose |
| --- | --- | --- |
| `opencode/opencode.json` | `~/.config/opencode/opencode.json` | Main config: models, agents, MCP servers, plugins, shell. |
| `opencode/AGENTS.md` | `~/.config/opencode/AGENTS.md` | Global runtime agent instructions (JIRA handling, planning mode, testing/lint rules). |
| `opencode/agents/` | `~/.config/opencode/agents/` | Custom subagent definitions (`lint-fixer`, `orchestrator`, `single-task-worker`, etc.). |
| `opencode/command/`, `opencode/commands/` | `~/.config/opencode/...` | Slash command definitions. |
| `opencode/plugins/` | `~/.config/opencode/plugins/` | TypeScript plugins (e.g. `crg-plugin.ts`). |
| `opencode/prompts/` | `~/.config/opencode/prompts/` | Reusable prompt templates. |
| `opencode/external-agents/skills/` | `~/.agents/skills/` | Agent skills, one directory per skill with a `SKILL.md`. |

## Conventions

- **Skills**: each skill is `external-agents/skills/<name>/SKILL.md` with YAML
  frontmatter (`name`, `description`, optional `allowed-tools`,
  `disable-model-invocation`). When adding a skill globally (`~/.agents/skills`),
  also copy it here to keep the repo in sync.
- **Agents**: defined as markdown in `opencode/agents/` and registered in the
  `agent` block of `opencode.json`.
- **`AGENTS.md` vs this file**: `opencode/AGENTS.md` is *runtime* instruction
  content shipped to the OpenCode config dir. This file documents the directory
  layout and is not shipped.
- After changing anything here, remind the user to run `create_hardlinks.sh` if
  the change needs to take effect in the live config.
