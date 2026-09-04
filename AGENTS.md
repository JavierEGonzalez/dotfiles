# Dotfiles and Configs
This is a highly personalized config repository for the user's specific workflow.

If I request a change in dotfiles, look for them in this repository first.

The files are constantly being updated and tweaked for users' workflow.

The current preferences that might change
- Tmux over zellij
- Preferring Alacritty for terminal
- zsh and oh-my-zsh
- Custom Nvim config as editor
- aliases and scripts for workflow improvements

User uses the create_hardlinks.sh file to add configurations to the respective config directory.

## OpenCode config
If I ask you to do anything related to OpenCode configuration (agents, skills,
commands, plugins, prompts, `opencode.json`), look in the `opencode/` directory
first. See [AGENTS.oc.md](./AGENTS.oc.md) for its layout and how it
maps into `~/.config/opencode/` and `~/.agents/` via `create_hardlinks.sh`.

## Hardlink model

All dotfiles edits go in this repo. `create_hardlinks.sh` distributes them into
place via hard links. Do not edit `~/.config/opencode/`, `~/.agents/`, or any
other linked destination directly — edit the source file here instead.
After making changes, remind the user to run `create_hardlinks.sh` (or the
relevant `link_*` function) if the change needs to take effect in the live config.

The repo `.gitignore` is `*`. New files are invisible to git unless explicit
unignore rules are added to `.gitignore`. Always verify a file is tracked before
assuming it will be committed.
