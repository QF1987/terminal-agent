# Terminal Agent 🤖

AI 驱动的终端设备管理助手 — 用自然语言管理自助购药机网络。

[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-blue.svg)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/Node.js-%3E%3D20-green.svg)](https://nodejs.org/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)

## 简介

Terminal Agent 是一个基于 LLM 的智能运维助手，通过自然语言交互管理自助购药机终端设备。支持设备查询、状态监控、故障分析、批量配置和远程重启等操作。

> 💡 本项目是 MVP Demo，使用模拟数据，适合学习 Agent 开发和 LLM 工具调用模式。

## 功能

- 🔍 **设备查询** — 按区域、状态、类型筛选设备列表
- 📊 **状态监控** — 查看单台设备的详细运行状态和统计
- 📋 **故障日志** — 查询和分析设备故障记录
- 📈 **统计分析** — 按区域/类型/状态聚合统计
- ⚙️ **配置管理** — 修改设备配置（屏幕亮度、超时时间等）
- 🔄 **远程重启** — 模拟远程重启设备
- 🧠 **技能扩展** — 支持故障智能分析和批量配置等扩展技能

## 快速开始

```bash
# 克隆项目
git clone https://github.com/YOUR_USERNAME/terminal-agent.git
cd terminal-agent

# 安装依赖
npm install

# 编译
npm run build

# 配置 API Key
cp .env.example .env
# 编辑 .env，填入你的 API Key

# 运行
OPENROUTER_API_KEY=your_key node packages/cli/dist/index.js
```

### 环境变量

| 变量 | 必填 | 说明 |
|------|------|------|
| `OPENROUTER_API_KEY` | 是* | OpenRouter API Key |
| `OPENAI_API_KEY` | 是* | 或 OpenAI API Key（二选一） |
| `LLM_PROVIDER` | 否 | LLM 提供商，默认 `openrouter` |
| `LLM_MODEL` | 否 | 模型 ID，默认 `google/gemini-2.0-flash-001` |
| `LLM_MAX_TOKENS` | 否 | 最大 token 数 |

> 支持任何 OpenAI 兼容的 API（OpenAI、DeepSeek、本地 Ollama 等）

## 项目结构

```
terminal-agent/
├── packages/
│   ├── core/          # 数据层：类型定义、存储、模拟数据
│   ├── tools/         # 工具层：6 个 Agent 工具
│   └── cli/           # 入口层：Agent 组装、系统提示词
├── skills/            # 技能扩展
│   ├── batch-config/  # 批量配置管理
│   └── fault-analysis/# 故障智能分析
└── scripts/           # 构建脚本
```

### 包依赖

```
cli → tools → core
  ↘ pi-ai, pi-agent-core
```

## 技术栈

- **TypeScript** — 类型安全的 JavaScript
- **npm workspaces** — Monorepo 管理
- **[pi-agent-core](https://github.com/nickthecook/pi-agent-core)** — Agent 运行时（工具调用循环）
- **[pi-ai](https://github.com/nickthecook/pi-ai)** — LLM 抽象层（支持 OpenAI 兼容 API）
- **[TypeBox](https://github.com/sinclairzx81/typebox)** — 参数校验

## 对话示例

```
你: 查一下华东地区有哪些设备离线了
Agent: 正在查询华东地区离线设备...
       找到 3 台离线设备：
       1. 上海浦东-张江药房-01 (SH-PD-001) - 离线 2 小时
       2. 杭州西湖-文三路药房-03 (HZ-XH-003) - 离线 5 小时
       3. 南京鼓楼-中央路药房-02 (NJ-GL-002) - 离线 12 小时
       ⚠️ 建议对离线超过 6 小时的设备安排巡检。

你: 今天天气怎么样
Agent: 我专注于自助购药机终端管理，无法查询天气信息。
       如果您需要管理购药机设备，我可以帮您：...
```

## 开发

```bash
# 清理编译产物
npm run clean

# 重新编译
npm run build
```

## 扩展技能

在 `skills/` 目录下创建新技能：

```
skills/
└── my-skill/
    ├── SKILL.md      # 技能描述（必需）
    └── tools.mjs     # 附加工具（可选）
```

Agent 启动时会自动扫描并加载技能。

## 许可证

[Apache License 2.0](./LICENSE)

## 联系方式

<table>
<tr>
<td align="center" width="200">
<img src="docs/qrcodes/wechat-official.jpg" width="150"><br>
<b>微信公众号「灵枢」</b><br>
<sub>关注获取更多 AI Agent 实践</sub>
</td>
<td align="center" width="200">
<img src="docs/qrcodes/wechat-personal.jpg" width="150"><br>
<b>个人微信</b><br>
<sub>加我交流 Agent 开发</sub>
</td>
</tr>
</table>

## 相关文章

- 📖 [当 AI 学会"把脉"：我用自然语言管理 50 台自助购药机](https://mp.weixin.qq.com/s/IGr5xwuyT9lM9zX3jPY6Ew)（公众号首发）
- 📖 Agent 的终局不是"万能"，而是"专精"（即将发布）
