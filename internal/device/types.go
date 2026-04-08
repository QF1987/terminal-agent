// ============================================================
// types.go - 数据模型定义
// ============================================================
// 定义了所有数据结构（struct）和常量
// 类似 TypeScript 的 interface + enum
// ============================================================

package device

import "time"

// ─── 设备状态常量 ─────────────────────────────────────────
// Go 用 const 定义常量，没有 enum，通常用字符串或 iota
const (
	StatusOnline       = "online"       // 在线
	StatusOffline      = "offline"      // 离线
	StatusError        = "error"        // 故障
	StatusMaintenance  = "maintenance"  // 维护中
)

// ─── 区域和设备类型 ───────────────────────────────────────
// var：定义变量（这里是全局只读的切片）
// []string：字符串切片（类似 TypeScript 的 string[]）
var Regions = []string{"华东", "华南", "华北", "西南", "华中"}
var DeviceTypes = []string{"自助购药机-标准版", "自助购药机-冷藏版", "自助购药机-大型版"}

// ─── 日志相关常量 ─────────────────────────────────────────
const (
	LogHardware      = "hardware"       // 硬件故障
	LogSoftware      = "software"       // 软件故障
	LogNetwork       = "network"        // 网络故障
	LogMedicineStock = "medicine_stock" // 库存告警
)

const (
	SeverityLow       = "low"       // 低
	SeverityMedium    = "medium"    // 中
	SeverityHigh      = "high"      // 高
	SeverityCritical  = "critical"  // 严重
)

// ─── 数据结构定义 ─────────────────────────────────────────
// Go 用 struct 定义结构体，类似 TypeScript 的 interface
// 字段后面的 `json:"xxx"` 是 JSON 序列化标签（类似 TS 的 @JsonProperty）

// DeviceConfig：设备配置（对应 proto DeviceConfig）
// 只保留所有终端类型通用的配置项，业务相关配置放 CustomConfigJSON
type DeviceConfig struct {
	ScreenBrightness  int    `json:"screenBrightness"`  // 屏幕亮度，0-100
	VolumeLevel       int    `json:"volumeLevel"`       // 音量，0-100
	AutoRebootEnabled bool   `json:"autoRebootEnabled"` // 是否开启自动重启
	AutoRebootTime    string `json:"autoRebootTime"`    // 自动重启时间 "HH:MM"
	CustomConfigJSON  string `json:"customConfigJson"`  // 业务自定义配置（JSON 字符串）
}

// DeviceMetrics：设备性能指标（对应 proto DeviceMetrics）
// 用于 StatusReport 和未来的 KAIROS 感知层
type DeviceMetrics struct {
	CPUPercent           float32 `json:"cpuPercent"`           // CPU 使用率，0.0-100.0
	MemoryPercent        float32 `json:"memoryPercent"`        // 内存使用率，0.0-100.0
	DiskPercent          float32 `json:"diskPercent"`          // 磁盘使用率，0.0-100.0
	NetworkRxBytes       int64   `json:"networkRxBytes"`       // 累计网络接收字节
	NetworkTxBytes       int64   `json:"networkTxBytes"`       // 累计网络发送字节
	TransactionCountToday int32  `json:"transactionCountToday"` // 今日交易笔数
	UptimeSeconds        int32   `json:"uptimeSeconds"`        // 运行时长（秒）
}

// DeviceStats：设备统计
type DeviceStats struct {
	TotalTransactions int `json:"totalTransactions"` // 总交易数
	TodayTransactions int `json:"todayTransactions"` // 今日交易数
	Uptime            int `json:"uptime"`            // 运行时长（小时）
	FaultCount        int `json:"faultCount"`        // 故障次数
}

// DeviceCapability：设备能力声明（对应 proto DeviceCapability）
type DeviceCapability struct {
	FirmwareVersion   string   `json:"firmwareVersion"`   // 固件版本号
	ProtoVersion      int32    `json:"protoVersion"`      // 协议版本号
	SupportedFeatures []string `json:"supportedFeatures"` // 支持的功能列表
}

