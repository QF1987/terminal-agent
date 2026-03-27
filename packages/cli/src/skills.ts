// ============================================================
// skills.ts - 技能加载系统
// ============================================================
// 这是最复杂的部分，参照 OpenClaw 的实现
// 核心思路：扫描目录 → 加载技能 → 限制 token → 格式化成 prompt
// ============================================================

// ─── Node.js 内置模块 ─────────────────────────────────────
import { createHash } from "node:crypto";  // 生成哈希（用于版本号）
import { readFile, stat, readdir } from "node:fs/promises"; // 文件系统操作（异步版本）
import path from "node:path";
import { pathToFileURL } from "node:url";  // 文件路径转 file:// URL（用于动态 import）

// ─── 类型导入 ─────────────────────────────────────────────
import type { AgentTool } from "@mariozechner/pi-agent-core";
import { Type } from "@sinclair/typebox";
import { loadSkillsFromDir, formatSkillsForPrompt } from "@mariozechner/pi-coding-agent";
// loadSkillsFromDir：扫描目录，解析 SKILL.md，返回技能列表
// formatSkillsForPrompt：把技能列表格式化成 XML 文本（注入到 system prompt）

import type { Skill } from "@mariozechner/pi-coding-agent";
// Skill：技能的数据结构（name, description, baseDir, content 等）

// ─── 配置常量（参照 OpenClaw 默认值）──────────────────
// 最多加载多少个技能到 prompt 里
const DEFAULT_MAX_SKILLS_IN_PROMPT = 150;
// 技能内容在 prompt 里最多占多少字符（防止 prompt 太长）
const DEFAULT_MAX_SKILLS_PROMPT_CHARS = 30000;

// ─── 接口定义 ─────────────────────────────────────────────
// 技能加载配置（所有字段都是可选的）
export interface SkillsConfig {
  maxSkillsInPrompt?: number;    // 最多加载多少个技能
  maxSkillsPromptChars?: number; // 技能内容最大字符数
  skillFilter?: string[];        // 白名单：只加载指定名称的技能
  extraDirs?: string[];          // 额外扫描目录（除了默认的 skills/）
}

// 已加载的技能（技能定义 + 附加工具）
export interface LoadedSkill {
  skill: Skill;        // 技能元数据（名称、描述、内容）
  tools: AgentTool[];  // 这个技能提供的附加工具
}

// 技能快照（一次性返回所有加载结果）
export interface SkillSnapshot {
  prompt: string;      // 格式化后的 prompt 文本（XML 格式）
  version: string;     // 版本号（内容的 hash，用于判断是否有变化）
  skills: LoadedSkill[]; // 已加载的技能列表
  truncated: boolean;  // 是否被截断（超出 token 限制）
}

/**
 * 扫描 skills 目录，加载所有技能
 * 参照 OpenClaw：多来源扫描 + 过滤 + token 限制 + snapshot 版本
 * 
 * @param workspaceDir 项目根目录
 * @param config 可选配置（过滤、限制等）
 * @returns SkillSnapshot 包含 prompt、版本号、已加载技能
 */
