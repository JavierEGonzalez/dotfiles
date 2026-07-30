#!/usr/bin/env bun
/// <reference types="bun-types" />
/**
 * setup_opencode.ts
 *
 * Interactive wizard that builds ~/.config/opencode/opencode.json.
 *
 * Navigation:
 *   ↑/↓ or j/k   move
 *   space        toggle (multi-select)
 *   a / n        select all / none (multi-select)
 *   type         filter the list
 *   enter        confirm
 *   esc / ctrl-c cancel
 *
 * Usage:
 *   bun run setup_opencode.ts
 */

import { execSync } from "child_process";
import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "fs";
import * as path from "path";

// ─── paths ───────────────────────────────────────────────────────────────────
const DOTFILES_ROOT = path.resolve(import.meta.dir);
const OPENCODE_DOTFILES = path.join(DOTFILES_ROOT, "opencode");
const SKILLS_DIR = path.join(OPENCODE_DOTFILES, "external-agents", "skills");
const AGENTS_DIR = path.join(OPENCODE_DOTFILES, "agents");
const OPENCODE_CONFIG_DIR = path.join(process.env.HOME!, ".config", "opencode");
const CONFIG_DEST = path.join(OPENCODE_CONFIG_DIR, "opencode.json");
const CONFIG_DOTFILES = path.join(OPENCODE_DOTFILES, "opencode.json");

// ─── ansi ────────────────────────────────────────────────────────────────────
const useColor = !process.env.NO_COLOR && process.stdout.isTTY;
const wrap = (open: string, close: string) => (s: string) =>
  useColor ? `\x1b[${open}m${s}\x1b[${close}m` : s;

const c = {
  reset: "\x1b[0m",
  bold: wrap("1", "22"),
  dim: wrap("2", "22"),
  red: wrap("31", "39"),
  green: wrap("32", "39"),
  yellow: wrap("33", "39"),
  blue: wrap("34", "39"),
  magenta: wrap("35", "39"),
  cyan: wrap("36", "39"),
  gray: wrap("90", "39"),
  inverse: wrap("7", "27"),
};

const HIDE_CURSOR = "\x1b[?25l";
const SHOW_CURSOR = "\x1b[?25h";

// ─── key handling ────────────────────────────────────────────────────────────
interface Key {
  name: string;
  seq: string;
}

function parseKey(buf: Buffer): Key {
  const seq = buf.toString();
  switch (seq) {
    case "\x03":
      return { name: "ctrl-c", seq };
    case "\x1b":
      return { name: "escape", seq };
    case "\r":
    case "\n":
      return { name: "return", seq };
    case " ":
      return { name: "space", seq };
    case "\x7f":
    case "\b":
      return { name: "backspace", seq };
    case "\x1b[A":
    case "\x1bOA":
      return { name: "up", seq };
    case "\x1b[B":
    case "\x1bOB":
      return { name: "down", seq };
    case "\x1b[D":
    case "\x1bOD":
      return { name: "left", seq };
    case "\x1b[C":
    case "\x1bOC":
      return { name: "right", seq };
    case "\x1b[5~":
      return { name: "pageup", seq };
    case "\x1b[6~":
      return { name: "pagedown", seq };
    case "\x1b[H":
    case "\x1bOH":
      return { name: "home", seq };
    case "\x1b[F":
    case "\x1bOF":
      return { name: "end", seq };
    default:
      return { name: "char", seq };
  }
}

function readKey(): Promise<Key> {
  return new Promise((resolve) => {
    const onData = (buf: Buffer) => {
      process.stdin.off("data", onData);
      resolve(parseKey(buf));
    };
    process.stdin.on("data", onData);
  });
}

function enterRawMode() {
  process.stdin.setRawMode(true);
  process.stdin.resume();
  process.stdout.write(HIDE_CURSOR);
}

function exitRawMode() {
  process.stdout.write(SHOW_CURSOR);
  process.stdin.setRawMode(false);
  process.stdin.pause();
}

function cancel(): never {
  exitRawMode();
  console.log(`\n  ${c.red("Cancelled.")} No files written.\n`);
  process.exit(0);
}

// ─── renderer ────────────────────────────────────────────────────────────────
/**
 * Redraws a fixed-height block in place so the list does not scroll the
 * terminal on every keypress.
 */
