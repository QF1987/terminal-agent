// ============================================================
// go-cli-tools.ts - Go CLI 工具封装
// ============================================================
// 通过 exec 调用 Go 编写的 device-ctl CLI 工具
// 实现 Agent 框架(TS) → CLI工具集(Go) → 设备API/数据源
// ============================================================

import { Type } from "@sinclair/typebox";
import type { AgentTool } from "@mariozechner/pi-agent-core";
import { execSync } from "node:child_process";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Go CLI 路径
const CLI_PATH = join(__dirname, "../../../bin/device-ctl");

// 执行 Go CLI 命令
function execCli(args: string[]): string {
  try {
    const cmd = `${CLI_PATH} ${args.join(" ")}`;
    const result = execSync(cmd, {
      encoding: "utf-8",
      timeout: 10000,
      maxBuffer: 1024 * 1024
    });
    return result.trim();
  } catch (error: any) {
    return `执行失败: ${error.message}`;
  }
}

// ─── 设备列表工具 ─────────────────────────────────────────
export function createGoListDevicesTool(): AgentTool {
  return {
    name: "go_list_devices",
    label: "查询设备列表(Go)",
    description: "查询和筛选设备列表，可按区域、状态、类型和关键字搜索。调用 Go CLI 实现。",
    parameters: Type.Object({
      region: Type.Optional(Type.String({ description: "区域，如华东、华南、华北、西南、华中" })),
      status: Type.Optional(Type.String({ description: "设备状态: online/offline/error/maintenance" })),
      type: Type.Optional(Type.String({ description: "设备类型" })),
      keyword: Type.Optional(Type.String({ description: "设备名称或地址关键字" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["list"];
      if (input.region) args.push("--region", input.region);
      if (input.status) args.push("--status", input.status);
      if (input.type) args.push("--type", input.type);
      if (input.keyword) args.push("--keyword", input.keyword);
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_list_devices", args: input } };
    }
  };
}

// ─── 设备详情工具 ─────────────────────────────────────────
export function createGoDeviceInfoTool(): AgentTool {
  return {
    name: "go_device_info",
    label: "查看设备详情(Go)",
    description: "查看单台设备的详细信息，包括运行统计和配置。调用 Go CLI 实现。",
    parameters: Type.Object({ device_id: Type.String({ description: "设备ID，如 DEV-001" }) }),
    async execute(_toolCallId: string, input: any) {
      const result = execCli(["info", input.device_id]);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_device_info", args: input } };
    }
  };
}

// ─── 设备统计工具 ─────────────────────────────────────────
export function createGoDeviceStatsTool(): AgentTool {
  return {
    name: "go_device_stats",
    label: "查看设备统计(Go)",
    description: "查看设备运行统计数据，包括交易量、运行时长、故障次数。调用 Go CLI 实现。",
    parameters: Type.Object({
      device_id: Type.String({ description: "设备ID" }),
      days: Type.Optional(Type.Number({ description: "统计天数，默认7天" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["stats", input.device_id];
      if (input.days) args.push("--days", String(input.days));
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_device_stats", args: input } };
    }
  };
}

// ─── 故障日志工具 ─────────────────────────────────────────
export function createGoLogsTool(): AgentTool {
  return {
    name: "go_logs",
    label: "查看故障日志(Go)",
    description: "查看设备故障日志，可按设备、严重程度、类型筛选。调用 Go CLI 实现。",
    parameters: Type.Object({
      device: Type.Optional(Type.String({ description: "按设备ID筛选" })),
      severity: Type.Optional(Type.String({ description: "严重程度: low/medium/high/critical" })),
      type: Type.Optional(Type.String({ description: "日志类型: hardware/software/network/medicine_stock" })),
      days: Type.Optional(Type.Number({ description: "最近几天，默认7天" })),
      limit: Type.Optional(Type.Number({ description: "返回条数，默认20" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["logs"];
      if (input.device) args.push("--device", input.device);
      if (input.severity) args.push("--severity", input.severity);
      if (input.type) args.push("--type", input.type);
      if (input.days) args.push("--days", String(input.days));
      if (input.limit) args.push("--limit", String(input.limit));
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_logs", args: input } };
    }
  };
}

// ─── 设备状态概览工具 ─────────────────────────────────────
export function createGoMonitorStatusTool(): AgentTool {
  return {
    name: "go_monitor_status",
    label: "设备状态概览(Go)",
    description: "查看所有设备的状态概览，统计在线/离线/故障/维护中的设备数量。调用 Go CLI 实现。",
    parameters: Type.Object({}),
    async execute(_toolCallId: string, _input: any) {
      const result = execCli(["monitor", "status"]);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_monitor_status" } };
    }
  };
}

// ─── 告警查看工具 ─────────────────────────────────────────
export function createGoMonitorAlertsTool(): AgentTool {
  return {
    name: "go_monitor_alerts",
    label: "查看告警(Go)",
    description: "查看设备告警信息，可按设备、严重程度筛选。调用 Go CLI 实现。",
    parameters: Type.Object({
      device: Type.Optional(Type.String({ description: "按设备ID筛选" })),
      severity: Type.Optional(Type.String({ description: "严重程度: low/medium/high/critical" })),
      limit: Type.Optional(Type.Number({ description: "返回条数，默认20" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["monitor", "alerts"];
      if (input.device) args.push("--device", input.device);
      if (input.severity) args.push("--severity", input.severity);
      if (input.limit) args.push("--limit", String(input.limit));
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_monitor_alerts", args: input } };
    }
  };
}

// ─── 重启设备工具 ─────────────────────────────────────────
export function createGoRebootTool(): AgentTool {
  return {
    name: "go_reboot_device",
    label: "重启设备(Go)",
    description: "远程重启指定设备。调用 Go CLI 实现。",
    parameters: Type.Object({
      device_id: Type.String({ description: "设备ID" }),
      force: Type.Optional(Type.Boolean({ description: "强制重启" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["reboot", input.device_id];
      if (input.force) args.push("--force");
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_reboot_device", args: input } };
    }
  };
}

// ─── 固件管理工具 ─────────────────────────────────────────
export function createGoFirmwareCheckTool(): AgentTool {
  return {
    name: "go_firmware_check",
    label: "检查固件更新(Go)",
    description: "检查设备是否有可用的固件更新。调用 Go CLI 实现。",
    parameters: Type.Object({ region: Type.Optional(Type.String({ description: "按区域筛选" })) }),
    async execute(_toolCallId: string, input: any) {
      const args = ["firmware", "check"];
      if (input.region) args.push("--region", input.region);
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_firmware_check", args: input } };
    }
  };
}

export function createGoFirmwareUpgradeTool(): AgentTool {
  return {
    name: "go_firmware_upgrade",
    label: "升级固件(Go)",
    description: "升级指定设备的固件版本。调用 Go CLI 实现。",
    parameters: Type.Object({
      device_id: Type.String({ description: "设备ID" }),
      version: Type.Optional(Type.String({ description: "目标版本，默认最新" })),
      schedule: Type.Optional(Type.String({ description: "计划升级时间，如 02:00" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["firmware", "upgrade", input.device_id];
      if (input.version) args.push("--version", input.version);
      if (input.schedule) args.push("--schedule", input.schedule);
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_firmware_upgrade", args: input } };
    }
  };
}

// ─── 终端信息工具 ─────────────────────────────────────────
export function createGoTerminalInfoTool(): AgentTool {
  return {
    name: "go_terminal_info",
    label: "查看终端硬件信息(Go)",
    description: "查看设备终端硬件信息，包括CPU、内存、磁盘使用率。调用 Go CLI 实现。",
    parameters: Type.Object({ device_id: Type.String({ description: "设备ID" }) }),
    async execute(_toolCallId: string, input: any) {
      const result = execCli(["terminal", "info", input.device_id]);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_terminal_info", args: input } };
    }
  };
}

export function createGoTerminalNetworkTool(): AgentTool {
  return {
    name: "go_terminal_network",
    label: "查看网络状态(Go)",
    description: "查看设备网络连接状态，包括IP、信号强度、流量统计。调用 Go CLI 实现。",
    parameters: Type.Object({ device_id: Type.String({ description: "设备ID" }) }),
    async execute(_toolCallId: string, input: any) {
      const result = execCli(["terminal", "network", input.device_id]);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_terminal_network", args: input } };
    }
  };
}

// ─── 批量操作工具 ─────────────────────────────────────────
export function createGoBatchRebootTool(): AgentTool {
  return {
    name: "go_batch_reboot",
    label: "批量重启设备(Go)",
    description: "批量重启指定区域的设备。调用 Go CLI 实现。",
    parameters: Type.Object({
      region: Type.String({ description: "目标区域" }),
      confirm: Type.Optional(Type.Boolean({ description: "确认执行" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["batch", "reboot", "--region", input.region];
      if (input.confirm) args.push("--confirm");
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_batch_reboot", args: input } };
    }
  };
}

// ─── Auth 工具 ─────────────────────────────────────────────
export function createGoAuthWhoamiTool(): AgentTool {
  return {
    name: "go_auth_whoami",
    label: "查看当前用户(Go)",
    description: "查看当前登录用户信息。调用 Go CLI 实现。",
    parameters: Type.Object({}),
    async execute(_toolCallId: string, _input: any) {
      const result = execCli(["auth", "whoami"]);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_auth_whoami" } };
    }
  };
}

export function createGoAuthGrantTool(): AgentTool {
  return {
    name: "go_auth_grant",
    label: "授权用户(Go)",
    description: "授权用户访问指定区域或设备。调用 Go CLI 实现。",
    parameters: Type.Object({
      user: Type.String({ description: "目标用户ID（必填）" }),
      region: Type.Optional(Type.String({ description: "授权区域" })),
      device: Type.Optional(Type.String({ description: "授权设备ID" })),
      role: Type.Optional(Type.String({ description: "授权角色" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["auth", "grant", "--user", input.user];
      if (input.region) args.push("--region", input.region);
      if (input.device) args.push("--device", input.device);
      if (input.role) args.push("--role", input.role);
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_auth_grant", args: input } };
    }
  };
}

export function createGoAuthRevokeTool(): AgentTool {
  return {
    name: "go_auth_revoke",
    label: "撤销授权(Go)",
    description: "撤销用户的访问权限。调用 Go CLI 实现。",
    parameters: Type.Object({
      user: Type.String({ description: "目标用户ID（必填）" }),
      region: Type.Optional(Type.String({ description: "撤销区域" })),
      device: Type.Optional(Type.String({ description: "撤销设备ID" }))
    }),
    async execute(_toolCallId: string, input: any) {
      const args = ["auth", "revoke", "--user", input.user];
      if (input.region) args.push("--region", input.region);
      if (input.device) args.push("--device", input.device);
      const result = execCli(args);
      return { content: [{ type: "text" as const, text: result }], details: { command: "go_auth_revoke", args: input } };
    }
  };
}

// ─── 一键创建所有 Go CLI 工具 ─────────────────────────────
export function createGoCLITools(): AgentTool[] {
  return [
    createGoListDevicesTool(),
    createGoDeviceInfoTool(),
    createGoDeviceStatsTool(),
    createGoLogsTool(),
    createGoMonitorStatusTool(),
    createGoMonitorAlertsTool(),
    createGoRebootTool(),
    createGoFirmwareCheckTool(),
    createGoFirmwareUpgradeTool(),
    createGoTerminalInfoTool(),
    createGoTerminalNetworkTool(),
    createGoBatchRebootTool(),
    createGoAuthWhoamiTool(),
    createGoAuthGrantTool(),
    createGoAuthRevokeTool()
  ];
}
