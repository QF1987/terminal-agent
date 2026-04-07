# KAIROS 设计草案：terminal-agent 主动感知层

> 受 Anthropic KAIROS 概念启发——从"人问设备"到"设备找人"。
> 本文档将 KAIROS 哲学映射到 terminal-agent 的 DeviceOps 架构中。

---

## 一、为什么 terminal-agent 天然适合 KAIROS

KAIROS 的前提是：AI 需要持续感知用户上下文，在"恰当的时机"主动开口。

设备管理的场景比通用 AI 助手更容易落地，因为：

| 通用 AI（ChatGPT 等） | 设备管理（terminal-agent） |
|---|---|
| 需要侵入式感知（屏幕、文件、聊天） | 设备本身就在产生数据（心跳、指标、日志） |
| 用户隐私敏感度极高 | 设备数据天然可采集 |
| "恰当时机"难定义 | 异常 = 时机，规则清晰 |
| 主动介入容易惹人烦 | 设备异常时主动告警是刚需 |

**核心判断：KAIROS 在设备管理领域不是锦上添花，是必须有的能力。**

---

## 二、架构总览

```
┌─────────────────────────────────────────────────────┐
│                   Operator（人）                      │
│         终端 / 手机 / Web Dashboard                   │
└──────────────────────┬──────────────────────────────┘
                       │ 主动推送（告警 / 建议 / 报告）
                       ▼
┌─────────────────────────────────────────────────────┐
│              KAIROS Engine（时机引擎）                 │
│                                                      │
│   ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│   │ Perceiver │  │ Detector │  │ Decider          │  │
│   │ 感知器     │  │ 检测器    │  │ 决策器            │  │
│   │           │  │          │  │                  │  │
│   │ 收集原始   │  │ 识别模式  │  │ 判断：要不要说？  │  │
│   │ 设备数据   │  │ 和异常    │  │ 说什么？怎么说？  │  │
│   └──────────┘  └──────────┘  └──────────────────┘  │
│        ▲              ▲               │              │
│        │              │               ▼              │
│   ┌────┴──────────────┴───────────────────────────┐ │
│   │            Time Window Buffer                  │ │
│   │         滑动时间窗口（去噪 + 聚合）              │ │
│   └───────────────────────────────────────────────┘ │
└──────────────────────┬──────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
   ┌─────────┐  ┌───────────┐  ┌──────────┐
   │ 心跳数据  │  │ 性能指标   │  │ 事件日志  │
   │ Heartbeat│  │ Metrics   │  │ Events   │
   └─────────┘  └───────────┘  └──────────┘
```

---

## 三、三层架构详解

### 3.1 Perceiver（感知器）—— "一直在看"

职责：持续采集设备数据，填充时间窗口缓冲区。

**数据源：**

```typescript
interface PerceptionFrame {
  timestamp: Date;
  deviceId: string;
  
  // 心跳层
  heartbeat: {
    alive: boolean;
    latencyMs: number;          // 心跳延迟
    missedConsecutive: number;  // 连续丢失次数
  };
  
  // 性能层
  metrics: {
    cpuPercent: number;
    memoryPercent: number;
    diskPercent: number;
    networkBytesPerSec: number;
    transactionRate: number;    // 交易速率（每分钟）
  };
  
  // 事件层
  events: {
    type: 'fault' | 'config_change' | 'reboot' | 'transaction_fail';
    severity: LogSeverity;
    message: string;
  }[];
}
```

**时间窗口缓冲区：**

```typescript
// 滑动窗口，保留最近 N 分钟的数据
interface TimeWindowBuffer {
  deviceId: string;
  frames: PerceptionFrame[];  // 按时间排序
  windowMinutes: number;       // 默认 30 分钟
  
  // 聚合方法
  avgCpu(): number;
  avgMemory(): number;
  faultCount(type?: LogType): number;
  trend(metric: string): 'rising' | 'falling' | 'stable';
}
```

---

### 3.2 Detector（检测器）—— "识别出问题"

职责：从感知数据中识别模式和异常。

**检测规则（内置 + 可扩展）：**

```typescript
interface DetectionRule {
  id: string;
  name: string;               // 规则名称
  description: string;
  
  // 触发条件
  condition: (buffer: TimeWindowBuffer, device: Device) => boolean;
  
  // 输出
  generateSignal(buffer: TimeWindowBuffer, device: Device): Signal;
}

interface Signal {
  ruleId: string;
  deviceId: string;
  severity: 'info' | 'warning' | 'critical';
  category: string;           // 分类
  title: string;              // 一句话标题
  detail: string;             // 详细描述
  suggestedAction?: string;   // 建议操作
  detectedAt: Date;
}
```

**内置检测规则示例：**

