#!/usr/bin/env bun
/// <reference types="bun-types" />
/**
 * setup_opencode.ts
 *
 * Interactive wizard that builds ~/.config/opencode/opencode.json from scratch:
 *  - Shows available models (via `opencode models`)
 *  - Walks each agent and lets you pick a model (or keep current)
 *  - Walks each skill in external-agents/skills and asks which to activate via skills.paths
 *  - Writes the final config to ~/.config/opencode/opencode.json AND dotfiles/opencode/opencode.json
 *
 * Usage:
 *   bun run setup_opencode.ts
 *   # or after chmod +x:
 *   ./setup_opencode.ts
 */

import { execSync } from "child_process";
import { existsSync, readdirSync, readFileSync, writeFileSync } from "fs";
import * as path from "path";
import * as readline from "readline";

// ─── paths ───────────────────────────────────────────────────────────────────
const DOTFILES_ROOT = path.resolve(import.meta.dir);
const OPENCODE_DOTFILES = path.join(DOTFILES_ROOT, "opencode");
const SKILLS_DIR = path.join(OPENCODE_DOTFILES, "external-agents", "skills");
const AGENTS_DIR = path.join(OPENCODE_DOTFILES, "agents");
const OPENCODE_CONFIG_DIR = path.join(
  process.env.HOME!,
  ".config",
  "opencode"
);
const CONFIG_DEST = path.join(OPENCODE_CONFIG_DIR, "opencode.json");
const CONFIG_DOTFILES = path.join(OPENCODE_DOTFILES, "opencode.json");

// ─── readline helper ─────────────────────────────────────────────────────────
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

function ask(question: string): Promise<string> {
  return new Promise((resolve) => rl.question(question, resolve));
}

function askYN(question: string, defaultYes = true): Promise<boolean> {
  const hint = defaultYes ? "[Y/n]" : "[y/N]";
  return ask(`${question} ${hint}: `).then((a) => {
    const trimmed = a.trim().toLowerCase();
    if (trimmed === "") return defaultYes;
    return trimmed === "y" || trimmed === "yes";
  });
}

// ─── model picker ────────────────────────────────────────────────────────────
function getAvailableModels(): string[] {
  try {
    const out = execSync("opencode models", { encoding: "utf8" });
    return out
      .trim()
      .split("\n")
      .map((l: string) => l.trim())
      .filter(Boolean);
  } catch {
    console.warn("⚠  Could not run `opencode models`. Falling back to empty list.");
    return [];
  }
}

async function pickModel(
  models: string[],
  agentName: string,
  currentModel: string
): Promise<string> {
  console.log(`\n  Available models:`);
  models.forEach((m, i) => {
    const mark = m === currentModel ? " ◀ current" : "";
    console.log(`    ${String(i + 1).padStart(2)}. ${m}${mark}`);
  });
  console.log(`     0. Keep current (${currentModel})`);

  while (true) {
    const raw = await ask(`  Pick model for "${agentName}" [0-${models.length}]: `);
    const n = parseInt(raw.trim(), 10);
    if (raw.trim() === "" || n === 0) return currentModel;
    if (n >= 1 && n <= models.length) return models[n - 1]!;
    console.log("  Invalid choice, try again.");
  }
}

// ─── skill discovery ─────────────────────────────────────────────────────────
function getSkillName(skillDir: string): { name: string; description: string } {
  const skillFile = path.join(skillDir, "SKILL.md");
  if (!existsSync(skillFile)) return { name: path.basename(skillDir), description: "" };
  const content = readFileSync(skillFile, "utf8");
  const nameMatch = content.match(/^name:\s*(.+)$/m);
  const descMatch = content.match(/^description:\s*(.+)$/m);
  return {
    name: nameMatch?.[1]?.trim() ?? path.basename(skillDir),
    description: descMatch?.[1]?.trim() ?? "",
  };
}

// ─── agent discovery ─────────────────────────────────────────────────────────
interface AgentConfig {
  mode?: string;
  model?: string;
  options?: Record<string, unknown>;
  permission?: Record<string, unknown>;
  [key: string]: unknown;
}

function getAgentFiles(): string[] {
  if (!existsSync(AGENTS_DIR)) return [];
  return readdirSync(AGENTS_DIR)
    .filter((f: string) => f.endsWith(".md"))
    .map((f: string) => path.join(AGENTS_DIR, f));
}

