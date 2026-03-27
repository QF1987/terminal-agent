// ============================================================
// mock.ts - 模拟数据生成器
// ============================================================
// 这个文件负责生成逼真的测试数据
// 包括：60+ 台设备、故障日志、随机但合理的状态分布
// ============================================================

import { DEVICE_TYPES, LOG_SEVERITIES, LOG_TYPES, REGIONS } from "./types.js";

// ─── 数据池 ─────────────────────────────────────────────
// 每个区域对应几个城市，每个城市有详细地址
// Record<Region, [string, string][]> 的含义：
//   - key 是 Region（"华东" | "华南" | ...）
//   - value 是 [城市名, 地址][] 的数组（二维数组）
const regionCities: Record<string, [string, string][]> = {
  "华东": [
    ["上海", "浦东新区张江高科博云路88号"],
    ["杭州", "滨江区江南大道3900号"],
    ["南京", "建邺区庐山路168号"],
    ["苏州", "工业园区星湖街218号"],
    ["宁波", "鄞州区首南中路777号"],
    ["合肥", "高新区创新大道2800号"]
  ],
  "华南": [
    ["广州", "天河区珠江新城华夏路16号"],
    ["深圳", "南山区科苑南路2666号"],
    ["佛山", "南海区桂澜北路6号"],
    ["东莞", "南城街道鸿福路200号"],
    ["厦门", "思明区湖滨东路95号"],
    ["福州", "仓山区浦上大道272号"]
  ],
  "华北": [
    ["北京", "朝阳区望京街10号"],
    ["天津", "滨海新区第二大街58号"],
    ["石家庄", "长安区中山东路39号"],
    ["青岛", "崂山区海尔路182号"],
    ["济南", "历下区经十路9777号"],
    ["太原", "小店区长风街123号"]
  ],
  "西南": [
    ["成都", "高新区天府大道中段530号"],
    ["重庆", "渝北区金开大道99号"],
    ["昆明", "官渡区春城路289号"],
    ["贵阳", "观山湖区林城西路8号"],
    ["南宁", "青秀区民族大道136号"],
    ["拉萨", "城关区北京中路51号"]
  ],
  "华中": [
    ["武汉", "洪山区关山大道473号"],
    ["长沙", "岳麓区麓谷大道658号"],
    ["郑州", "郑东新区商务内环路2号"],
    ["南昌", "红谷滩区丰和中大道912号"],
    ["洛阳", "洛龙区开元大道256号"],
    ["宜昌", "西陵区沿江大道52号"]
  ]
};

const pharmacyNames = ["社区健康站", "便民药房", "康宁药房", "仁和药房", "安心药柜", "惠民药房"];
const medicineCategories = [
  "感冒退烧",
  "肠胃用药",
  "慢病常备",
  "外用护理",
  "儿童常备",
  "维生素保健",
  "冷链药品"
];
const firmwareVersions = ["v2.3.1", "v2.3.4", "v2.4.0", "v2.4.2", "v2.5.0"];
const faultMessages = {
  hardware: ["出药电机转速异常", "冷藏压缩机温度偏高", "触控屏局部失灵", "扫码模组识别率下降"],
  software: ["库存同步任务超时", "结算服务进程异常重启", "价格配置校验失败", "本地缓存写入错误"],
  network: ["4G 链路抖动", "主备网络切换失败", "与中心平台连接超时", "心跳包丢失超过阈值"],
  medicine_stock: ["退烧药库存低于安全阈值", "冷链胰岛素库存为零", "常用止咳药补货延迟", "部分批次药品临近效期"]
};

// ─── 伪随机数生成器（PRNG）────────────────────────────────
// 为什么要自己写？因为 Math.random() 每次运行结果不同
// 用固定种子的 PRNG，每次生成的数据一样，方便测试和复现
// 
// 算法：线性同余生成器（LCG）
// 公式：next = (a * current + c) mod m
// 这里 a = 1664525, c = 1013904223, m = 2^32
function createRng(seed = 20260325) {
  // >>> 0：无符号右移 0 位，效果是转成 32 位无符号整数
  let value = seed >>> 0;
  
  // 返回一个闭包（函数里访问外部变量 value）
  // 每次调用返回 [0, 1) 之间的随机数
  return () => {
    value = (value * 1664525 + 1013904223) >>> 0;  // 计算下一个值
    return value / 0x100000000;  // 除以 2^32，归一化到 [0, 1)
  };
}

// 创建随机数生成器实例
const random = createRng();

// ─── 辅助函数 ─────────────────────────────────────────────

// 从数组中随机选一个元素
// readonly T[]：只读数组（不能修改原数组），T 是泛型参数
function pick<T>(items: readonly T[]): T {
  // Math.floor(random() * items.length)：生成 [0, length) 的随机整数
  return items[Math.floor(random() * items.length)]!;
}

// 生成 [min, max] 范围内的随机整数（包含两端）
function integer(min: number, max: number) {
  return Math.floor(random() * (max - min + 1)) + min;
}