class Frame {
  private height = 0;

  draw(lines: string[]) {
    if (this.height > 0) process.stdout.write(`\x1b[${this.height}A`);
    const out = lines.map((l) => `\x1b[2K${l}`).join("\n");
    process.stdout.write(out + "\n");
    // Blank out any lines left over from a taller previous frame.
    for (let i = lines.length; i < this.height; i++) {
      process.stdout.write("\x1b[2K\n");
    }
    this.height = Math.max(lines.length, this.height);
  }

  done(lines: string[]) {
    this.draw(lines);
    this.height = 0;
  }
}

// ─── prompt components ───────────────────────────────────────────────────────
interface Choice<T> {
  label: string;
  value: T;
  hint?: string;
}

const VIEWPORT = 10;

function matches(choice: Choice<unknown>, filter: string): boolean {
  if (!filter) return true;
  const needle = filter.toLowerCase();
  return (
    choice.label.toLowerCase().includes(needle) ||
    (choice.hint ?? "").toLowerCase().includes(needle)
  );
}

function viewportSlice<T>(items: T[], cursor: number, size: number) {
  if (items.length <= size) return { start: 0, items };
  let start = cursor - Math.floor(size / 2);
  start = Math.max(0, Math.min(start, items.length - size));
  return { start, items: items.slice(start, start + size) };
}

async function selectOne<T>(
  message: string,
  choices: Choice<T>[],
  initialIndex = 0
): Promise<T> {
  if (choices.length === 0) throw new Error("selectOne called with no choices");

  const frame = new Frame();
  let filter = "";
  let visible = choices;
  let cursor = Math.max(0, Math.min(initialIndex, choices.length - 1));

  const render = (final = false) => {
    const lines: string[] = [];
    lines.push(`${c.green("?")} ${c.bold(message)}`);

    if (final) {
      const picked = visible[cursor]!;
      lines[0] = `${c.green("✓")} ${c.bold(message)} ${c.cyan(picked.label)}`;
      frame.done(lines);
      return;
    }

    if (filter) lines.push(`  ${c.dim("filter:")} ${c.yellow(filter)}`);

    const { start, items } = viewportSlice(visible, cursor, VIEWPORT);
    if (items.length === 0) {
      lines.push(`  ${c.red("no matches")}`);
    }
    items.forEach((choice, i) => {
      const idx = start + i;
      const active = idx === cursor;
      const pointer = active ? c.cyan("❯") : " ";
      const label = active ? c.cyan(choice.label) : choice.label;
      const hint = choice.hint ? ` ${c.gray(choice.hint)}` : "";
      lines.push(`${pointer} ${label}${hint}`);
    });

    if (visible.length > VIEWPORT) {
      lines.push(
        `  ${c.dim(`${cursor + 1}/${visible.length} — type to filter`)}`
      );
    } else {
      lines.push(`  ${c.dim("↑/↓ move · type to filter · enter select")}`);
    }
    frame.draw(lines);
  };

  const reFilter = () => {
    const previous = visible[cursor];
    visible = choices.filter((ch) => matches(ch, filter));
    const keep = previous ? visible.indexOf(previous) : -1;
    cursor = keep >= 0 ? keep : 0;
  };

  render();
  while (true) {
    const key = await readKey();
    if (key.name === "ctrl-c" || key.name === "escape") cancel();

    if (key.name === "return") {
      if (visible.length === 0) continue;
      render(true);
      return visible[cursor]!.value;
    }
    if (key.name === "up") {
      cursor = cursor <= 0 ? visible.length - 1 : cursor - 1;
    } else if (key.name === "down") {
      cursor = cursor >= visible.length - 1 ? 0 : cursor + 1;
    } else if (key.name === "pageup") {
      cursor = Math.max(0, cursor - VIEWPORT);
    } else if (key.name === "pagedown") {
      cursor = Math.min(visible.length - 1, cursor + VIEWPORT);
    } else if (key.name === "home") {
      cursor = 0;
    } else if (key.name === "end") {
      cursor = visible.length - 1;
    } else if (key.name === "backspace") {
      filter = filter.slice(0, -1);
      reFilter();
    } else if (key.name === "space" || key.name === "char") {
      if (key.seq.length === 1 && key.seq >= " ") {
        filter += key.seq;
        reFilter();
      }
    }
    render();
  }
}

