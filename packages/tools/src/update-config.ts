// ============================================================
// update-config.ts - 修改设备配置
// ============================================================
// 修改单台设备的配置参数（亮度、音量、超时等）
// 只传要改的字段，没传的保持原值（部分更新）
// ============================================================

import { Type } from "@sinclair/typebox";
import type { AgentTool } from "@mariozechner/pi-agent-core";
import type { DeviceStore } from "@terminal-agent/core";

export function createUpdateConfigTool(store: DeviceStore): AgentTool {
  return {
    name: "update_config",
    label: "修改设备配置",
    description: "修改设备配置参数，如交易超时时间、屏幕亮度、音量、自动重启等",

    // deviceId 是必填的，其他配置字段都是可选的
    parameters: Type.Object({
      deviceId: Type.String({ description: "设备编号" }),
      transactionTimeout: Type.Optional(Type.Number({ description: "交易超时时间（秒）" })),
      screenBrightness: Type.Optional(Type.Number({ description: "屏幕亮度 0-100" })),
      volumeLevel: Type.Optional(Type.Number({ description: "音量 0-100" })),
      autoRebootEnabled: Type.Optional(Type.Boolean({ description: "是否启用自动重启" })),
      // Type.Boolean：布尔类型（true/false）
      autoRebootTime: Type.Optional(Type.String({ description: "自动重启时间，如 03:00" }))
    }),

    async execute(_toolCallId: string, input: any) {
      // ─── 解构赋值 ───────────────────────────────────
      // const { deviceId, ...config } = input
      // 把 input 对象拆开：deviceId 单独拿出来，剩下的打包进 config
      // 比如 input = { deviceId: "SH-001", screenBrightness: 80, volumeLevel: 50 }
      // → deviceId = "SH-001", config = { screenBrightness: 80, volumeLevel: 50 }
      const { deviceId, ...config } = input;

      // ─── 参数校验 ───────────────────────────────────
      // !== undefined：检查这个字段有没有传（没传就是 undefined）
      // 传了才校验，没传跳过
      if (config.screenBrightness !== undefined && (config.screenBrightness < 0 || config.screenBrightness > 100)) {
        return {
          content: [{ type: "text" as const, text: "错误：屏幕亮度必须在 0 到 100 之间" }],
          details: { success: false }
        };
      }
      if (config.volumeLevel !== undefined && (config.volumeLevel < 0 || config.volumeLevel > 100)) {
        return {
          content: [{ type: "text" as const, text: "错误：音量必须在 0 到 100 之间" }],
          details: { success: false }
        };
      }

      // ─── 执行更新 ─────────────────────────────────
      // 调用 store 的方法，传入 deviceId 和要改的配置
      const result = store.updateDeviceConfig(deviceId, config);
      
      // 找不到设备
      if (!result) {
        return {
          content: [{ type: "text" as const, text: `未找到设备 ${deviceId}` }],
          details: { success: false }
        };
      }

      // ─── 生成变更说明 ─────────────────────────────
      // Object.entries(config)：把对象转成 [[key, value], ...]
      // .map(([key, value]) => ...)：解构每个元组，格式化成字符串
      // .join("\n")：用换行符连接
      const changes = Object.entries(config).map(([key, value]) => `  ${key}: ${value}`).join("\n");

      return {
        content: [{ type: "text" as const, text: `设备 ${deviceId} 配置已更新：\n${changes}` }],
        details: { success: true }
      };
    }
  };
}
