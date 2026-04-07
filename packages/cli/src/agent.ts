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

  // 小米模型标记了 reasoning=true 但实际不支持 developer 角色，强制关闭
  if (model.provider === "openrouter" && model.id.startsWith("xiaomi/")) {
    (model as any).reasoning = false;
  }

  if (process.env.LLM_MAX_TOKENS) {
    model.maxTokens = parseInt(process.env.LLM_MAX_TOKENS, 10);
  }

  const agent = new Agent({
    initialState: { systemPrompt: fullPrompt, model, tools: allTools, thinkingLevel: "off" },
    onPayload: (payload: any) => {
      if (process.env.DEBUG_AGENT) {
        // 完整输出 payload（截断 messages 内容）
        const p = { ...payload };
        if (p.messages) {
          p.messages = p.messages.map((m: any) => ({
            role: m.role,
            content: typeof m.content === 'string' ? m.content.slice(0, 50) : m.content
          }));
        }
        if (p.tools) {
          p.tools_sample = p.tools.slice(0, 3).map((t: any) => ({ name: t.function?.name || t.name, type: t.type }));
          p.tools_count = p.tools.length;
          p.tools_first = p.tools[0]; // 完整看第一个tool的格式
        }
        console.log('[DEBUG] payload:', JSON.stringify(p));
      }
    }
  });

  return agent;
}
