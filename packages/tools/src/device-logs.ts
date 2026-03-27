// ============================================================
// device-logs.ts - 查询故障日志
// ============================================================
// 查询设备的故障记录，支持多种筛选条件
// 输出格式：每条日志一行，包含时间、设备、严重程度、类型、描述
// ============================================================

import { Type } from "@sinclair/typebox";
import type { AgentTool } from "@mariozechner/pi-agent-core";
import { logTypeToChinese, severityToChinese } from "@terminal-agent/core";
import type { DeviceStore } from "@terminal-agent/core";

export function createDeviceLogsTool(store: DeviceStore): AgentTool {
  return {
    name: "device_logs",
    label: "查询故障日志",
    description: "查询设备故障和操作日志，可按设备、严重程度、类型和天数筛选",

    // 所有参数都是可选的（Type.Optional）
    // 不传参数 = 查询所有设备、所有类型、最近 7 天
    parameters: Type.Object({
      deviceId: Type.Optional(Type.String({ description: "设备编号" })),
      severity: Type.Optional(Type.String({ description: "严重程度：low、medium、high、critical" })),
      type: Type.Optional(Type.String({ description: "故障类型：hardware、software、network、medicine_stock" })),
      
      // Type.Number：数字类型（不是字符串）
      days: Type.Optional(Type.Number({ description: "查询天数，默认 7" })),
      limit: Type.Optional(Type.Number({ description: "返回条数，默认 20" }))
    }),

    async execute(_toolCallId: string, input: any) {
      // 调用 store 的 getLogs 方法（内部会处理筛选逻辑）
      const logs = store.getLogs(input);

      // 条件表达式：如果有日志就格式化，没有就显示"未找到"
      const text = logs.length > 0
        ? logs.map((log: any) =>
            // 每条日志格式：[时间] 设备ID | 严重程度 | 类型: 描述 (状态)
            `[${new Date(log.timestamp).toLocaleString("zh-CN")}] ${log.deviceId} | ${severityToChinese[log.severity]} | ${logTypeToChinese[log.type]}: ${log.message}${log.resolved ? " (已解决)" : " (未解决)"}`
          ).join("\n")  // 用换行符连接所有行
        : "未找到匹配的日志";

      return {
        // 模板字符串：用反引号包裹，${} 插入变量
        content: [{ type: "text" as const, text: `共 ${logs.length} 条日志：\n${text}` }],
        details: { count: logs.length }
      };
    }
  };
}
