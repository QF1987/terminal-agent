// ============================================================
// device-stats.ts - 设备统计分析
// ============================================================
// 按区域、类型或状态分组统计设备数据
// 输出：总体概况 + 分组统计（设备数、交易数、故障数）
// ============================================================

import { Type } from "@sinclair/typebox";
import type { AgentTool } from "@mariozechner/pi-agent-core";
import { statusToChinese } from "@terminal-agent/core";
import type { DeviceStore } from "@terminal-agent/core";

export function createDeviceStatsTool(store: DeviceStore): AgentTool {
  return {
    name: "device_stats",
    label: "设备统计分析",
    description: "设备统计分析，可按区域、类型或状态分组",

    parameters: Type.Object({
      groupBy: Type.Optional(Type.String({ description: "分组方式：region（区域）、type（类型）、status（状态）" })),
      days: Type.Optional(Type.Number({ description: "统计天数，默认 7" }))
    }),

    async execute(_toolCallId: string, input: any & { groupBy?: string; days?: number }) {
      // 获取所有设备（不传筛选条件 = 全部）
      const devices = store.listDevices({});

      // 分组字段，默认按状态分组
      const groupBy = input.groupBy || "status";

      // ─── 分组逻辑 ─────────────────────────────────────
      // Map<string, any[]>：key 是分组值（如 "online"），value 是设备数组
      const groups = new Map<string, any[]>();
      
      for (const device of devices) {
        // (device as any)[groupBy]：动态取属性
        // as any：类型断言，告诉 TypeScript"我知道这个属性存在"
        // 因为 groupBy 是运行时才知道的字符串，TS 无法静态检查
        const key = (device as any)[groupBy] || "未知";
        
        // 如果这个分组还没创建，先创建空数组
        if (!groups.has(key)) groups.set(key, []);
        
        // ! 非空断言：我确定 get(key) 不会返回 undefined
        groups.get(key)!.push(device);
      }

      // ─── 分组标签转换 ─────────────────────────────────
      // 箭头函数：const 函数名 = (参数) => 返回值
      // 如果是状态分组，把英文转成中文；否则原样返回
      const groupLabel = (value: string) => {
        if (groupBy === "status") return statusToChinese[value];
        return value;
      };

      // ─── 生成分组统计 ─────────────────────────────────
      // [...groups.entries()]：把 Map 转成 [[key, value], ...] 数组
      const summary = [...groups.entries()].map(([key, devs]) => {
        // 对每个分组的设备做统计
        const totalTx = devs.reduce((s, d) => s + d.stats.totalTransactions, 0);
        const todayTx = devs.reduce((s, d) => s + d.stats.todayTransactions, 0);
        const faults = devs.reduce((s, d) => s + d.stats.faultCount, 0);
        
        return {
          "分组": groupLabel(key),  // 转换后的标签
          "设备数": devs.length,
          "总交易": totalTx,
          "今日交易": todayTx,
          "总故障": faults
        };
      });

      // 获取全局统计
      const overview = store.getStatsOverview();

      return {
        // 输出两个部分：总体概况 + 分组统计
        content: [{ type: "text" as const, text: JSON.stringify({ "总体概况": overview, "分组统计": summary }, null, 2) }],
        details: { totalDevices: devices.length }
      };
    }
  };
}
