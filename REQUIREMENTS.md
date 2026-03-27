# Terminal Agent - MVP Demo Requirements

## Overview
Build an AI-powered terminal/device management demo. Users interact via natural language (TUI) to manage simulated self-service medicine dispensing machines (自助购药机).

## Tech Stack
- TypeScript, monorepo with npm workspaces
- Use `@mariozechner/pi-ai` for LLM abstraction
- Use `@mariozechner/pi-agent-core` for agent runtime (tool calling loop)
- Use `@mariozechner/pi-tui` for terminal UI
- Node.js >= 20

## Project Structure
```
terminal-agent/
├── package.json              # root workspace
├── tsconfig.json
├── tsconfig.base.json
├── packages/
│   ├── core/                 # device models, mock data, store
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   └── src/
│   │       ├── index.ts
│   │       ├── types.ts      # Device, DeviceStatus, FaultLog, etc.
│   │       ├── store.ts      # In-memory device store with mock data
│   │       └── mock.ts       # Generate realistic mock data (50+ devices)
│   ├── tools/                # Agent tools
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   └── src/
│   │       ├── index.ts
│   │       ├── list-devices.ts     # List/search/filter devices
│   │       ├── device-status.ts    # Get single device status
│   │       ├── device-logs.ts      # Query fault/operation logs
│   │       ├── update-config.ts    # Modify device configuration
│   │       ├── device-stats.ts     # Statistics and analysis
│   │       └── reboot-device.ts    # Remote reboot simulation
│   └── cli/                  # TUI entry point
│       ├── package.json
│       ├── tsconfig.json
│       └── src/
│           ├── index.ts      # Main entry
│           ├── agent.ts      # Assemble agent with tools
│           └── prompts.ts    # System prompt for the agent
```

## Data Models (packages/core/src/types.ts)

```typescript
// 设备状态
type DeviceStatus = 'online' | 'offline' | 'error' | 'maintenance';

// 设备区域
type Region = '华东' | '华南' | '华北' | '西南' | '华中';

// 设备类型
type DeviceType = '自助购药机-标准版' | '自助购药机-冷藏版' | '自助购药机-大型版';

// 设备信息
interface Device {
  id: string;              // e.g. "SH-PD-001"
  name: string;            // e.g. "上海浦东-张江药房-01"
  type: DeviceType;
  region: Region;
  address: string;
  status: DeviceStatus;
  lastHeartbeat: Date;
  firmware: string;        // e.g. "v2.3.1"
  config: DeviceConfig;
  installedAt: Date;
  stats: DeviceStats;
}

// 设备配置
interface DeviceConfig {
  transactionTimeout: number;    // 交易超时(秒)
  screenBrightness: number;      // 屏幕亮度 0-100
  volumeLevel: number;           // 音量 0-100
  autoRebootEnabled: boolean;
  autoRebootTime: string;        // "03:00"
  medicineCategory: string[];    // 支持的药品类别
}

// 设备运行统计
interface DeviceStats {
  totalTransactions: number;
  todayTransactions: number;
  uptime: number;               // 运行时长(小时)
  faultCount: number;           // 总故障次数
}

// 故障日志
interface FaultLog {
  id: string;
  deviceId: string;
  timestamp: Date;
  type: 'hardware' | 'software' | 'network' | 'medicine_stock';
  severity: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  resolved: boolean;
  resolvedAt?: Date;
}
```

## Mock Data Requirements (packages/core/src/mock.ts)
- Generate 50+ devices across all regions
- Realistic Chinese names and addresses
- Random status distribution: ~70% online, ~10% offline, ~15% error, ~5% maintenance
- Generate fault logs (past 7 days, varying severity)
- Realistic transaction stats
- Devices should have varied firmware versions, configs, and install dates

## Agent Tools (packages/tools/)

### 1. list_devices
- Description: 查询和筛选设备列表
- Parameters: region?, status?, type?, keyword? (search name/address)
- Returns: filtered device list with summary

### 2. device_status
- Description: 查看单台设备的详细状态
- Parameters: deviceId (required)
- Returns: full device info including config, stats, recent faults

### 3. device_logs
- Description: 查询设备故障/操作日志
- Parameters: deviceId?, severity?, type?, days? (default 7), limit?
- Returns: fault log list with summary

### 4. update_config
- Description: 修改设备配置
- Parameters: deviceId (required), config partial update
- Returns: old config, new config, confirmation

### 5. device_stats
- Description: 设备统计分析
- Parameters: groupBy? ('region' | 'type' | 'status'), days? (default 7)
- Returns: aggregated statistics, top/bottom performers

### 6. reboot_device
- Description: 远程重启设备（模拟）
- Parameters: deviceId (required)
- Returns: confirmation with estimated reboot time

## System Prompt (packages/cli/src/prompts.ts)
Write a system prompt in Chinese that:
- Defines the agent as a "终端管理助手" (Terminal Management Assistant)
- Explains it manages self-service medicine dispensing machines
- Instructs it to use tools to query data, analyze, and execute commands
- Encourages it to proactively suggest actions (e.g., "这台设备故障率较高，建议安排巡检")
- Output should be concise and professional

## CLI Entry (packages/cli/src/index.ts)
- Use pi-ai to create an LLM client (support OpenAI-compatible API, read from env)
- Use pi-agent-core to create an agent loop with the tools
- Simple REPL: read user input → agent processes → print response → repeat
- No fancy TUI needed for MVP, just stdin/stdout chat loop is fine
- Show a welcome banner with project name and basic instructions

## Important Notes
- All code comments and UI text in Chinese
- Use ES modules ("type": "module")
- Build with TypeScript, target ES2022
- Keep it simple - this is a demo, not production code
- The project should be runnable with: `npm install && npm run build && node packages/cli/dist/index.js`
- LLM API key should come from env var (OPENAI_API_KEY or compatible)
- Make the agent work with any OpenAI-compatible API (OpenAI, DeepSeek, local Ollama, etc.)