async function multiSelect<T>(
  message: string,
  choices: Choice<T>[],
  initialChecked: boolean[] = []
): Promise<T[]> {
  if (choices.length === 0) return [];

  const frame = new Frame();
  const checked = choices.map((_, i) => initialChecked[i] ?? false);
  let cursor = 0;

  const render = (final = false) => {
    const lines: string[] = [];
    const count = checked.filter(Boolean).length;

    if (final) {
      lines.push(
        `${c.green("✓")} ${c.bold(message)} ${c.cyan(`${count} selected`)}`
      );
      frame.done(lines);
      return;
    }
    lines.push(`${c.green("?")} ${c.bold(message)} ${c.dim(`(${count} selected)`)}`);

    const { start, items } = viewportSlice(choices, cursor, VIEWPORT);
    items.forEach((choice, i) => {
      const idx = start + i;
      const active = idx === cursor;
      const box = checked[idx] ? c.green("◉") : c.gray("◯");
      const pointer = active ? c.cyan("❯") : " ";
      const label = active ? c.cyan(choice.label) : choice.label;
      const hint = choice.hint ? ` ${c.gray(choice.hint)}` : "";
      lines.push(`${pointer} ${box} ${label}${hint}`);
    });

    lines.push(
      `  ${c.dim("↑/↓ move · space toggle · a all · n none · enter confirm")}`
    );
    frame.draw(lines);
  };

  render();
  while (true) {
    const key = await readKey();
    if (key.name === "ctrl-c" || key.name === "escape") cancel();

    if (key.name === "return") {
      render(true);
      return choices.filter((_, i) => checked[i]).map((ch) => ch.value);
    }
    if (key.name === "up") {
      cursor = cursor <= 0 ? choices.length - 1 : cursor - 1;
    } else if (key.name === "down") {
      cursor = cursor >= choices.length - 1 ? 0 : cursor + 1;
    } else if (key.name === "pageup") {
      cursor = Math.max(0, cursor - VIEWPORT);
    } else if (key.name === "pagedown") {
      cursor = Math.min(choices.length - 1, cursor + VIEWPORT);
    } else if (key.name === "space") {
      checked[cursor] = !checked[cursor];
    } else if (key.name === "char") {
      const ch = key.seq.toLowerCase();
      if (ch === "a") checked.fill(true);
      else if (ch === "n") checked.fill(false);
      else if (ch === "j") cursor = cursor >= choices.length - 1 ? 0 : cursor + 1;
      else if (ch === "k") cursor = cursor <= 0 ? choices.length - 1 : cursor - 1;
    }
    render();
  }
}

async function confirm(message: string, initial = true): Promise<boolean> {
  return selectOne<boolean>(
    message,
    [
      { label: "Yes", value: true },
      { label: "No", value: false },
    ],
    initial ? 0 : 1
  );
}

// ─── domain helpers ──────────────────────────────────────────────────────────
function getAvailableModels(): string[] {
  try {
    return execSync("opencode models", { encoding: "utf8" })
      .trim()
      .split("\n")
      .map((l: string) => l.trim())
      .filter(Boolean);
  } catch {
    return [];
  }
}

/**
 * Reads a scalar key out of a SKILL.md YAML frontmatter block.
 * Handles plain scalars as well as folded (`>`) and literal (`|`) blocks,
 * whose value lives on the following indented lines.
 */