| 规则 ID | 名称 | 触发条件 | 严重级别 |
|---------|------|----------|----------|
| `heartbeat_drop` | 心跳丢失 | 连续 3 次心跳未响应 | critical |
| `cpu_spike` | CPU 飙高 | 近 10 分钟平均 CPU > 90% | warning |
| `disk_critical` | 磁盘告急 | 磁盘使用率 > 95% | critical |
| `transaction_dry` | 交易异常 | 近 2 小时交易量为 0（设备应在线） | warning |
| `repeated_fault` | 反复故障 | 同一设备 24h 内同一类型故障 ≥ 3 次 | warning |
| `regional_anomaly` | 区域异常 | 同一区域 ≥ 30% 设备同时异常 | critical |
| `firmware_outdated` | 固件落后 | 设备固件版本低于最新版 ≥ 2 个版本 | info |
| `night_reboot_miss` | 夜间重启未执行 | 配置了自动重启但未执行 | info |
| `heartbeat_jitter` | 心跳抖动 | 心跳延迟标准差 > 阈值 | warning |
| `memory_leak` | 内存泄漏嫌疑 | 内存使用率 4h 内持续上升 > 20% | warning |

---

### 3.3 Decider（决策器）—— "要不要说？说什么？"

职责：这是 KAIROS 的核心——判断"恰当的时机"。

**时机判断不只是"有异常就告警"，需要考虑：**

```typescript
interface DecisionContext {
  signal: Signal;
  device: Device;
  
  // 上下文感知
  operatorOnline: boolean;      // 操作员是否在线
  recentAlerts: number;         // 最近 1 小时已发告警数
  similarSignalsToday: number;  // 今日同类信号数
  timeOfDay: 'work' | 'rest' | 'night';  // 当前时段
  
  // 历史判断
  wasRecentlyAlerted: boolean;  // 是否最近已告警过同类问题
  isEscalation: boolean;        // 是否是升级（之前是 warning，现在变 critical）
}

interface Decision {
  action: 'alert_now' | 'batch_later' | 'suppress' | 'escalate';
  channel: 'push' | 'daily_report' | 'weekly_report';
  priority: number;             // 1-10
  message: string;              // 推送内容
}

// 决策逻辑
function decide(ctx: DecisionContext): Decision {
  // 规则 1: critical 且操作员在线 → 立即推送
  if (ctx.signal.severity === 'critical' && ctx.operatorOnline) {
    return { action: 'alert_now', channel: 'push', priority: 9, message: ... };
  }
  
  // 规则 2: critical 但操作员离线 → 升级（短信/电话）
  if (ctx.signal.severity === 'critical' && !ctx.operatorOnline) {
    return { action: 'escalate', channel: 'push', priority: 10, message: ... };
  }
  
  // 规则 3: warning 且今日已发太多告警 → 批量到日报
  if (ctx.signal.severity === 'warning' && ctx.recentAlerts > 5) {
    return { action: 'batch_later', channel: 'daily_report', priority: 3, message: ... };
  }
  
  // 规则 4: 同一设备同一问题已告警过 → 抑制（避免重复骚扰）
  if (ctx.wasRecentlyAlerted && !ctx.isEscalation) {
    return { action: 'suppress', ... };
  }
  
  // 规则 5: 夜间非 critical → 攒到第二天
  if (ctx.timeOfDay === 'night' && ctx.signal.severity !== 'critical') {
    return { action: 'batch_later', channel: 'daily_report', priority: 2, message: ... };
  }
  
  // 默认: 正常推送
  return { action: 'alert_now', channel: 'push', priority: 5, message: ... };
}
```

**防骚扰机制（关键）：**

```
告警疲劳 = KAIROS 的反面

如果 AI 天天主动找你，但 90% 是噪音，你很快就会关掉它。
所以 Decider 的核心不是"什么时候说"，而是"什么时候不说"。

原则：
1. 同一问题不重复告警（除非状态升级）
2. 非紧急不夜间打扰
3. 批量 > 单条（日报汇总低级别告警）
4. 告警必须附带建议操作（不能只说"有问题"）
5. 用户可以标记"不需要此类告警"（学习机制）
```

---

## 四、推送通道设计

```typescript
interface NotificationChannel {
  type: 'wechat' | 'sms' | 'email' | 'webhook' | 'daily_digest';
  
  // 适用条件
  matchSeverity: LogSeverity[];
  matchTime: ('work' | 'rest' | 'night')[];
  
  // 发送
  send(message: NotificationPayload): Promise<void>;
}

// 推送内容模板
interface NotificationPayload {
  title: string;           // "🚨 设备 SH-PD-003 心跳丢失"
  body: string;            // 详细信息
  deviceId: string;
  actionUrl?: string;      // 一键跳转到设备详情
  suggestedAction: string; // "建议：远程重启或安排现场巡检"
  timestamp: Date;
}
```

**通道优先级：**

| 严重级别 | 在线时 | 离线时 |
|----------|--------|--------|
| critical | 微信推送 + 声音 | 短信 + 电话 |
| warning | 微信推送 | 次日日报 |
| info | 日报汇总 | 周报汇总 |

---

## 五、KAIROS 在现有架构中的位置

