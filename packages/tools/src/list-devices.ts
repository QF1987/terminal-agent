// ============================================================
// list-devices.ts - 设备列表查询工具
// ============================================================
// 这是一个 Agent Tool（Agent 工具）
// Agent 是一个 AI 助手，它可以通过"工具"来操作数据
// 比如：用户说"列出华东地区的设备"，Agent 就调用这个工具
// ============================================================

// ─── 导入依赖 ─────────────────────────────────────────────
import { Type } from "@sinclair/typebox";
// TypeBox：一个运行时类型定义库
// 为什么不用 TypeScript 的 interface？因为 TS 的类型在编译后就没了
// Agent 需要在运行时知道参数的结构（比如传给 LLM 做参数校验）
// TypeBox 在运行时也能拿到类型信息

import type { AgentTool } from "@mariozechner/pi-agent-core";
// AgentTool：Agent 工具的接口定义，规定工具必须有 name、description、parameters、execute

import { statusToChinese } from "@terminal-agent/core";
// 从 core 包导入状态中文映射

import type { DeviceStore } from "@terminal-agent/core";
// DeviceStore：设备存储类，用来查询数据

// ─── 辅助函数：按状态统计设备数量 ─────────────────────────
// any[]：表示"任意类型的数组"
// 这里用 any 是因为只需要 device.status 字段，不想引入完整类型
function summarizeByStatus(devices: any[]) {
  // Map<string, number>：key 是状态名，value 是数量
  const buckets = new Map<string, number>();
  
  for (const device of devices) {
    // .get(key) || 0：如果 key 不存在返回 0，然后 +1
    // 相当于：buckets[status] = (buckets[status] || 0) + 1
    buckets.set(device.status, (buckets.get(device.status) || 0) + 1);
  }
  
  // [...buckets.entries()]：把 Map 转成 [[key, value], ...] 数组
  // .map(([status, count]) => ...)：解构每个元组
  // .join("，")：用逗号连接成字符串
  // || "无匹配设备"：如果结果是空字符串，用默认值
  return [...buckets.entries()]
    .map(([status, count]) => `${statusToChinese[status]} ${count} 台`)
    .join("，") || "无匹配设备";
}

// ─── 创建工具的工厂函数 ───────────────────────────────────
// 工厂函数：接收 store，返回一个配置好的工具对象
// 为什么要用工厂？因为工具需要访问 store，但工具对象本身不含 store
// 通过闭包（closure）把 store 注入进去
export function createListDevicesTool(store: DeviceStore): AgentTool {
  return {
    // ── 工具基本信息 ──
    name: "list_devices",               // 工具名称，Agent 调用时用这个名字
    label: "查询设备列表",               // 显示名称
    description: "查询和筛选设备列表，可按区域、状态、类型和关键字搜索",
    // ↑ 这个描述会告诉 Agent 什么时候该用这个工具

    // ── 参数定义（TypeBox 风格）────────────────────────────
    // Type.Object({...})：定义一个对象参数
    // Type.Optional(...)：表示这个参数是可选的（可以不传）
    // Type.String({ description: "..." })：字符串类型，带描述
    parameters: Type.Object({
      region: Type.Optional(Type.String({ description: "区域，如华东、华南、华北、西南、华中" })),
      status: Type.Optional(Type.String({ description: "设备状态，如 online、offline、error、maintenance" })),
      type: Type.Optional(Type.String({ description: "设备类型" })),
      keyword: Type.Optional(Type.String({ description: "设备名称或地址关键字" }))
    }),

    // ── 执行函数 ─────────────────────────────────────────
    // async：表示这是异步函数，可以 await 其他异步操作
    // _toolCallId：工具调用 ID（下划线开头表示"我知道有这个参数，但这里不用"）
    // input：用户传入的参数，类型是 any（因为运行时才知道具体值）
    async execute(_toolCallId: string, input: any) {
      // 调用 store 的方法查询设备
      const devices = store.listDevices(input);

      // 生成统计摘要
      const summary = `共找到 ${devices.length} 台设备，状态分布：${summarizeByStatus(devices)}`;

      // .slice(0, 20)：只取前 20 台（防止数据太多撑爆 prompt）
      // .map()：把每个 device 转成显示格式
      const list = devices.slice(0, 20).map((device: any) => ({
        "设备编号": device.id,
        "设备名称": device.name,
        "区域": device.region,
        "设备类型": device.type,
        "设备状态": statusToChinese[device.status],
        "地址": device.address,
        "今日交易": device.stats.todayTransactions
      }));

      // ── 返回结果 ─────────────────────────────────────
      // Agent Tool 必须返回 { content: [...], details: {...} }
      // content：给 Agent 看的内容（LLM 会读这个来生成回复）
      // details：额外数据（可选，用于日志/调试）
      return {
        content: [{ 
          type: "text" as const,    // as const：让 TypeScript 知道这是字面量 "text"，不是 string
          // JSON.stringify(obj, null, 2)：把对象转成格式化的 JSON 字符串
          // null：不做替换，2：缩进 2 个空格
          text: JSON.stringify({ 摘要: summary, 设备列表: list }, null, 2) 
        }],
        details: { count: devices.length }
      };
    }
  };
}