export async function buildSkillSnapshot(
  workspaceDir: string,
  config: SkillsConfig = {}
): Promise<SkillSnapshot> {
  // 合并配置（用 ?? 空值合并：左边是 null/undefined 就用右边的默认值）
  const limits = {
    maxSkillsInPrompt: config.maxSkillsInPrompt ?? DEFAULT_MAX_SKILLS_IN_PROMPT,
    maxSkillsPromptChars: config.maxSkillsPromptChars ?? DEFAULT_MAX_SKILLS_PROMPT_CHARS,
  };

  // ═══════════════════════════════════════════════════════════
  // 1. 多来源扫描（参照 OpenClaw 的 7 个来源）
  // ═══════════════════════════════════════════════════════════
  const allSkills: Skill[] = [];
  // Map 用于去重：同名技能只保留第一个
  const seen = new Map<string, Skill>();

  // 来源 1: 项目 skills/ 目录（项目级技能）
  await scanSource(path.resolve(workspaceDir, "skills"), "workspace", seen, allSkills);

  // 来源 2: 用户目录（用户自定义技能）
  // process.env.HOME：macOS/Linux 的用户主目录
  // process.env.USERPROFILE：Windows 的用户主目录
  const home = process.env.HOME || process.env.USERPROFILE;
  if (home) {
    await scanSource(path.join(home, ".terminal-agent", "skills"), "user", seen, allSkills);
  }

  // 来源 3: 额外指定的目录
  // ?? []：如果 config.extraDirs 是 null/undefined，用空数组
  for (const dir of config.extraDirs ?? []) {
    await scanSource(path.resolve(dir), "extra", seen, allSkills);
  }

  // ═══════════════════════════════════════════════════════════
  // 2. 过滤（如果指定了白名单）
  // ═══════════════════════════════════════════════════════════
  let filtered = allSkills;
  if (config.skillFilter && config.skillFilter.length > 0) {
    // .filter() 只保留白名单里的技能
    // ! 非空断言：告诉 TypeScript "我确定 skillFilter 不是 null"
    filtered = allSkills.filter(s => config.skillFilter!.includes(s.name));
  }

  // ═══════════════════════════════════════════════════════════
  // 3. Token 限制
  // ═══════════════════════════════════════════════════════════
  // 先按数量限制
  let limited = filtered.slice(0, limits.maxSkillsInPrompt);
  let truncated = filtered.length > limited.length;

  // 再按字符数限制（二分查找最大可容纳数量）
  // 为什么要二分？因为格式化后的字符数和技能数量不是线性关系
  // 二分查找效率更高：O(log n) vs O(n)
  if (!fitsInPrompt(limited, limits.maxSkillsPromptChars)) {
    let lo = 0, hi = limited.length;
    while (lo < hi) {
      const mid = Math.ceil((lo + hi) / 2);
      if (fitsInPrompt(limited.slice(0, mid), limits.maxSkillsPromptChars)) lo = mid;
      else hi = mid - 1;
    }
    limited = limited.slice(0, lo);
    truncated = true;
  }

  // ═══════════════════════════════════════════════════════════
  // 4. 格式化成 prompt
  // ═══════════════════════════════════════════════════════════
  // formatSkillsForPrompt：把技能列表转成 XML 格式的文本
  // 输出类似：
  // <available_skills>
  //   <skill name="fault-analysis" description="...">
  //     [完整 SKILL.md 内容]
  //   </skill>
  // </available_skills>
  const prompt = formatSkillsForPrompt(limited);

  // ═══════════════════════════════════════════════════════════
  // 5. 生成版本号
  // ═══════════════════════════════════════════════════════════
  // SHA-256 哈希：把内容变成固定长度的"指纹"
  // 内容变了，哈希就变；内容没变，哈希一样
  // 用途：判断 prompt 是否变化，避免重复发送给 LLM
  const version = createHash("sha256").update(prompt).digest("hex").slice(0, 12);
  // .digest("hex")：输出十六进制字符串
  // .slice(0, 12)：取前 12 位（够用了，不用全 64 位）

  // ═══════════════════════════════════════════════════════════
  // 6. 加载每个技能的附加工具
  // ═══════════════════════════════════════════════════════════
  const loadedSkills: LoadedSkill[] = [];
  for (const skill of limited) {
    const tools = await loadSkillTools(skill);
    loadedSkills.push({ skill, tools });
  }

  if (truncated) {
    console.log(`  ⚠️  Skills truncated: ${limited.length}/${filtered.length}`);
  }

  return { prompt, version, skills: loadedSkills, truncated };
}

/**
 * 扫描单个来源目录，加载技能到 seen Map（去重）和 allSkills 数组
 * 
 * @param dir 要扫描的目录路径
 * @param source 来源标识（"workspace" | "user" | "extra"）
 * @param seen 去重用的 Map（key 是技能名称）
 * @param allSkills 收集所有技能的数组
 */
async function scanSource(
  dir: string,
  source: string,
  seen: Map<string, Skill>,
  allSkills: Skill[]
): Promise<void> {
  // try-catch：捕获异常，防止目录不存在时报错
  try {
    // loadSkillsFromDir：扫描目录，解析每个子目录的 SKILL.md
    // 返回 { skills: [...], diagnostics: [...] }
    const { skills, diagnostics } = loadSkillsFromDir({ dir, source });

    // 输出警告信息（比如 SKILL.md 格式不对）
    for (const d of diagnostics) {
      if (d.type === "warning") {
        console.log(`  ⚠️  ${d.message} (${d.path})`);
      }
    }

    // 去重：同名技能只保留第一个遇到的
    for (const skill of skills) {
      if (!seen.has(skill.name)) {
        seen.set(skill.name, skill);  // 记录已见过
        allSkills.push(skill);        // 加入列表
        console.log(`  📦 ${skill.name} — ${skill.description.slice(0, 60)}...`);
      }
    }
  } catch {
    // 目录不存在或读取失败，静默跳过（不报错）
  }
}

