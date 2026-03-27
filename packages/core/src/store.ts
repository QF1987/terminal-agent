// ============================================================
// store.ts - 设备数据存储层
// ============================================================
// 这是一个"内存数据库"，数据存在内存里，程序重启就没了
// 实际项目会换成真正的数据库（MySQL、PostgreSQL 等）
// ============================================================

import { generateMockData } from "./mock.js";
// import type：只导入类型，不导入值
// TypeScript 编译后类型会被删除，所以用 import type 更清晰
import type { Device, DeviceConfig, DeviceFilters, FaultLog, LogFilters } from "./types.js";

// ─── 深拷贝函数 ─────────────────────────────────────────────
// 为什么要拷贝？防止外部修改 store 内部的数据
// 比如：你把 device 对象给了别人，别人改了 device.status，store 里的数据也变了
// 拷贝一份，外部改的是副本，不影响原始数据

// 展开运算符 ... 只做浅拷贝（第一层复制，嵌套对象还是同一个引用）
// 所以 Date、数组、嵌套对象要手动创建新实例
function cloneDevice(device: Device): Device {
  return {
    ...device,                                    // 浅拷贝所有顶层字段
    lastHeartbeat: new Date(device.lastHeartbeat), // Date 对象要新建，不然还是指向同一个
    installedAt: new Date(device.installedAt),
    config: { 
      ...device.config,                           // 拷贝配置对象
      medicineCategory: [...device.config.medicineCategory]  // 数组也要新建
    },
    stats: { ...device.stats }                    // 拷贝统计对象
  };
}

function cloneLog(log: FaultLog): FaultLog {
  return {
    ...log,
    timestamp: new Date(log.timestamp),
    // 三元运算符：如果 resolvedAt 存在就新建 Date，否则是 undefined
    resolvedAt: log.resolvedAt ? new Date(log.resolvedAt) : undefined
  };
}

// ─── DeviceStore 类 ─────────────────────────────────────────
// class：定义一个类，包含数据（属性）和行为（方法）
// 类似其他语言的 class，但 TypeScript 的 class 更像语法糖
export class DeviceStore {
  // 属性声明（TypeScript 风格）
  // Map<K, V>：键值对集合，类似其他语言的 HashMap/Dictionary
  // 这里 key 是设备 ID（string），value 是 Device 对象
  devices: Map<string, Device>;
  logs: FaultLog[];

  // 构造函数：创建实例时自动调用
  // seedData 有默认值 = generateMockData()，不传参数就用 mock 数据
  constructor(seedData = generateMockData()) {
    // new Map([...])：把数组转成 Map
    // .map((d) => [d.id, cloneDevice(d)])：把每个 device 变成 [id, device] 的元组
    this.devices = new Map(seedData.devices.map((d) => [d.id, cloneDevice(d)]));
    this.logs = seedData.faultLogs.map(cloneLog);
  }

  // ─── 查询设备列表 ───────────────────────────────────────
  // filters: DeviceFilters = {}：参数类型是 DeviceFilters，默认空对象
  // = {} 的作用：调用时不传参数也不会报错，用空对象代替
  listDevices(filters: DeviceFilters = {}): Device[] {
    // ?. 可选链：如果 filters.keyword 是 undefined/null，整个表达式返回 undefined
    // .trim().toLowerCase()：去掉首尾空格，转小写（搜索时不区分大小写）
    const keyword = filters.keyword?.trim().toLowerCase();
    
    // [...this.devices.values()]：把 Map 的值转成数组
    // .filter()：数组方法，过滤出符合条件的元素
    return [...this.devices.values()].filter((device) => {
      // && 逻辑与：两个条件都满足才继续
      // 如果传了 region 参数，且设备的 region 不匹配，就排除
      if (filters.region && device.region !== filters.region) return false;
      if (filters.status && device.status !== filters.status) return false;
      if (filters.type && device.type !== filters.type) return false;
      
      // 没传关键字，前面的条件都过了，就保留
      if (!keyword) return true;
      
      // .includes()：检查字符串是否包含某个子串
      // 名称或地址包含关键字就保留
      return device.name.toLowerCase().includes(keyword) || device.address.toLowerCase().includes(keyword);
    });
  }