// 按权重生成设备状态（模拟真实分布）
// 约 70% 在线，10% 离线，15% 故障，5% 维护
function weightedStatus(index: number) {
  // 用设备序号做哈希，保证同一序号每次状态一样
  const bucket = (index * 7) % 20;
  if (bucket < 14) return "online";       // 14/20 = 70%
  if (bucket < 16) return "offline";      // 2/20 = 10%
  if (bucket < 19) return "error";        // 3/20 = 15%
  return "maintenance";                   // 1/20 = 5%
}

// 生成设备 ID：城市缩写 + 序号
// 如 "SH-PD-001"（上海第1台）
function formatId(cityName: string, sequence: number) {
  // Array.from(cityName)：把字符串转成字符数组（支持中文）
  // .slice(0, 2)：取前两个字
  // .join("")：合并成字符串
  // .toUpperCase()：转大写
  const cityCode = Array.from(cityName).slice(0, 2).join("").toUpperCase();
  // padStart(3, "0")：左边补零到3位
  return `${cityCode}-PD-${String(sequence).padStart(3, "0")}`;
}

// 生成设备名称：城市 + 药房名 + 序号
// 如 "上海浦东-张江药房-01"
function buildName(cityName: string, pharmacyName: string, sequence: number) {
  return `${cityName}-${pharmacyName}-${String(sequence).padStart(2, "0")}`;
}

function createDevice(region: string, cityName: string, address: string, sequence: number) {
  const status = weightedStatus(sequence - 1);
  const type = DEVICE_TYPES[(sequence - 1) % DEVICE_TYPES.length];
  const heartbeatOffsetHours = status === "online" ? integer(0, 2) : integer(6, 48);
  const totalTransactions = integer(12000, 168000);
  const todayTransactions = status === "online" ? integer(18, 210) : integer(0, 32);
  const faultCount = status === "error" ? integer(12, 48) : integer(0, 16);
  const supportsColdChain = type === "自助购药机-冷藏版";

  return {
    id: formatId(cityName, sequence),
    name: buildName(cityName, pick(pharmacyNames), sequence),
    type,
    region,
    address,
    status,
    lastHeartbeat: new Date(Date.now() - heartbeatOffsetHours * 60 * 60 * 1000),
    firmware: pick(firmwareVersions),
    config: {
      transactionTimeout: integer(60, 180),
      screenBrightness: integer(45, 95),
      volumeLevel: integer(20, 85),
      autoRebootEnabled: random() > 0.2,
      autoRebootTime: `${String(integer(1, 5)).padStart(2, "0")}:00`,
      medicineCategory: supportsColdChain
        ? [...new Set(["冷链药品", ...Array.from({ length: 3 }, () => pick(medicineCategories))])]
        : [...new Set(Array.from({ length: 4 }, () => pick(medicineCategories)).filter((item) => item !== "冷链药品"))]
    },
    installedAt: new Date(Date.now() - integer(90, 1500) * 24 * 60 * 60 * 1000),
    stats: {
      totalTransactions,
      todayTransactions,
      uptime: integer(120, 9600),
      faultCount
    }
  };
}

function createFaultLogs(devices: any[]) {
  const logs = [];
  let logSequence = 1;

  for (const device of devices) {
    const baseCount = device.status === "error"
      ? integer(4, 8)
      : device.status === "offline"
        ? integer(2, 5)
        : device.status === "maintenance"
          ? integer(1, 3)
          : integer(0, 3);

    for (let i = 0; i < baseCount; i += 1) {
      const type = pick(LOG_TYPES);
      const severity = device.status === "error" && i === 0 ? "critical" : pick(LOG_SEVERITIES);
      const timestamp = new Date(Date.now() - integer(2, 7 * 24) * 60 * 60 * 1000);
      const resolved = severity === "critical" ? random() > 0.55 : random() > 0.35;
      const resolvedAt = resolved
        ? new Date(timestamp.getTime() + integer(1, 36) * 60 * 60 * 1000)
        : undefined;

      logs.push({
        id: `LOG-${String(logSequence).padStart(5, "0")}`,
        deviceId: device.id,
        timestamp,
        type,
        severity,
        message: pick(faultMessages[type]),
        resolved,
        resolvedAt
      });
      logSequence += 1;
    }
  }

  return logs.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
}

export function generateMockData() {
  const devices = [];
  let sequence = 1;

  for (const region of REGIONS) {
    const cities = regionCities[region];
    for (const [cityName, address] of cities) {
      devices.push(createDevice(region, cityName, address, sequence));
      sequence += 1;
      devices.push(createDevice(region, cityName, `${address.replace(/\d+号$/, "")}${integer(100, 999)}号`, sequence));
      sequence += 1;
    }
  }

  while (devices.length < 60) {
    const region = pick(REGIONS);
    const [cityName, address] = pick(regionCities[region]);
    devices.push(createDevice(region, cityName, `${address.replace(/\d+号$/, "")}${integer(10, 999)}号`, sequence));
    sequence += 1;
  }

  return {
    devices,
    faultLogs: createFaultLogs(devices)
  };
}
