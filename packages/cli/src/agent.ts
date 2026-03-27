// ============================================================
// agent.ts - Agent 组装层
// ============================================================
// 这个文件负责把所有零件组装成一个完整的 Agent
// 就像搭积木：模型（LLM）+ 工具（Tools）+ 提示词（Prompt）= Agent
// ============================================================

// ─── Node.js 内置模块 ─────────────────────────────────────
import path from "node:path";           // 路径处理（拼接、解析路径）
import { fileURLToPath } from "node:url"; // 把 file:// URL 转成文件路径

// ─── 第三方库 ─────────────────────────────────────────────
import { Agent } from "@mariozechner/pi-agent-core";
// Agent：核心类，管理 LLM 调用、工具执行、对话循环

import { getModel } from "@mariozechner/pi-ai";
// getModel：根据 provider 和 modelId 创建 LLM 客户端
// 支持 OpenAI、OpenRouter、DeepSeek 等兼容 API

// ─── 自定义模块 ───────────────────────────────────────────
import { DeviceStore } from "@terminal-agent/core";
import { createDeviceManagementTools } from "@terminal-agent/tools";
import { SYSTEM_PROMPT } from "./prompts.js";
import { buildSkillSnapshot, createReadTool, type SkillsConfig } from "./skills.js";

// ─── __dirname 的 ES Module 写法 ──────────────────────────
// 在 CommonJS（require）里，__dirname 是全局变量，直接可用
// 在 ES Module（import）里，没有 __dirname，需要手动计算
// import.meta.url：当前文件的 file:// URL
// fileURLToPath：转成文件系统路径
// path.dirname()：取目录部分
const __dirname = path.dirname(fileURLToPath(import.meta.url));

// ─── 创建 Agent 的工厂函数 ───────────────────────────────
// SkillsConfig：可选参数，控制技能加载（过滤、数量限制等）
export async function createTerminalAgent(skillsConfig?: SkillsConfig) {
  // 1. 创建数据存储
  const store = new DeviceStore();
  
  // 2. 计算项目根目录
  // __dirname 指向 packages/cli/src/
  // 往上三级：src → cli → packages → 项目根目录
  const workspaceDir = path.resolve(__dirname, "..", "..", "..");

  // 3. 扫描并加载技能
  // buildSkillSnapshot 会：
  //   - 扫描 skills/ 目录
  //   - 读取每个 SKILL.md
  //   - 按 token 限制截断
  //   - 生成版本号（内容 hash）
  console.log("🔍 扫描技能目录...");
  const snapshot = await buildSkillSnapshot(workspaceDir, skillsConfig);

  // 4. 组装系统提示词
  // 基础提示词 + 技能索引（XML 格式）
  // LLM 会读这个提示词，知道自己的角色和可用技能
  const fullPrompt = SYSTEM_PROMPT + (snapshot.prompt ? "\n\n" + snapshot.prompt : "");

  // 5. 组装工具列表
  // 核心工具（6个设备管理工具）
  const coreTools = createDeviceManagementTools(store);
  
  // read 工具：让 Agent 能按需读取 SKILL.md 完整内容
  // 因为提示词里只有技能索引（名称+描述），完整内容需要调用 read 获取
  const readTool = createReadTool(path.join(workspaceDir, "skills"));
  
  // 技能附加工具：每个 skill 可以提供额外的工具
  // .flatMap()：把二维数组展平成一维
  // 比如 [[tool1, tool2], [tool3]] → [tool1, tool2, tool3]
  const skillTools = snapshot.skills.flatMap(s => s.tools);
  
  // 合并所有工具
  const allTools = [...coreTools, readTool, ...skillTools];

  // 打印加载信息
  const totalSkills = snapshot.skills.length;
  console.log(`✅ ${totalSkills} 个技能, ${allTools.length} 个工具 (snapshot: ${snapshot.version})\n`);

  // 6. 创建 LLM 模型客户端
  // 从环境变量读取配置，没设置就用默认值
  const provider = (process.env.LLM_PROVIDER || "openrouter") as any;
  const modelId = process.env.LLM_MODEL || "xiaomi/mimo-v2-pro";
  const model = getModel(provider, modelId);

  // 限制 maxTokens（防止账户余额不足）
  // 如果设置了 LLM_MAX_TOKENS 环境变量，就用它覆盖模型的默认值
  if (process.env.LLM_MAX_TOKENS) {
    model.maxTokens = parseInt(process.env.LLM_MAX_TOKENS, 10);
  }

  // 7. 创建 Agent 实例
  // initialState：Agent 的初始配置
  //   - systemPrompt：系统提示词（定义 Agent 角色和行为）
  //   - model：LLM 模型（用哪个 AI）
  //   - tools：可用工具列表
  const agent = new Agent({
    initialState: {
      systemPrompt: fullPrompt,
      model,
      tools: allTools
    }
  });

  return agent;
}
