# terminal-agent gRPC 通信设计

> 讨论日期：2026-04-07
> 状态：设计阶段，待实施

---

## 一、协议选型结论

**统一用 gRPC，一套协议搞定所有场景。**

- 状态上送：gRPC unary RPC
- 远程控制：gRPC server streaming（CommandStream）
- 不引入 MQTT，避免双协议的维护成本
- 设备量级（几十到几百台），gRPC 完全够用

**决策背景：**
- 之前项目用 gRPC（上送）+ MQTT（控制），是 EMQX 工程便利驱动，非技术需求驱动
- terminal-agent 团队小、设备量少，不需要 MQTT 的离线队列和 QoS 能力
- gRPC server streaming 可以替代 MQTT 的推送能力

---

## 二、服务定义

```protobuf
syntax = "proto3";
package terminal_agent.v1;

// ─── 设备端 → 服务端 ───────────────────────────────────

service DeviceService {
  // 心跳（轻量，每 30-60 秒）
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  
  // 状态上报（完整快照，每 5 分钟）
  rpc ReportStatus(StatusReport) returns (ReportResponse);
  
  // 事件上报（故障/交易，实时）
  rpc ReportEvent(EventReport) returns (ReportResponse);
}

// ─── 服务端 → 设备端 ───────────────────────────────────

service CommandService {
  // 指令流（server streaming，设备连上挂着，服务端有指令就推）
  rpc CommandStream(CommandStreamRequest) returns (stream Command);
}

// ─── 消息定义 ──────────────────────────────────────────

message HeartbeatRequest {
  string device_id = 1;
  int64  timestamp = 2;
  // 轻量指标（不传完整状态）
  float  cpu_percent = 3;
  float  memory_percent = 4;
  float  disk_percent = 5;
  int32  uptime_seconds = 6;
}

message HeartbeatResponse {
  bool   has_command = 1;   // 提示设备：有待执行指令（可选优化）
  int64  server_time = 2;   // 时钟同步
}

message StatusReport {
  string device_id = 1;
  int64  timestamp = 2;
  string status = 3;         // online / offline / error / maintenance
  string firmware_version = 4;
  DeviceMetrics metrics = 5;
  DeviceConfig config = 6;
}

message DeviceMetrics {
  float  cpu_percent = 1;
  float  memory_percent = 2;
  float  disk_percent = 3;
  int64  network_rx_bytes = 4;
  int64  network_tx_bytes = 5;
  int32  transaction_count_today = 6;
  int32  uptime_seconds = 7;
}

message DeviceConfig {
  int32  transaction_timeout = 1;
  int32  screen_brightness = 2;
  int32  volume_level = 3;
  bool   auto_reboot_enabled = 4;
  string auto_reboot_time = 5;
  repeated string medicine_category = 6;
}

message EventReport {
  string device_id = 1;
  int64  timestamp = 2;
  string event_type = 3;     // fault / transaction_fail / config_change / reboot
  string severity = 4;       // low / medium / high / critical
  string message = 5;
  string detail_json = 6;    // 扩展字段，JSON 格式
}

message ReportResponse {
  bool   accepted = 1;
  string message = 2;
}

// ─── 指令相关 ──────────────────────────────────────────

message CommandStreamRequest {
  string device_id = 1;
}

message Command {
  string command_id = 1;
  string command_type = 2;   // reboot / update_config / upgrade_firmware / custom
  string payload_json = 3;   // 指令参数，JSON 格式
  int64  issued_at = 4;
  int64  timeout_seconds = 5;
}
```

---

## 三、认证方案（分阶段）

### 第一阶段：Token（现在用）

```
设备注册时：
  服务端生成唯一 token（64 字符随机）
  存入 devices 表的 token 字段
  通过安全通道下发到设备

设备认证：
  gRPC metadata 带 token
  服务端 interceptor 验证
  合法 → 放行；非法 → UNAUTHENTICATED
```

**优点：** 最简单，所有平台零依赖
**缺点：** 存在设备上可被提取（反编译/文件读取）

### 第二阶段：HMAC 签名（业务成熟后升级）

```
设备注册时：
  服务端签发 device_id + device_secret（64 字符随机）

设备每次请求：
  signature = HMAC-SHA256(secret, device_id + timestamp + nonce)
  放在 gRPC metadata 的 AuthContext 里

服务端验证：
  1. timestamp 在合理范围内（防重放）
  2. 签名用存储的 device_secret 验证通过
```

**优点：** 不直接传 secret，跨平台零依赖（HMAC 所有语言都有标准库）
**缺点：** 比 token 多几行代码