// Device：设备信息（主数据模型）
type Device struct {
	ID            string           `json:"id"`            // 设备ID
	Name          string           `json:"name"`          // 设备名称
	Type          string           `json:"type"`          // 设备类型
	Region        string           `json:"region"`        // 所属区域
	Address       string           `json:"address"`       // 安装地址
	Status        string           `json:"status"`        // 当前状态
	LastHeartbeat time.Time        `json:"lastHeartbeat"` // 最后心跳时间
	Firmware      string           `json:"firmware"`      // 固件版本
	Config        DeviceConfig     `json:"config"`        // 设备配置
	InstalledAt   time.Time        `json:"installedAt"`   // 安装时间
	Stats         DeviceStats      `json:"stats"`         // 运行统计
	Token         string           `json:"token"`         // 设备令牌（第一阶段认证）
	DeviceSecret  string           `json:"deviceSecret"`  // 设备密钥（预留给 HMAC 签名阶段）
	Capabilities  DeviceCapability `json:"capabilities"`  // 设备能力声明
}

// FaultLog：故障日志
type FaultLog struct {
	ID         string     `json:"id"`                   // 日志ID
	DeviceID   string     `json:"deviceId"`             // 设备ID
	Timestamp  time.Time  `json:"timestamp"`            // 发生时间
	Type       string     `json:"type"`                 // 故障类型
	Severity   string     `json:"severity"`             // 严重程度
	Message    string     `json:"message"`              // 故障描述
	Resolved   bool       `json:"resolved"`             // 是否已解决
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"` // 解决时间（指针，可为 nil）
}

// ─── 筛选条件 ─────────────────────────────────────────────

// DeviceFilters：设备查询筛选条件
type DeviceFilters struct {
	Region  string // 区域
	Status  string // 状态
	Type    string // 类型
	Keyword string // 关键字
}

// LogFilters：日志查询筛选条件
type LogFilters struct {
	DeviceID string // 设备ID
	Severity string // 严重程度
	Type     string // 日志类型
	Days     int    // 最近几天
	Limit    int    // 返回条数
}

// ─── 心跳 & 状态上报（对应 proto 请求/响应） ──────────────

// HeartbeatRequest：心跳请求（对应 proto HeartbeatRequest）
type HeartbeatRequest struct {
	DeviceID      string           `json:"deviceId"`      // 设备唯一标识
	Timestamp     int64            `json:"timestamp"`     // 设备本地时间毫秒时间戳
	CPUPercent    float32          `json:"cpuPercent"`    // CPU 使用率
	MemoryPercent float32          `json:"memoryPercent"` // 内存使用率
	DiskPercent   float32          `json:"diskPercent"`   // 磁盘使用率
	UptimeSeconds int32            `json:"uptimeSeconds"` // 累计运行时长（秒）
	Capability    *DeviceCapability `json:"capability,omitempty"` // 设备能力声明（可选）
}

// HeartbeatResponse：心跳响应（对应 proto HeartbeatResponse）
type HeartbeatResponse struct {
	HasPendingCommand bool  `json:"hasPendingCommand"` // 是否有待执行指令
	ServerTime        int64 `json:"serverTime"`        // 服务端当前时间毫秒时间戳
}

// StatusReport：状态上报（对应 proto StatusReport）
type StatusReport struct {
	DeviceID        string       `json:"deviceId"`        // 设备唯一标识
	Timestamp       int64        `json:"timestamp"`       // 上报时间毫秒时间戳
	Status          string       `json:"status"`          // 设备状态 online/offline/error/maintenance
	FirmwareVersion string       `json:"firmwareVersion"` // 当前固件版本号
	Metrics         DeviceMetrics `json:"metrics"`        // 性能指标快照
	Config          DeviceConfig `json:"config"`          // 当前生效的配置
}

