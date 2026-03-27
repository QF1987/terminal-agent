// ============================================================
// reboot-device.ts - 远程重启设备
// ============================================================
// 模拟远程重启操作
// 重启后设备状态变为"在线"（除非是维护中状态）
// 注意：这是演示用的模拟操作，不是真正重启物理设备
// ============================================================

import { Type } from "@sinclair/typebox";
import type { AgentTool } from "@mariozechner/pi-agent-core";
import { statusToChinese } from "@terminal-agent/core";
import type { DeviceStore } from "@terminal-agent/core";

export function createRebootDeviceTool(store: DeviceStore): AgentTool {
  return {
    name: "reboot_device",
    label: "远程重启设备",
    description: "远程重启指定设备（模拟操作），重启后设备状态将恢复为在线",

    // 只有一个必填参数
    parameters: Type.Object({
      deviceId: Type.String({ description: "设备编号" })
    }),

    // input: any & { deviceId: string }
    // 交叉类型：确保 input 有 deviceId 字段
    async execute(_toolCallId: string, input: any & { deviceId: string }) {
      // 调用 store 的重启方法
      const result = store.rebootDevice(input.deviceId);

      // 找不到设备
      if (!result) {
        return {
          content: [{ type: "text" as const, text: `未找到设备 ${input.deviceId}` }],
          details: { success: false }
        };
      }

      // 返回重启结果
      // 显示之前的状态和当前状态，方便对比
      return {
        content: [{
          type: "text" as const,
          text: `设备 ${input.deviceId} 已发送重启指令。之前状态：${statusToChinese[result.previousStatus]}，当前状态：${statusToChinese[result.device.status]}`
        }],
        details: { success: true }
      };
    }
  };
}
