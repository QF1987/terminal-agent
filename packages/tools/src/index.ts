// ============================================================
// index.ts - 工具包入口
// ============================================================
// 这个文件是 tools 包的"门面"
// 对外导出所有工具的创建函数，以及一个一键创建全部的函数
// ============================================================

// ─── 类型导入 ─────────────────────────────────────────────
import type { AgentTool } from "@mariozechner/pi-agent-core";
import type { DeviceStore } from "@terminal-agent/core";

// ─── 导入各工具的创建函数 ─────────────────────────────────
// 每个工具文件都导出一个 createXxxTool(store) 函数
// 这里导入它们，统一管理
import { createDeviceLogsTool } from "./device-logs.js";
import { createDeviceStatsTool } from "./device-stats.js";
import { createDeviceStatusTool } from "./device-status.js";
import { createListDevicesTool } from "./list-devices.js";
import { createRebootDeviceTool } from "./reboot-device.js";
import { createUpdateConfigTool } from "./update-config.js";

// ─── 重新导出（Re-export）────────────────────────────────
// export * from：把子模块的所有导出"转发"出去
// 这样其他包只需要 import { createListDevicesTool } from "@terminal-agent/tools"
// 不用关心具体在哪个文件里
export * from "./list-devices.js";
export * from "./device-status.js";
export * from "./device-logs.js";
export * from "./update-config.js";
export * from "./device-stats.js";
export * from "./reboot-device.js";

// ─── 一键创建所有工具 ─────────────────────────────────────
// 工厂函数：接收 store，返回包含所有工具的数组
// 用途：agent.ts 里一行代码就能拿到全部工具
export function createDeviceManagementTools(store: DeviceStore): AgentTool[] {
  return [
    createListDevicesTool(store),   // 查询设备列表
    createDeviceStatusTool(store),  // 查看设备状态
    createDeviceLogsTool(store),    // 查询故障日志
    createUpdateConfigTool(store),  // 修改设备配置
    createDeviceStatsTool(store),   // 统计分析
    createRebootDeviceTool(store)   // 重启设备
  ];
}
