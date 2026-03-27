// ============================================================
// core/src/index.ts - 核心包入口
// ============================================================
// Re-export：把子模块的导出"转发"出去
// 这样其他包 import 时只需要：
//   import { DeviceStore, Device, ... } from "@terminal-agent/core"
// 不用关心具体在 types.ts、store.ts 还是 mock.ts 里
// ============================================================

export * from "./types.js";   // 导出所有类型定义（Device、FaultLog、常量等）
export * from "./mock.js";    // 导出 generateMockData 函数
export * from "./store.js";   // 导出 DeviceStore 类
