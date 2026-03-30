# Terminal Agent 🤖

**设备越来越多，人管不过来了。让设备自己管自己。**

AI 驱动的终端设备管理方案 — 从"人管设备"到"设备自治"。

[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-blue.svg)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/Node.js-%3E%3D20-green.svg)](https://nodejs.org/)
[![Go](https://img.shields.io/badge/Go-%3E%3D1.21-cyan.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)

## 为什么需要 DeviceOps？

设备管理的现状：

| 方案 | 能做什么 | 缺什么 |
|------|----------|--------|
| **MDM** | 装软件、锁设备、推配置 | 不会思考，不知道为什么出问题 |
| **IoT 平台** | 采集数据、做看板、发告警 | 告警之后还是靠人 |
| **传统运维** | SSH 上去看日志、手动重启 | 设备一多就撑不住 |

> 三个方案解决了"看见"和"控制"，但都没解决**"决策"**。
>
> DeviceOps 的核心理念：**不是人管设备，而是设备自管理。**

三层架构：

- **感知层** — 设备知道自己怎么了（网络断了、磁盘满了、温度过高）
- **决策层** — 设备知道该怎么做（根据规则和历史自动判断恢复策略）
- **执行层** — 设备自己动手（自动执行恢复，人只看结果）

## 简介

Terminal Agent 是 DeviceOps 理念的 MVP 实现，提供两种使用方式：

- **🧠 TS Agent** — 基于 LLM 的智能运维助手，用自然语言管理设备
- **⚡ Go CLI** — 独立的命令行工具 `device-ctl`，无需 LLM 直接操作

> 💡 本项目是 MVP Demo，使用模拟数据，适合学习 Agent 开发和 CLI 工具开发。

## 功能

- 🔍 **设备查询** — 按区域、状态、类型筛选设备列表
- 📊 **状态监控** — 查看单台设备的详细运行状态和统计
- 📋 **故障日志** — 查询和分析设备故障记录
- 📈 **统计分析** — 按区域/类型/状态聚合统计
- ⚙️ **配置管理** — 修改设备配置（屏幕亮度、超时时间等）
- 🔄 **远程重启** — 模拟远程重启设备
- 🧠 **技能扩展** — 支持故障智能分析和批量配置等扩展技能
- ⚡ **Go CLI** — 独立命令行工具，支持 list/info/stats/logs/monitor/reboot/auth/batch/firmware/terminal 全套命令

## 快速开始

### 方式一：TS Agent（LLM 驱动，自然语言交互）

```bash
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

### 方式二：Go CLI（独立命令行，无需 LLM）

```bash
# 进入 Go 目录
cd cmd/device-ctl

# 编译
go build -o device-ctl .

# 运行
./device-ctl list                    # 查看所有设备
./device-ctl list --area 华东        # 按区域筛选
./device-ctl info SH-PD-001          # 查看设备详情
./device-ctl stats SH-PD-001         # 设备统计
./device-ctl logs --level error      # 故障日志
./device-ctl monitor status          # 状态概览
./device-ctl monitor alerts          # 查看告警
./device-ctl reboot SH-PD-001        # 重启设备
```

Go CLI 基于 [cobra](https://github.com/spf13/cobra) 框架，所有命令支持 `--help` 查看用法。

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
├── packages/              # TS Agent 部分
│   ├── core/              # 数据层：类型定义、存储、模拟数据
│   ├── tools/             # 工具层：Agent 工具 + Go CLI 桥接
│   └── cli/               # 入口层：Agent 组装、系统提示词
├── cmd/                   # Go CLI 部分
│   └── device-ctl/        # CLI 入口（main.go）
├── internal/              # Go 内部包
│   ├── cmd/               # 子命令实现（list/info/stats/logs/monitor/reboot/auth/batch/firmware/terminal）
│   ├── device/            # 设备类型定义
│   └── store/             # 数据存储层（Mock 数据）
├── skills/                # 技能扩展
│   ├── batch-config/      # 批量配置管理
│   └── fault-analysis/    # 故障智能分析
└── scripts/               # 构建脚本
```

### 包依赖

```
# TS Agent
cli → tools → core
  ↘ pi-ai, pi-agent-core

# Go CLI
device-ctl → internal/cmd → internal/store, internal/device
```

## 技术栈

### TS Agent
- **TypeScript** — 类型安全的 JavaScript
- **npm workspaces** — Monorepo 管理
- **[pi-agent-core](https://github.com/nickthecook/pi-agent-core)** — Agent 运行时（工具调用循环）
- **[pi-ai](https://github.com/nickthecook/pi-ai)** — LLM 抽象层（支持 OpenAI 兼容 API）
- **[TypeBox](https://github.com/sinclairzx81/typebox)** — 参数校验

### Go CLI
- **Go** — 高性能编译型语言
- **[Cobra](https://github.com/spf13/cobra)** — CLI 框架
- **Mock 数据层** — 50 台设备，5 个区域，100 条故障日志

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
