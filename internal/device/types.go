package device

import "time"

// 设备状态
const (
	StatusOnline       = "online"
	StatusOffline      = "offline"
	StatusError        = "error"
	StatusMaintenance  = "maintenance"
)

// 区域
var Regions = []string{"华东", "华南", "华北", "西南", "华中"}

// 设备类型
var DeviceTypes = []string{"自助购药机-标准版", "自助购药机-冷藏版", "自助购药机-大型版"}

// 日志类型
const (
	LogHardware     = "hardware"
	LogSoftware     = "software"
	LogNetwork      = "network"
	LogMedicineStock = "medicine_stock"
)

// 日志严重程度
const (
	SeverityLow       = "low"
	SeverityMedium    = "medium"
	SeverityHigh      = "high"
	SeverityCritical  = "critical"
)

type DeviceConfig struct {
	TransactionTimeout int      `json:"transactionTimeout"`
	ScreenBrightness   int      `json:"screenBrightness"`
	VolumeLevel        int      `json:"volumeLevel"`
	AutoRebootEnabled  bool     `json:"autoRebootEnabled"`
	AutoRebootTime     string   `json:"autoRebootTime"`
	MedicineCategory   []string `json:"medicineCategory"`
}

type DeviceStats struct {
	TotalTransactions int `json:"totalTransactions"`
	TodayTransactions int `json:"todayTransactions"`
	Uptime            int `json:"uptime"`      // 小时
	FaultCount        int `json:"faultCount"`
}

type Device struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Type          string       `json:"type"`
	Region        string       `json:"region"`
	Address       string       `json:"address"`
	Status        string       `json:"status"`
	LastHeartbeat time.Time    `json:"lastHeartbeat"`
	Firmware      string       `json:"firmware"`
	Config        DeviceConfig `json:"config"`
	InstalledAt   time.Time    `json:"installedAt"`
	Stats         DeviceStats  `json:"stats"`
}

type FaultLog struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"deviceId"`
	Timestamp  time.Time `json:"timestamp"`
	Type       string    `json:"type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Resolved   bool      `json:"resolved"`
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
}

// 筛选条件
type DeviceFilters struct {
	Region  string
	Status  string
	Type    string
	Keyword string
}

type LogFilters struct {
	DeviceID string
	Severity string
	Type     string
	Days     int
	Limit    int
}
