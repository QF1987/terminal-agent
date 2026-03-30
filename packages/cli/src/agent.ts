// ============================================================
// agent.ts - Agent 组装层（Go CLI 版本）
// ============================================================
import path from "node:path";
import { fileURLToPath } from "node:url";
import { Agent } from "@mariozechner/pi-agent-core";
import { getModel } from "@mariozechner/pi-ai";
import { createGoCLITools } from "@terminal-agent/tools";
import { SYSTEM_PROMPT } from "./prompts.js";
import { buildSkillSnapshot, createReadTool, type SkillsConfig } from "./skills.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export async function createTerminalAgent(skillsConfig?: SkillsConfig) {
  const workspaceDir = path.resolve(__dirname, "..", "..", "..");

  console.log("🔍 扫描技能目录...");
  const snapshot = await buildSkillSnapshot(workspaceDir, skillsConfig);

  const fullPrompt = SYSTEM_PROMPT + (snapshot.prompt ? "\n\n" + snapshot.prompt : "");

  // Go CLI 工具
  const goCLITools = createGoCLITools();
  const readTool = createReadTool(path.join(workspaceDir, "skills"));
  const skillTools = snapshot.skills.flatMap(s => s.tools);
  const allTools = [...goCLITools, readTool, ...skillTools];

  const totalSkills = snapshot.skills.length;
  console.log(`✅ ${totalSkills} 个技能, ${allTools.length} 个工具 (snapshot: ${snapshot.version})\n`);

  const provider = (process.env.LLM_PROVIDER || "openrouter") as any;
  const modelId = process.env.LLM_MODEL || "xiaomi/mimo-v2-pro";
  const model = getModel(provider, modelId);

  if (process.env.LLM_MAX_TOKENS) {
    model.maxTokens = parseInt(process.env.LLM_MAX_TOKENS, 10);
  }

  const agent = new Agent({
    initialState: { systemPrompt: fullPrompt, model, tools: allTools }
  });

  return agent;
}