// StatusReportResponse：状态上报响应（对应 proto StatusReportResponse）
type StatusReportResponse struct {
	Accepted bool   `json:"accepted"` // 是否接受此次上报
	Message  string `json:"message"`  // 结果描述
}

// ─── 事件上报（对应 proto EventReport） ──────────────────

// 事件类型常量
const (
	EventFault            = "fault"             // 设备故障
	EventTransactionFail  = "transaction_fail"  // 交易失败
	EventConfigChange     = "config_change"     // 配置变更
	EventReboot           = "reboot"            // 设备重启
	EventStockAlert       = "stock_alert"       // 库存告警
	EventCustom           = "custom"            // 自定义事件
)

// 事件严重程度常量（复用已有的 SeverityLow/Medium/High/Critical）

// EventReport：事件上报（对应 proto EventReport）
type EventReport struct {
	DeviceID  string `json:"deviceId"`  // 设备唯一标识
	Timestamp int64  `json:"timestamp"` // 事件发生时间毫秒时间戳
	EventType string `json:"eventType"` // 事件类型
	Severity  string `json:"severity"`  // 严重程度
	Message   string `json:"message"`   // 事件描述
	DetailJSON string `json:"detailJson"` // 扩展详情（JSON 字符串）
}

// EventReportResponse：事件上报响应（对应 proto EventReportResponse）
type EventReportResponse struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
}

// ─── 指令（对应 proto Command + CommandResult） ──────────

// 指令类型常量
const (
	CommandTypeReboot          = "reboot"           // 重启设备
	CommandTypeUpdateConfig    = "update_config"    // 修改配置
	CommandTypeUpgradeFirmware = "upgrade_firmware" // OTA 固件升级
	CommandTypeCustom          = "custom"           // 自定义指令
)

// 指令状态常量
const (
	CommandStatusPending   = "pending"    // 待下发
	CommandStatusSent      = "sent"       // 已推送给设备
	CommandStatusExecuting = "executing"  // 设备执行中
	CommandStatusCompleted = "completed"  // 执行成功
	CommandStatusFailed    = "failed"     // 执行失败
	CommandStatusTimeout   = "timeout"    // 执行超时
	CommandStatusRejected  = "rejected"   // 设备拒绝执行
)

// Command：指令数据模型（对应 proto Command）
type Command struct {
	ID             string     `json:"id"`             // 内部编号 CMD-001
	CommandID      string     `json:"commandId"`      // UUID，与 proto 一致
	DeviceID       string     `json:"deviceId"`       // 目标设备
	CommandType    string     `json:"commandType"`    // 指令类型
	PayloadJSON    string     `json:"payloadJson"`    // 指令参数（JSON 字符串）
	Status         string     `json:"status"`         // 指令状态
	TimeoutSeconds int64      `json:"timeoutSeconds"` // 执行超时（秒），proto 是 int64
	IssuedAt       int64      `json:"issuedAt"`       // 服务端下发时间毫秒时间戳（proto 是 int64）
	SentAt         *time.Time `json:"sentAt,omitempty"`
	ExecutedAt     *time.Time `json:"executedAt,omitempty"`
	ResultMessage  string     `json:"resultMessage"`
	CreatedBy      string     `json:"createdBy"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// CommandResult：指令执行结果（对应 proto CommandResult）
type CommandResult struct {
	CommandID  string `json:"commandId"`  // 指令 ID
	DeviceID   string `json:"deviceId"`   // 设备 ID
	Status     string `json:"status"`     // success/failed/timeout/rejected
	Message    string `json:"message"`    // 执行结果描述
	ExecutedAt int64  `json:"executedAt"` // 执行完成时间毫秒时间戳
}

// CommandResultResponse：指令执行结果响应（对应 proto CommandResultResponse）
type CommandResultResponse struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
}

// CommandFilters：指令查询筛选条件
type CommandFilters struct {
	DeviceID string // 设备ID
	Status   string // 指令状态
	Type     string // 指令类型
	Limit    int    // 返回条数
}
