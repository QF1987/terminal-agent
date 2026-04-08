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

// DeviceConfig：设备配置
type DeviceConfig struct {
	TransactionTimeout int      `json:"transactionTimeout"` // 交易超时（秒）
	ScreenBrightness   int      `json:"screenBrightness"`   // 屏幕亮度（%）
	VolumeLevel        int      `json:"volumeLevel"`        // 音量（%）
	AutoRebootEnabled  bool     `json:"autoRebootEnabled"`  // 是否启用自动重启
	AutoRebootTime     string   `json:"autoRebootTime"`     // 自动重启时间
	MedicineCategory   []string `json:"medicineCategory"`   // 药品分类
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
	FirmwareVersion   string   `json:"firmware_version"`   // 固件版本号
	ProtoVersion      int      `json:"proto_version"`      // 协议版本号
	SupportedFeatures []string `json:"supported_features"` // 支持的功能列表
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

// ─── 指令相关常量 ─────────────────────────────────────────
const (
	CommandTypeReboot         = "reboot"           // 重启设备
	CommandTypeUpdateConfig   = "update_config"    // 修改配置
	CommandTypeUpgradeFirmware = "upgrade_firmware" // OTA 固件升级
	CommandTypeCustom         = "custom"           // 自定义指令
)

const (
	CommandStatusPending   = "pending"    // 待下发
	CommandStatusSent      = "sent"       // 已推送给设备
	CommandStatusExecuting = "executing"  // 设备执行中
	CommandStatusCompleted = "completed"  // 执行成功
	CommandStatusFailed    = "failed"     // 执行失败
	CommandStatusTimeout   = "timeout"    // 执行超时
	CommandStatusRejected  = "rejected"   // 设备拒绝执行
)

// Command：指令数据模型
type Command struct {
	ID             string     `json:"id"`             // 内部编号 CMD-001
	CommandID      string     `json:"commandId"`      // UUID，与 proto 一致，设备靠这个回报
	DeviceID       string     `json:"deviceId"`       // 目标设备
	CommandType    string     `json:"commandType"`    // 指令类型
	PayloadJSON    string     `json:"payloadJson"`    // 指令参数（JSON 字符串）
	Status         string     `json:"status"`         // 指令状态
	TimeoutSeconds int        `json:"timeoutSeconds"` // 执行超时（秒）
	IssuedAt       time.Time  `json:"issuedAt"`       // 服务端下发时间
	SentAt         *time.Time `json:"sentAt,omitempty"`    // 实际推送给设备的时间
	ExecutedAt     *time.Time `json:"executedAt,omitempty"` // 设备回报执行完成时间
	ResultMessage  string     `json:"resultMessage"`  // 设备回报的执行结果描述
	CreatedBy      string     `json:"createdBy"`      // 谁发起的（操作员/API）
	CreatedAt      time.Time  `json:"createdAt"`      // 记录创建时间
	UpdatedAt      time.Time  `json:"updatedAt"`      // 记录更新时间
}

// CommandFilters：指令查询筛选条件
type CommandFilters struct {
	DeviceID string // 设备ID
	Status   string // 指令状态
	Type     string // 指令类型
	Limit    int    // 返回条数
}
