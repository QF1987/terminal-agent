// ============================================================
// device-status.ts - 查看单台设备详细状态
// ============================================================
// 与 list_devices 不同：这个工具查看"一台"设备的完整信息
// 包括配置、统计、近期故障等
// ============================================================

import { Type } from "@sinclair/typebox";
import type { AgentTool } from "@mariozechner/pi-agent-core";
import { logTypeToChinese, severityToChinese, statusToChinese } from "@terminal-agent/core";
import type { DeviceStore } from "@terminal-agent/core";

export function createDeviceStatusTool(store: DeviceStore): AgentTool {
  return {
    name: "device_status",
    label: "查看设备状态",
    description: "查看单台设备的详细状态，包括配置、运行统计和近期故障",

    // 只有一个必填参数：设备编号
    parameters: Type.Object({
      deviceId: Type.String({ description: "设备编号，如 SH-PD-001" })
    }),

    // execute 的参数类型：any & { deviceId: string }
    // 意思是：input 的类型是 any（因为运行时才知道），但我确定它有 deviceId 字段
    // & 是交叉类型：把多个类型合并成一个
    async execute(_toolCallId: string, input: any & { deviceId: string }) {
      // 查询设备
      const device = store.getDevice(input.deviceId);
      
      // 找不到设备，返回错误信息
      if (!device) {
        return {
          content: [{ type: "text" as const, text: `未找到设备 ${input.deviceId}` }],
          details: { found: false }
        };
      }

      // 获取近期故障（最多 5 条）
      const logs = store.getRecentFaults(input.deviceId, 5);

      // 组装详细信息对象
      // 用中文作为 key，这样 JSON 输出直接就是中文
      const info = {
        "设备编号": device.id,
        "设备名称": device.name,
        "区域": device.region,
        "设备类型": device.type,
        "状态": statusToChinese[device.status],
        "地址": device.address,
        "固件版本": device.firmware,
        
        // .toLocaleString("zh-CN")：按中文格式显示日期时间
        // 输出类似 "2026/3/25 14:30:00"
        "最后心跳": new Date(device.lastHeartbeat).toLocaleString("zh-CN"),
        
        "总交易数": device.stats.totalTransactions,
        "今日交易": device.stats.todayTransactions,
        "运行时长": `${device.stats.uptime} 小时`,  // 模板字符串：用 ${} 插入变量
        "故障次数": device.stats.faultCount,
        
        // 嵌套对象：配置详情
        "配置": {
          "交易超时": `${device.config.transactionTimeout} 秒`,
          "屏幕亮度": `${device.config.screenBrightness}%`,
          "音量": `${device.config.volumeLevel}%`,
          
          // 三元运算符：条件 ? 值1 : 值2
          // 如果 autoRebootEnabled 为 true，显示时间和具体时间；否则显示"否"
          "自动重启": device.config.autoRebootEnabled 
            ? `是 (${device.config.autoRebootTime})` 
            : "否",
          
          // .join("、")：用顿号连接数组元素
          "药品类别": device.config.medicineCategory.join("、")
        },

        // 故障日志列表
        // .map()：把每条日志转成格式化字符串
        // || "无近期故障"：如果结果是空字符串，显示默认文本
        "近期故障": logs.map((log: any) =>
          `[${new Date(log.timestamp).toLocaleString("zh-CN")}] ${severityToChinese[log.severity]} - ${logTypeToChinese[log.type]}: ${log.message}${log.resolved ? " (已解决)" : ""}`
        ).join("\n") || "无近期故障"
      };

      return {
        content: [{ type: "text" as const, text: JSON.stringify(info, null, 2) }],
        details: { found: true }
      };
    }
  };
}