/**
 * 检查技能列表格式化后是否在字符限制内
 * 
 * @param skills 技能列表
 * @param maxChars 最大字符数
 * @returns true = 在限制内，false = 超出限制
 */
function fitsInPrompt(skills: Skill[], maxChars: number): boolean {
  // 先格式化成文本，再检查长度
  return formatSkillsForPrompt(skills).length <= maxChars;
}

/**
 * 加载单个技能的附加工具
 * 
 * 技能目录下可以有 tools.mjs 或 tools.js 文件
 * 文件导出方式：
 *   - export default function() { return [...] }  （默认导出函数）
 *   - export const tools = [...]                   （命名导出数组）
 * 
 * @param skill 技能对象
 * @returns 工具数组
 */
async function loadSkillTools(skill: Skill): Promise<AgentTool[]> {
  // 尝试 .mjs 和 .js 两种扩展名
  for (const ext of [".mjs", ".js"]) {
    const toolsPath = path.join(skill.baseDir, `tools${ext}`);
    try {
      // stat：检查文件是否存在（不存在会抛异常）
      await stat(toolsPath);
      
      // 动态 import：运行时加载 ES Module
      // pathToFileURL：把文件路径转成 file:// URL（import 需要 URL 格式）
      const mod = await import(pathToFileURL(toolsPath).href);
      
      // 检查导出方式
      if (typeof mod.default === "function") {
        // 默认导出函数：调用它获取工具列表
        return mod.default();
      }
      if (Array.isArray(mod.tools)) {
        // 命名导出数组：直接用
        return mod.tools;
      }
    } catch {
      // 文件不存在或加载失败，跳过
    }
  }
  return [];  // 没有工具文件，返回空数组
}

/**
 * 创建 "read" 工具 — 让 Agent 能按需读取 SKILL.md 完整内容
 * 
 * 工作流程：
 *   1. system prompt 里有 <available_skills> 索引（只含名称+描述）
 *   2. LLM 看到感兴趣的技能，调用 read 工具读取 SKILL.md
 *   3. LLM 根据完整内容决定如何使用这个技能
 * 
 * 这是 OpenClaw/Claude Code 的核心模式：
 * 不是把所有技能内容都塞进 prompt（太占 token），
 * 而是让 LLM 自己按需读取
 * 
 * @param skillsDir 技能目录路径
 */
export function createReadTool(skillsDir: string): AgentTool {
  return {
    name: "read",
    label: "读取文件",
    description: "读取指定文件的内容，用于加载技能的 SKILL.md 或查看其他文件",
    
    // 参数：文件路径（字符串）
    parameters: Type.Object({
      file_path: Type.String({ description: "文件路径，可以是绝对路径或相对于项目根目录" })
    }),

    async execute(_toolCallId: string, input: { file_path: string }) {
      let filePath = input.file_path;

      // path.isAbsolute()：判断是否是绝对路径（以 / 开头）
      // 如果是相对路径，基于项目根目录解析
      if (!path.isAbsolute(filePath)) {
        // skillsDir 指向 skills/ 目录
        // .. 往上一级就是项目根目录
        filePath = path.join(skillsDir, "..", filePath);
      }

      try {
        // readFile：异步读取文件内容
        // "utf-8"：指定编码，返回字符串（不指定返回 Buffer）
        const content = await readFile(filePath, "utf-8");
        
        // 截断：超过 8000 字符的内容截断（防止 prompt 太长）
        const truncated = content.length > 8000
          ? content.slice(0, 8000) + "\n... (文件较大，已截断)"
          : content;
        
        return {
          content: [{ type: "text" as const, text: truncated }],
          details: { path: filePath, size: content.length }
        };
      } catch (error: any) {
        // error.message：异常信息
        return {
          content: [{ type: "text" as const, text: `读取失败：${error.message}` }],
          details: { path: filePath, error: true }
        };
      }
    }
  };
}
