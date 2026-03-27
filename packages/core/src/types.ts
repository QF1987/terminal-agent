// ============================================================
// types.ts - 数据类型定义
// ============================================================
// TypeScript 的类型系统：在编译时检查代码，避免运行时错误
// 比如：如果把 string 赋值给 number 类型的变量，编译就会报错
// ============================================================

// ─── 常量数组 + as const ───────────────────────────────────────
// as const 的作用：把数组变成"只读元组"，让 TypeScript 知道里面每个元素的具体值
// 比如 DEVICE_STATUS[0] 的类型是 "online"（字面量类型），而不是 string
export const DEVICE_STATUS = ["online", "offline", "error", "maintenance"] as const;
export const REGIONS = ["华东", "华南", "华北", "西南", "华中"] as const;
export const DEVICE_TYPES = ["自助购药机-标准版", "自助购药机-冷藏版", "自助购药机-大型版"] as const;
export const LOG_TYPES = ["hardware", "software", "network", "medicine_stock"] as const;
export const LOG_SEVERITIES = ["low", "medium", "high", "critical"] as const;

// ─── 类型推导 ─────────────────────────────────────────────────
// typeof DEVICE_STATUS 获取数组的类型
// [number] 表示"数组中任意元素的类型"
// 结果就是 "online" | "offline" | "error" | "maintenance"（联合类型）
// 这样写的好处：改了上面的数组，类型自动跟着变，不用手动同步
export type DeviceStatus = (typeof DEVICE_STATUS)[number];
export type Region = (typeof REGIONS)[number];
export type DeviceType = (typeof DEVICE_TYPES)[number];
export type LogType = (typeof LOG_TYPES)[number];
export type LogSeverity = (typeof LOG_SEVERITIES)[number];

// ─── Record<K, V> 类型 ────────────────────────────────────────
// Record<DeviceStatus, string> 表示"对象的 key 必须是 DeviceStatus 的值，value 是 string"
// 比如：{ online: "在线", offline: "离线", ... }
// 作用：确保所有状态都有对应的中文翻译，不能漏
export const statusToChinese: Record<DeviceStatus, string> = {
  online: "在线",
  offline: "离线",
  error: "故障",
  maintenance: "维护中"
};

export const logTypeToChinese: Record<LogType, string> = {
  hardware: "硬件故障",
  software: "软件故障",
  network: "网络故障",
  medicine_stock: "药品库存"
};

export const severityToChinese: Record<LogSeverity, string> = {
  low: "低",
  medium: "中",
  high: "高",
  critical: "严重"
};

// ─── interface（接口）─────────────────────────────────────────
// interface 定义对象的"形状"，规定有哪些字段、什么类型
// 类似其他语言的 struct/class，但 interface 只描述数据结构，没有方法
export interface DeviceConfig {
  transactionTimeout: number;   // 交易超时（秒）
  screenBrightness: number;     // 屏幕亮度 0-100
  volumeLevel: number;          // 音量 0-100
  autoRebootEnabled: boolean;   // 是否开启自动重启
  autoRebootTime: string;       // 自动重启时间，格式 "03:00"
  medicineCategory: string[];   // 支持的药品类别数组
}

export interface DeviceStats {
  totalTransactions: number;    // 总交易数
  todayTransactions: number;    // 今日交易数
  uptime: number;               // 运行时长（小时）
  faultCount: number;           // 故障次数
}

export interface Device {
  id: string;                   // 设备编号，如 "SH-PD-001"
  name: string;                 // 设备名称，如 "上海浦东-张江药房-01"
  type: DeviceType;             // 设备类型（限制为 DeviceType 的值）
  region: Region;               // 所属区域
  address: string;              // 地址
  status: DeviceStatus;         // 当前状态
  lastHeartbeat: Date;          // 最后心跳时间（Date 是 JS 内置的日期对象）
  firmware: string;             // 固件版本
  config: DeviceConfig;         // 配置（嵌套 interface）
  installedAt: Date;            // 安装日期
  stats: DeviceStats;           // 运行统计（嵌套 interface）
}

export interface FaultLog {
  id: string;                   // 日志 ID
  deviceId: string;             // 关联的设备 ID
  timestamp: Date;              // 故障时间
  type: LogType;                // 故障类型
  severity: LogSeverity;        // 严重程度
  message: string;              // 故障描述
  resolved: boolean;            // 是否已解决
  resolvedAt?: Date;            // 解决时间（? 表示可选字段，可能不存在）
}

// ─── 筛选条件接口 ─────────────────────────────────────────────
// 所有字段都是 ? 可选，表示可以按任意组合筛选
export interface DeviceFilters {
  region?: string;              // 按区域筛选
  status?: string;              // 按状态筛选
  type?: string;                // 按类型筛选
  keyword?: string;             // 按关键字搜索（名称/地址）
}

export interface LogFilters {
  deviceId?: string;            // 按设备筛选
  severity?: string;            // 按严重程度筛选
  type?: string;                // 按故障类型筛选
  days?: number;                // 最近几天（默认7天）
  limit?: number;               // 返回条数限制
}