### 第三阶段：设备证书（设备量大 + 安全合规要求高时）

```
- Android/iOS：私钥存 Keystore/Secure Enclave
- Linux/嵌入式：TPM 模块
- 最终演进到 mTLS
```

**proto 中的认证消息（第二阶段用）：**

```protobuf
message AuthContext {
  string device_id = 1;
  int64  timestamp = 2;        // 毫秒时间戳
  string nonce = 3;            // 随机字符串，防重放
  string signature = 4;        // HMAC-SHA256(secret, device_id + timestamp + nonce)
}
```

---

## 四、指令下发模型

**用 Push（gRPC server streaming），不用 Pull（轮询）。**

```
设备启动 → 连接 CommandStream → 挂着等
服务端有指令 → 通过 stream 推送
设备执行 → 返回结果（通过 ReportEvent 上报）
```

**断线处理：**
- 设备端：gRPC 自动重连 + 指数退避
- 重连后：服务端推送断线期间积压的指令（从数据库查 pending commands）
- 作为 fallback：可加一个 `FetchPendingCommands` unary RPC，stream 断了临时用

**Push vs Pull 对比（100 台设备）：**

| | Pull | Push |
|---|---|---|
| 实时性 | 5-10秒延迟 | 秒级 |
| 连接数 | 每秒 20 个短连接 | 100 个常驻连接 |
| 内存 | 几乎不占 | ~5-10 MB |
| 带宽 | 空轮询浪费 | 静默零开销 |

---

## 五、proto 版本策略

**原则：向后兼容，增量演进。**

- 新增字段：用新字段编号，旧设备忽略未知字段（protobuf 默认行为）
- 不删字段、不改字段编号、不改字段类型
- 废弃字段：标记 `reserved`，不复用编号
- 大版本变更：新建 proto package（`terminal_agent.v2`）

```
兼容示例：
  v1.0: HeartbeatRequest 有 device_id, timestamp
  v1.1: 新增 cpu_percent, memory_percent → 旧设备不传，服务端处理默认值
  v2.0: 消息结构大改 → 新建 v2 package，两套并存
```

### 能力协商机制

**问题：proto 字段向后兼容只解决了"能不能通信"，没解决"功能能不能用"。**

比如 v1.1 新增了 cpu/metrics 上报，但 v1.0 设备不传这些字段。服务端的 KAIROS 检测规则如果依赖 metrics，v1.0 设备就永远触发不了。

**解法：设备声明自己支持哪些能力，服务端按能力差异化处理。**

```protobuf
message DeviceCapability {
  string firmware_version = 1;          // "v2.3.1"
  int32  proto_version = 2;             // 协议版本号，如 1
  repeated string supported_features = 3;
  // 可选值：
  //   "heartbeat_basic"      - 基础心跳（只报 alive）
  //   "heartbeat_metrics"    - 心跳带 CPU/内存/磁盘指标
  //   "event_report"         - 事件上报（故障/交易）
  //   "status_report"        - 完整状态快照
  //   "config_push"          - 接收远程配置下发
  //   "command_reboot"       - 支持远程重启
  //   "command_firmware"     - 支持 OTA 固件升级
}
```

**设备上报时机：**
- 设备注册时 / 每次心跳时 / 固件升级后带上 DeviceCapability
- 服务端存入 devices 表的 capabilities 字段

**服务端使用方式：**
```
if device supports "heartbeat_metrics":
    KAIROS 检测规则 cpu_spike、memory_leak 生效
else:
    跳过，不触发告警

if device supports "config_push":
    可下发配置变更
else:
    提示"该设备固件版本过低，不支持远程配置"
```

**好处：**
- 新功能上线不用等所有设备升级
- 设备升级固件 → 声明新能力 → 服务端自动解锁对应功能
- 前端/UI 可以根据能力展示可用操作，避免用户点了没反应

---

## 六、跨平台兼容性

**终端支持矩阵：**
- Android (ARM)
- iOS (ARM)
- Linux x86 / ARM
- 嵌入式 Linux (ARM/MIPS)
- Windows x86

**gRPC 跨平台支持：**
- 全部都有官方或成熟的 gRPC 库
- Go、C++、Java/Kotlin、C# 等均有支持
- proto 文件统一，各平台生成各自代码

**认证存储按阶段：**
- 第一阶段（token）：所有平台存文件/app 私有目录，统一简单
- 第二阶段（HMAC）：所有平台标准库实现，零平台依赖
- 第三阶段（证书）：按平台选 Keystore/Secure Enclave/TPM/DPAPI

---

*文档版本：v0.1 | 2026-04-07 | 设计讨论记录*