  // ─── 查询单台设备 ─────────────────────────────────────
  // getDevice 返回 Device 或 undefined（找到了就返回设备，没找到返回 undefined）
  // 这是 TypeScript 的联合类型：Device | undefined
  getDevice(deviceId: string): Device | undefined {
    // Map.get(key)：按键查找值，找不到返回 undefined
    return this.devices.get(deviceId);
  }

  // ─── 查询故障日志 ─────────────────────────────────────
  getLogs(filters: LogFilters = {}): FaultLog[] {
    // Date.now()：当前时间戳（毫秒）
    // filters.days || 7：如果传了 days 用传的，否则默认 7
    // 计算 7 天前的时间
    const since = new Date(Date.now() - (filters.days || 7) * 24 * 60 * 60 * 1000);
    
    // 链式 filter：每次 filter 返回新数组，可以连续调用
    let logs = this.logs.filter((log) => log.timestamp >= since);
    if (filters.deviceId) logs = logs.filter((log) => log.deviceId === filters.deviceId);
    if (filters.severity) logs = logs.filter((log) => log.severity === filters.severity);
    if (filters.type) logs = logs.filter((log) => log.type === filters.type);
    
    // .slice(0, n)：取前 n 条
    return logs.slice(0, filters.limit || 20);
  }

  // ─── 获取设备近期故障 ─────────────────────────────────
  getRecentFaults(deviceId: string, limit = 5): FaultLog[] {
    // 参数默认值：limit = 5，不传就用 5
    return this.logs.filter((log) => log.deviceId === deviceId).slice(0, limit);
  }

  // ─── 更新设备配置 ─────────────────────────────────────
  // Partial<T>：把接口所有字段变成可选的
  // 比如 DeviceConfig 有 6 个必填字段，Partial<DeviceConfig> 就变成 6 个可选字段
  // 用途：只想改部分配置，不用传全部字段
  updateDeviceConfig(deviceId: string, partialConfig: Partial<DeviceConfig>) {
    const device = this.devices.get(deviceId);
    if (!device) return null;  // 找不到设备，返回 null

    // 保存旧配置（深拷贝，用于对比/回滚）
    const previousConfig = { ...device.config, medicineCategory: [...device.config.medicineCategory] };

    // Object.entries()：把对象转成 [key, value] 数组
    // Object.fromEntries()：把 [key, value] 数组转回对象
    // .filter(([, v]) => v !== undefined)：过滤掉值为 undefined 的字段
    // 作用：只更新传入的字段，没传的保持原值
    device.config = {
      ...device.config,
      ...Object.fromEntries(Object.entries(partialConfig).filter(([, v]) => v !== undefined)),
      medicineCategory: partialConfig.medicineCategory 
        ? [...partialConfig.medicineCategory]   // 传了就用新的
        : [...device.config.medicineCategory]   // 没传就保持原来的
    };

    return { 
      device, 
      previousConfig, 
      currentConfig: { ...device.config, medicineCategory: [...device.config.medicineCategory] } 
    };
  }

  // ─── 重启设备 ─────────────────────────────────────────
  rebootDevice(deviceId: string) {
    const device = this.devices.get(deviceId);
    if (!device) return null;

    const previousStatus = device.status;
    // 如果不是维护中状态，重启后变在线
    if (device.status !== "maintenance") device.status = "online";
    device.lastHeartbeat = new Date();  // 更新心跳时间

    return { device, previousStatus };
  }

  // ─── 统计概览 ─────────────────────────────────────────
  getStatsOverview() {
    const devices = [...this.devices.values()];
    
    // .reduce()：把数组"归约"成单个值
    // 参数：(累加器, 当前元素) => 新累加值, 初始值
    // 这里 s 是累加器（总和），d 是当前设备
    // s + d.stats.totalTransactions：把每台设备的交易数加起来
    return {
      totalDevices: devices.length,
      totalTransactions: devices.reduce((s, d) => s + d.stats.totalTransactions, 0),
      todayTransactions: devices.reduce((s, d) => s + d.stats.todayTransactions, 0),
      onlineDevices: devices.filter((d) => d.status === "online").length,
      unresolvedFaults: this.logs.filter((l) => !l.resolved).length
    };
  }
}