```
现有架构：
  terminal-agent
  ├── packages/core     # 数据模型
  ├── packages/tools    # Agent 工具（被动：list, status, logs...）
  └── packages/cli      # TUI 交互

新增 KAIROS 层：
  terminal-agent
  ├── packages/kairos/              # ← 新增
  │   ├── src/
  │   │   ├── perceiver.ts          # 感知器：数据采集 + 时间窗口
  │   │   ├── detector.ts           # 检测器：规则引擎
  │   │   ├── decider.ts            # 决策器：时机判断 + 防骚扰
  │   │   ├── notifier.ts           # 推送器：多通道通知
  │   │   ├── rules/                # 内置检测规则
  │   │   │   ├── heartbeat.ts
  │   │   │   ├── performance.ts
  │   │   │   ├── transaction.ts
  │   │   │   └── regional.ts
  │   │   └── daemon.ts             # 守护进程入口
  │   └── package.json
  ├── packages/core     # （不变）
  ├── packages/tools    # （不变）
  └── packages/cli      # （不变，但可接收 KAIROS 推送）
```

---

## 六、与多 Agent 体系的关系

KAIROS 天然适配你规划的多 Agent 路线：

```
第一步（现在）：单 Agent 加固
  └── 用户问 → Agent 答（被动模式）

第二步：加 KAIROS 感知层
  └── KAIROS daemon 常驻 → 检测异常 → 主动推送给用户
  └── 这一步不需要"多 Agent"，是单 Agent 的能力增强

第三步：加 review agent
  └── KAIROS 检测到异常 → spawn review agent 分析根因 → 推送结论
  └── 这时候 KAIROS 是"调度者"，review agent 是"执行者"

远期：完整脑区体系
  └── KAIROS = 脑干（自主神经系统，不需意识参与）
  └── Review Agent = 小脑（协调分析）
  └── Main Agent = 大脑皮层（高级决策）
```

**建议：先做第二步（KAIROS daemon），再做第三步（review agent）。**

理由：
- KAIROS 是独立模块，不依赖多 Agent 基础设施
- KAIROS 本身就是"设备自治"的核心能力
- 做完 KAIROS 后再加 review agent，天然有调度入口

---

## 七、实现路线图

### Phase 1：最小可用 KAIROS（1-2 周）

- [ ] `packages/kairos` 包初始化
- [ ] Perceiver：从现有 `device_status` / `device_logs` 工具读取数据
- [ ] Detector：实现 3 个核心规则（heartbeat_drop, cpu_spike, disk_critical）
- [ ] Decider：基本时机判断（critical 立即推，warning 进日报）
- [ ] Notifier：微信推送通道（复用 OpenClaw 的 channel）
- [ ] Daemon：Node.js 定时轮询（每 60 秒）

### Phase 2：规则丰富 + 防骚扰（2-4 周）

- [ ] 补齐所有内置规则
- [ ] 告警去重和抑制逻辑
- [ ] 日报 / 周报自动生成
- [ ] 用户可配置规则开关和阈值

### Phase 3：多 Agent 集成（4-6 周）

- [ ] KAIROS 信号触发 review agent（spawn）
- [ ] Review agent 返回根因分析 → 合并到告警消息
- [ ] 历史信号存储 + 趋势分析

---

## 八、一个具体场景演示

```
时间线：
08:00  SH-PD-003 正常运行，KAIROS 感知正常
08:15  SH-PD-003 心跳延迟从 50ms 涨到 800ms（Perceiver 记录）
08:17  SH-PD-003 心跳延迟到 2000ms（Detector: heartbeat_jitter 触发 → warning）
08:17  Decider 判断：工作时间 + warning → 微信推送
       
       📱 推送："⚠️ SH-PD-003（浦东张江药房）心跳延迟异常，
              近 2 分钟平均延迟 1400ms。建议检查网络状况。"

08:20  SH-PD-003 心跳完全丢失（Detector: heartbeat_drop 触发 → critical）
08:20  Decider 判断：critical + 操作员在线 → 立即推送

       📱 推送："🚨 SH-PD-003 心跳丢失！连续 3 次未响应。
              建议：远程重启或安排现场巡检。"

08:25  操作员远程重启设备
08:26  SH-PD-003 恢复心跳
08:26  KAIROS 检测到恢复 → 自动关闭该设备的告警

       📱 推送："✅ SH-PD-003 已恢复正常。心跳延迟 45ms。"

---

如果同样的事情发生在凌晨 2:00：
- warning → 抑制，进入次日日报
- critical → 仍然推送（但走短信通道，不走微信）
```

---

## 九、关键设计原则

1. **感知是基础设施，不是功能**——KAIROS 应该像操作系统的后台服务一样存在
2. **时机 > 频率**——宁可少说，不可乱说
3. **每条告警必须有行动建议**——"有问题"不叫告警，叫噪音
4. **学习用户偏好**——被标记为"不需要"的告警类型，降低优先级
5. **本地优先**——数据不出设备，解决信任问题（呼应文章中提到的"本地运行"路径）

---

*文档版本：v0.1 | 2026-04-07 | 初稿*