function readFrontmatterValue(frontmatter: string, key: string): string | undefined {
  const lines = frontmatter.split("\n");
  const start = lines.findIndex((l) => new RegExp(`^${key}:`).test(l));
  if (start === -1) return undefined;

  const inline = lines[start]!.slice(key.length + 1).trim();
  if (inline && !/^[>|][-+]?$/.test(inline)) {
    return inline.replace(/^["']|["']$/g, "");
  }

  const block: string[] = [];
  for (let i = start + 1; i < lines.length; i++) {
    const line = lines[i]!;
    if (line.trim() === "") continue;
    if (!/^\s/.test(line)) break;
    block.push(line.trim());
  }
  return block.length > 0 ? block.join(" ") : undefined;
}

function getSkillMeta(skillDir: string): { name: string; description: string } {
  const skillFile = path.join(skillDir, "SKILL.md");
  const fallback = path.basename(skillDir);
  if (!existsSync(skillFile)) return { name: fallback, description: "" };

  const content = readFileSync(skillFile, "utf8");
  const frontmatter = content.match(/^---\n([\s\S]*?)\n---/)?.[1] ?? content;
  return {
    name: readFrontmatterValue(frontmatter, "name") ?? fallback,
    description: readFrontmatterValue(frontmatter, "description") ?? "",
  };
}

function getAgentFileNames(): string[] {
  if (!existsSync(AGENTS_DIR)) return [];
  return readdirSync(AGENTS_DIR)
    .filter((f: string) => f.endsWith(".md"))
    .map((f: string) => path.basename(f, ".md"));
}

interface AgentConfig {
  mode?: string;
  model?: string;
  options?: Record<string, unknown>;
  permission?: Record<string, unknown>;
  [key: string]: unknown;
}

/** Result of picking a model for one agent. */
type ModelChoice =
  | { kind: "keep" }
  | { kind: "inherit" }
  | { kind: "model"; value: string };

function truncate(s: string, max: number): string {
  return s.length > max ? `${s.slice(0, max - 1)}…` : s;
}

function heading(title: string) {
  const bar = "━".repeat(46);
  console.log(`\n${c.gray(bar)}`);
  console.log(`  ${c.bold(title)}`);
  console.log(`${c.gray(bar)}\n`);
}

// ─── main ────────────────────────────────────────────────────────────────────
async function main() {
  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    console.error("This wizard needs an interactive terminal (TTY).");
    process.exit(1);
  }

  enterRawMode();
  process.on("exit", () => process.stdout.write(SHOW_CURSOR));

  console.log(`\n  ${c.bold(c.magenta("opencode setup wizard"))}`);
  console.log(`  ${c.dim("↑/↓ navigate · enter select · esc cancel")}\n`);

  let existingConfig: Record<string, unknown> = {};
  if (existsSync(CONFIG_DOTFILES)) {
    try {
      existingConfig = JSON.parse(readFileSync(CONFIG_DOTFILES, "utf8"));
    } catch {
      console.log(`  ${c.yellow("⚠")} Could not parse existing opencode.json.`);
    }
  }
  const existingAgents = (existingConfig.agent ?? {}) as Record<string, AgentConfig>;

  // ── agents ────────────────────────────────────────────────────────────────
  heading("AGENTS");

  const builtinDefaults: Record<string, AgentConfig> = {
    plan: { mode: "primary" },
    general: { mode: "subagent" },
    explore: { mode: "subagent" },
    scout: { mode: "subagent" },
    "single-task-worker": { mode: "subagent" },
  };

  const allAgentNames = Array.from(
    new Set([...Object.keys(builtinDefaults), ...getAgentFileNames(), ...Object.keys(existingAgents)])
  );

  console.log(`  ${c.dim("Current agent models:")}`);
  for (const name of allAgentNames) {
    const model = existingAgents[name]?.model;
    const shown = model ? c.cyan(model) : c.gray("(inherits default)");
    console.log(`    ${name.padEnd(20)} ${shown}`);
  }
  console.log();

  const newAgents: Record<string, AgentConfig> = {};
  for (const name of allAgentNames) {
    const base = existingAgents[name] ?? builtinDefaults[name] ?? {};
    newAgents[name] = { ...(base as Record<string, unknown>) };
  }

  const wantModels = await confirm("Configure agent models?", false);

  if (wantModels) {
    const models = getAvailableModels();
    if (models.length === 0) {
      console.log(
        `  ${c.yellow("⚠")} Could not read \`opencode models\`. Skipping model setup.`
      );
    } else {
      console.log(`  ${c.dim(`${models.length} models available.`)}\n`);

      const agentsToEdit = await multiSelect(
        "Which agents do you want to change?",
        allAgentNames.map((name) => ({
          label: name,
          value: name,
          hint: existingAgents[name]?.model ?? "(inherits default)",
        }))
      );

      for (const name of agentsToEdit) {
        const current = existingAgents[name]?.model;
        const choices: Choice<ModelChoice>[] = [
          {
            label: "Keep as is",
            value: { kind: "keep" },
            hint: current ?? "(inherits default)",
          },
          {
            label: "No explicit model",
            value: { kind: "inherit" },
            hint: "inherit opencode's default",
          },
          ...models.map((m) => ({
            label: m,
            value: { kind: "model" as const, value: m },
            hint: m === current ? "← current" : undefined,
          })),
        ];

        const picked = await selectOne(`Model for ${c.magenta(name)}`, choices);
        if (picked.kind === "inherit") {
          delete newAgents[name]!.model;
        } else if (picked.kind === "model") {
          newAgents[name]!.model = picked.value;
        }
      }
    }
  } else {
    console.log(`  ${c.dim("Keeping existing agent models.")}`);
  }

  // ── skills ────────────────────────────────────────────────────────────────
  heading("SKILLS");

  const existingSkillPaths = ((existingConfig.skills as { paths?: string[] })?.paths ??
    []) as string[];
  const autoLoadAll =
    existingSkillPaths.includes(SKILLS_DIR) || existingSkillPaths.includes("~/.agents/skills");

  let skillPaths: string[] = existingSkillPaths;

  if (!existsSync(SKILLS_DIR)) {
    console.log(`  ${c.yellow("⚠")} No skills directory at ${SKILLS_DIR}`);
  } else {
    const skillDirs = readdirSync(SKILLS_DIR, { withFileTypes: true })
      .filter((d) => d.isDirectory())
      .map((d) => path.join(SKILLS_DIR, d.name))
      .sort();

    const metas = skillDirs.map((dir) => ({ dir, ...getSkillMeta(dir) }));

    console.log(`  ${c.dim(`${metas.length} skills found in`)} ${c.gray(SKILLS_DIR)}\n`);
    for (const meta of metas) {
      console.log(`    ${c.cyan(meta.name.padEnd(26))} ${c.gray(truncate(meta.description, 70))}`);
    }
    console.log();

    const mode = await selectOne<"all" | "pick" | "keep">(
      "How do you want to load skills?",
      [
        {
          label: "Auto-load the whole skills directory",
          value: "all",
          hint: "~/.agents/skills",
        },
        { label: "Pick individual skills", value: "pick" },
        { label: "Keep current configuration", value: "keep" },
      ],
      autoLoadAll ? 0 : 1
    );

    if (mode === "all") {
      skillPaths = ["~/.agents/skills"];
    } else if (mode === "pick") {
      const selected = await multiSelect(
        "Select skills to enable",
        metas.map((meta) => ({
          label: meta.name,
          value: meta.dir,
          hint: truncate(meta.description, 54),
        })),
        metas.map((meta) => autoLoadAll || existingSkillPaths.includes(meta.dir))
      );
      skillPaths = selected;

      if (selected.length > 0) {
        console.log(`\n  ${c.dim("Selected skill paths:")}`);
        for (const p of selected) console.log(`    ${c.gray(p)}`);
      }
    }
  }

  // ── build config ──────────────────────────────────────────────────────────
  const newConfig: Record<string, unknown> = {
    $schema: "https://opencode.ai/config.json",
    instructions: existingConfig.instructions,
    plugin: existingConfig.plugin,
    permission: existingConfig.permission,
    agent: newAgents,
    skills: skillPaths.length > 0 ? { paths: skillPaths } : undefined,
    mcp: existingConfig.mcp,
    shell: existingConfig.shell ?? "zsh",
  };
  for (const key of Object.keys(newConfig)) {
    if (newConfig[key] === undefined) delete newConfig[key];
  }

  // ── preview ───────────────────────────────────────────────────────────────
  heading("PREVIEW");
  console.log(JSON.stringify(newConfig, null, 2));
  console.log();

  const write = await confirm("Write this config to disk?", true);
  if (!write) cancel();

  const json = JSON.stringify(newConfig, null, 2) + "\n";
  mkdirSync(OPENCODE_CONFIG_DIR, { recursive: true });
  writeFileSync(CONFIG_DEST, json, "utf8");
  writeFileSync(CONFIG_DOTFILES, json, "utf8");

  console.log(`\n  ${c.green("✓")} ${CONFIG_DEST}`);
  console.log(`  ${c.green("✓")} ${CONFIG_DOTFILES}`);
  console.log(`\n  ${c.dim("Restart opencode for changes to take effect.")}\n`);

  exitRawMode();
}

main().catch((err) => {
  exitRawMode();
  console.error(err);
  process.exit(1);
});