// ─── main ────────────────────────────────────────────────────────────────────
async function main() {
  console.log("\n╔══════════════════════════════════════╗");
  console.log("║    opencode setup wizard             ║");
  console.log("╚══════════════════════════════════════╝\n");

  // Load existing config as base
  let existingConfig: Record<string, unknown> = {};
  if (existsSync(CONFIG_DOTFILES)) {
    try {
      existingConfig = JSON.parse(readFileSync(CONFIG_DOTFILES, "utf8"));
    } catch {
      console.warn("⚠  Could not parse existing opencode.json, starting fresh.");
    }
  }

  const existingAgents = (existingConfig.agent ?? {}) as Record<string, AgentConfig>;

  // ── 1. Models ──────────────────────────────────────────────────────────────
  console.log("⟳  Fetching available models...");
  const models = getAvailableModels();
  if (models.length === 0) {
    console.log("  No models found. Make sure opencode is installed and authenticated.");
  } else {
    console.log(`  Found ${models.length} models.\n`);
  }

  // ── 2. Agents ──────────────────────────────────────────────────────────────
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  console.log("  AGENTS");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");

  // Built-in agents from opencode.json agent block
  const builtinAgentNames = ["plan", "general", "explore", "scout", "single-task-worker"];
  const builtinDefaults: Record<string, AgentConfig> = {
    plan: { mode: "primary", model: "github-copilot/gpt-5.5" },
    general: { mode: "subagent", model: "github-copilot/claude-sonnet-4.6" },
    explore: { mode: "subagent", model: "github-copilot/claude-haiku-4.5" },
    scout: { mode: "subagent", model: "github-copilot/claude-sonnet-4.6" },
    "single-task-worker": { mode: "subagent", model: "github-copilot/claude-sonnet-4.6" },
  };

  const agentFileNames = getAgentFiles().map((f) => path.basename(f, ".md"));

  // Union of known agents
  const allAgentNames = Array.from(
    new Set([...builtinAgentNames, ...agentFileNames])
  );

  const newAgents: Record<string, AgentConfig> = {};

  for (const agentName of allAgentNames) {
    const current = existingAgents[agentName] ?? builtinDefaults[agentName] ?? {};
    const currentModel = (current.model as string | undefined) ?? "github-copilot/claude-sonnet-4.6";
    console.log(`\n  Agent: ${agentName}`);
    if (current.mode) console.log(`  Mode:  ${current.mode}`);

    let chosenModel = currentModel;
    if (models.length > 0) {
      chosenModel = await pickModel(models, agentName, currentModel);
    } else {
      const raw = await ask(`  Model for "${agentName}" [${currentModel}]: `);
      if (raw.trim()) chosenModel = raw.trim();
    }

    newAgents[agentName] = {
      ...(current as Record<string, unknown>),
      model: chosenModel,
      options: current.options ?? {},
      permission: current.permission ?? {},
    };
  }

  // ── 3. Skills ──────────────────────────────────────────────────────────────
  console.log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  console.log("  SKILLS");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");

  const existingSkillPaths = ((existingConfig.skills as { paths?: string[] })?.paths ?? []) as string[];

  // External-agents path is already the auto-load path, no need to add each
  // skill individually. Ask if user wants to include it wholesale.
  const useExternalAgents = await askYN(
    `  Auto-load ALL skills from dotfiles/opencode/external-agents/skills (${SKILLS_DIR})?`,
    true
  );

  const skillPaths: string[] = useExternalAgents ? [SKILLS_DIR] : [];

  if (!useExternalAgents && existsSync(SKILLS_DIR)) {
    const skillDirs = readdirSync(SKILLS_DIR, { withFileTypes: true })
      .filter((d: import("fs").Dirent) => d.isDirectory())
      .map((d: import("fs").Dirent) => path.join(SKILLS_DIR, d.name));

    for (const skillDir of skillDirs) {
      const { name, description } = getSkillName(skillDir);
      const shortDesc = description.length > 60 ? description.slice(0, 57) + "..." : description;
      const wasEnabled = existingSkillPaths.some((p) =>
        p === skillDir || p === SKILLS_DIR
      );
      const enable = await askYN(
        `  Enable skill "${name}"?\n    ${shortDesc}`,
        wasEnabled
      );
      if (enable) skillPaths.push(skillDir);
    }
  }

  // ── 4. MCP servers ─────────────────────────────────────────────────────────
  console.log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  console.log("  MCP SERVERS  (keeping existing config)");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  const existingMcp = existingConfig.mcp ?? {};

  // ── 5. Build final config ──────────────────────────────────────────────────
  const newConfig: Record<string, unknown> = {
    $schema: "https://opencode.ai/config.json",
    instructions: existingConfig.instructions ?? [
      ".scratch/AGENTS.md",
      ".github/copilot-instructions.md",
    ],
    plugin: existingConfig.plugin ?? ["@plannotator/opencode@latest"],
    agent: newAgents,
    skills: skillPaths.length > 0 ? { paths: skillPaths } : undefined,
    mcp: existingMcp,
    shell: existingConfig.shell ?? "zsh",
  };

  // Remove undefined keys
  for (const key of Object.keys(newConfig)) {
    if (newConfig[key] === undefined) delete newConfig[key];
  }

  // ── 6. Preview ─────────────────────────────────────────────────────────────
  console.log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  console.log("  PREVIEW");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");
  console.log(JSON.stringify(newConfig, null, 2));

  const confirm = await askYN("\n  Write this config to disk?", true);
  if (!confirm) {
    console.log("\n  Aborted. No files written.");
    rl.close();
    process.exit(0);
  }

  // ── 7. Write ───────────────────────────────────────────────────────────────
  const json = JSON.stringify(newConfig, null, 2) + "\n";
  writeFileSync(CONFIG_DEST, json, "utf8");
  writeFileSync(CONFIG_DOTFILES, json, "utf8");

  console.log(`\n  ✓  Written to ${CONFIG_DEST}`);
  console.log(`  ✓  Written to ${CONFIG_DOTFILES}`);
  console.log("\n  Restart opencode for changes to take effect.\n");

  rl.close();
}

main().catch((err) => {
  console.error(err);
  rl.close();
  process.exit(1);
});
