// ============================================================
// mock.go - 模拟数据存储
// ============================================================
// 实现了 Store 接口，提供模拟数据
// 用于开发和测试，不需要真实 API
// ============================================================

package store

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/QF1987/terminal-agent-go/internal/device"
)

// ─── 接口定义 ─────────────────────────────────────────────
// Go 用 interface 定义接口（类似 TypeScript 的 interface）
// 任何实现了这些方法的 struct 都算"实现了"这个接口（鸭子类型）
type Store interface {
	ListDevices(f device.DeviceFilters) ([]device.Device, error)          // 列出设备
	GetDevice(id string) (*device.Device, error)                          // 获取单个设备
	GetDeviceStats(id string, days int) (*device.DeviceStats, error)      // 获取设备统计
	GetFaultLogs(f device.LogFilters) ([]device.FaultLog, error)          // 获取故障日志
	UpdateDeviceConfig(id string, config device.DeviceConfig) error       // 更新配置
	UpdateDeviceStatus(id string, status string) error                    // 更新设备状态
	UpdateDeviceCapabilities(id string, cap device.DeviceCapability) error // 更新设备能力
	RebootDevice(id string, force bool) error                             // 重启设备
	// 指令相关
	CreateCommand(cmd device.Command) error                               // 创建指令
	GetCommand(id string) (*device.Command, error)                        // 获取指令（按内部编号）
	GetCommandByUUID(commandID string) (*device.Command, error)           // 获取指令（按 proto UUID）
	ListCommands(f device.CommandFilters) ([]device.Command, error)       // 列出指令
	GetPendingCommands(deviceID string) ([]device.Command, error)         // 获取待下发指令
	UpdateCommandStatus(id string, status string, resultMessage string) error // 更新指令状态
	UpdateCommandResultByUUID(commandID string, status string, message string) error // 设备回报结果
	ExpireTimedOutCommands() (int64, error)                               // 超时指令标记
	// 日志相关
	CreateFaultLog(log device.FaultLog) error                             // 创建故障日志
}

// ─── MockStore 结构体 ─────────────────────────────────────
// 模拟数据存储，内部用切片（slice）存储数据
// 小写字母开头的字段是"私有"的（类似 TypeScript 的 private）
type MockStore struct {
	devices   []device.Device    // 设备列表
	faultLogs []device.FaultLog  // 故障日志列表
}

// NewMockStore()：工厂函数，创建并初始化 MockStore
// Go 没有构造函数，通常用 NewXxx() 函数代替
func NewMockStore() *MockStore {
	s := &MockStore{}  // & 取地址，返回指针
	s.generateData()   // 生成模拟数据
	return s
}

// generateData()：生成模拟数据
// 小写开头的方法是"私有"的，只能在包内部调用
func (s *MockStore) generateData() {
	now := time.Now()

	// 设备模板：名称、区域、地址
	// []struct{...}：匿名结构体切片
	deviceTemplates := []struct {
		name    string
		region  string
		address string
	}{
		// 华东（10 台）
		{"上海浦东-张江药房-01", "华东", "上海市浦东新区张江路100号"},
		{"上海浦东-张江药房-02", "华东", "上海市浦东新区张江路100号"},
		{"上海浦东-陆家嘴药房-01", "华东", "上海市浦东新区陆家嘴环路500号"},
		{"杭州西湖-文三路药房-01", "华东", "杭州市西湖区文三路200号"},
		{"杭州西湖-文三路药房-02", "华东", "杭州市西湖区文三路200号"},
		{"杭州西湖-文三路药房-03", "华东", "杭州市西湖区文三路200号"},
		{"南京鼓楼-中央路药房-01", "华东", "南京市鼓楼区中央路300号"},
		{"南京鼓楼-中央路药房-02", "华东", "南京市鼓楼区中央路300号"},
		{"苏州园区-星海街药房-01", "华东", "苏州市工业园区星海街50号"},
		{"合肥蜀山-长江西路药房-01", "华东", "合肥市蜀山区长江西路100号"},
		// 华南（8 台）
		{"广州天河-体育西路药房-01", "华南", "广州市天河区体育西路100号"},
		{"广州天河-体育西路药房-02", "华南", "广州市天河区体育西路100号"},
		{"深圳南山-科技园药房-01", "华南", "深圳市南山区科技园南路50号"},
		{"深圳南山-科技园药房-02", "华南", "深圳市南山区科技园南路50号"},
		{"深圳福田-华强北路药房-01", "华南", "深圳市福田区华强北路200号"},
		{"珠海香洲-凤凰路药房-01", "华南", "珠海市香洲区凤凰路100号"},
		{"东莞南城-鸿福路药房-01", "华南", "东莞市南城区鸿福路50号"},
		{"佛山禅城-汾江路药房-01", "华南", "佛山市禅城区汾江路100号"},
		// 华北（8 台）
		{"北京朝阳-望京药房-01", "华北", "北京市朝阳区望京西路100号"},
		{"北京朝阳-望京药房-02", "华北", "北京市朝阳区望京西路100号"},
		{"北京海淀-中关村药房-01", "华北", "北京市海淀区中关村大街50号"},
		{"北京海淀-中关村药房-02", "华北", "北京市海淀区中关村大街50号"},
		{"天津河西-友谊路药房-01", "华北", "天津市河西区友谊路100号"},
		{"石家庄长安-中山东路药房-01", "华北", "石家庄市长安区中山东路200号"},
		{"济南历下-泉城路药房-01", "华北", "济南市历下区泉城路100号"},
		{"青岛崂山-海尔路药房-01", "华北", "青岛市崂山区海尔路50号"},
		// 西南（5 台）
		{"成都锦江-春熙路药房-01", "西南", "成都市锦江区春熙路100号"},
		{"成都锦江-春熙路药房-02", "西南", "成都市锦江区春熙路100号"},
		{"重庆渝中-解放碑药房-01", "西南", "重庆市渝中区解放碑步行街50号"},
		{"重庆渝中-解放碑药房-02", "西南", "重庆市渝中区解放碑步行街50号"},
		{"昆明五华-东风西路药房-01", "西南", "昆明市五华区东风西路100号"},
		{"贵阳云岩-中华北路药房-01", "西南", "贵阳市云岩区中华北路50号"},
		// 华中（7 台）
		{"武汉武昌-光谷药房-01", "华中", "武汉市武昌区光谷大道100号"},
		{"武汉武昌-光谷药房-02", "华中", "武汉市武昌区光谷大道100号"},
		{"武汉洪山-街道口药房-01", "华中", "武汉市洪山区街道口50号"},
		{"长沙岳麓-麓山路药房-01", "华中", "长沙市岳麓区麓山路100号"},
		{"长沙岳麓-麓山路药房-02", "华中", "长沙市岳麓区麓山路100号"},
		{"郑州金水-花园路药房-01", "华中", "郑州市金水区花园路200号"},
		{"南昌红谷滩-红谷中大道药房-01", "华中", "南昌市红谷滩区红谷中大道50号"},
	}

	// 状态和类型选项
	statuses := []string{device.StatusOnline, device.StatusOffline, device.StatusError, device.StatusMaintenance}
	types := device.DeviceTypes

	// ─── 生成设备数据 ─────────────────────────────────────
	// for i, tmpl := range：Go 的 for-range 循环（类似 TypeScript 的 entries()）
	for i, tmpl := range deviceTemplates {
		// 随机状态，但大部分是在线的
		status := statuses[rand.Intn(len(statuses))]
		if i < 35 {
			status = device.StatusOnline
		}

		// append()：向切片添加元素（类似 Array.push()）
		s.devices = append(s.devices, device.Device{
			ID:            fmt.Sprintf("DEV-%03d", i+1),  // 格式化字符串，类似模板字符串
			Name:          tmpl.name,
			Type:          types[rand.Intn(len(types))],
			Region:        tmpl.region,
			Address:       tmpl.address,
			Status:        status,
			LastHeartbeat: now.Add(-time.Duration(rand.Intn(3600)) * time.Second), // 随机过去1小时内
			Firmware:      fmt.Sprintf("2.%d.%d", rand.Intn(3), rand.Intn(10)),
			Config: device.DeviceConfig{
				ScreenBrightness:  60 + rand.Intn(40),  // 60-99
				VolumeLevel:       50 + rand.Intn(50),   // 50-99
				AutoRebootEnabled: rand.Intn(2) == 1,    // 随机 true/false
				AutoRebootTime:    fmt.Sprintf("%02d:00", rand.Intn(6)),
				CustomConfigJSON:  fmt.Sprintf(`{"transaction_timeout":30,"medicine_category":["处方药","OTC","保健品"]}`),
			},
			InstalledAt: now.AddDate(-1, -rand.Intn(12), 0), // 随机过去1年内安装
			Stats: device.DeviceStats{
				Uptime:          100 + rand.Intn(8000),
				FaultCount:      rand.Intn(20),
				CustomStatsJSON: fmt.Sprintf(`{"total_transactions":%d,"today_transactions":%d}`, 1000+rand.Intn(9000), rand.Intn(200)),
			},
		})
	}

	// ─── 生成故障日志数据 ─────────────────────────────────
	for i := 0; i < 100; i++ {
		dev := s.devices[rand.Intn(len(s.devices))]
		severities := []string{device.SeverityLow, device.SeverityMedium, device.SeverityHigh, device.SeverityCritical}
		logTypes := []string{device.LogHardware, device.LogSoftware, device.LogNetwork, device.LogMedicineStock}

		// 故障消息模板（按类型分类）
		messages := map[string][]string{  // map：字典（类似 TypeScript 的 Record<string, string[]>）
			device.LogHardware: {
				"打印机卡纸，已自动恢复",
				"扫码器响应超时",
				"触摸屏校准偏移",
				"出药机构卡顿",
			},
			device.LogSoftware: {
				"应用崩溃，已自动重启",
				"数据库连接超时",
				"内存使用率过高(85%)",
				"系统更新失败",
			},
			device.LogNetwork: {
				"网络模块故障",
				"4G信号弱，切换至备用网络",
				"VPN连接中断",
				"心跳包丢失超过阈值",
			},
			device.LogMedicineStock: {
				"布洛芬库存低于警戒线",
				"板蓝根即将过期",
				"感冒灵库存为0",
				"药品补货延迟",
			},
		}

		logType := logTypes[rand.Intn(len(logTypes))]
		msgs := messages[logType]
		msg := msgs[rand.Intn(len(msgs))]

		// 随机是否已解决
		resolved := rand.Intn(3) > 0  // 2/3 概率已解决
		var resolvedAt *time.Time       // 指针类型，可以是 nil
		if resolved {
			t := now.Add(-time.Duration(rand.Intn(24)) * time.Hour)
			resolvedAt = &t  // & 取地址
		}

		s.faultLogs = append(s.faultLogs, device.FaultLog{
			ID:         fmt.Sprintf("LOG-%03d", i+1),
			DeviceID:   dev.ID,
			Timestamp:  now.Add(-time.Duration(rand.Intn(720)) * time.Hour), // 过去30天内
			Type:       logType,
			Severity:   severities[rand.Intn(len(severities))],
			Message:    msg,
			Resolved:   resolved,
			ResolvedAt: resolvedAt,
		})
	}
}

// ─── 接口方法实现 ─────────────────────────────────────────
// 大写开头的方法是"公开"的（类似 TypeScript 的 public）

// ListDevices：列出设备（支持筛选）
func (s *MockStore) ListDevices(f device.DeviceFilters) ([]device.Device, error) {
	var result []device.Device  // var 声明零值切片
	for _, d := range s.devices {  // _ 忽略索引
		// 筛选逻辑：如果指定了条件且不匹配，跳过
		if f.Region != "" && d.Region != f.Region {
			continue
		}
		if f.Status != "" && d.Status != f.Status {
			continue
		}
		if f.Type != "" && d.Type != f.Type {
			continue
		}
		if f.Keyword != "" {
			// 关键字搜索：在名称、地址、ID 中查找
			found := false
			for _, field := range []string{d.Name, d.Address, d.ID} {
				if contains(field, f.Keyword) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		result = append(result, d)
	}
	return result, nil  // Go 用返回值代替异常（类似 Go 的 error 模式）
}

// GetDevice：获取单个设备
func (s *MockStore) GetDevice(id string) (*device.Device, error) {
	for _, d := range s.devices {
		if d.ID == id {
			return &d, nil  // 返回指针（避免复制整个结构体）
		}
	}
	return nil, fmt.Errorf("设备不存在: %s", id)  // 返回错误
}

// GetDeviceStats：获取设备统计
func (s *MockStore) GetDeviceStats(id string, days int) (*device.DeviceStats, error) {
	dev, err := s.GetDevice(id)  // 调用其他方法
	if err != nil {
		return nil, err  // 错误传播
	}
	return &dev.Stats, nil
}

// GetFaultLogs：获取故障日志（支持筛选）
func (s *MockStore) GetFaultLogs(f device.LogFilters) ([]device.FaultLog, error) {
	var result []device.FaultLog
	limit := f.Limit
	if limit <= 0 {
		limit = 20  // 默认返回 20 条
	}

	for _, log := range s.faultLogs {
		if f.DeviceID != "" && log.DeviceID != f.DeviceID {
			continue
		}
		if f.Severity != "" && log.Severity != f.Severity {
			continue
		}
		if f.Type != "" && log.Type != f.Type {
			continue
		}
		if f.Days > 0 && time.Since(log.Timestamp).Hours() > float64(f.Days*24) {
			continue
		}
		result = append(result, log)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

// UpdateDeviceConfig：更新设备配置（模拟）
func (s *MockStore) UpdateDeviceConfig(id string, config device.DeviceConfig) error {
	for i, d := range s.devices {
		if d.ID == id {
			s.devices[i].Config = config  // 修改切片元素
			return nil
		}
	}
	return fmt.Errorf("设备不存在: %s", id)
}

// RebootDevice：重启设备（模拟）
func (s *MockStore) RebootDevice(id string, force bool) error {
	_, err := s.GetDevice(id)
	if err != nil {
		return err
	}
	// 模拟重启，实际什么都不做
	return nil
}

// ─── 辅助函数 ─────────────────────────────────────────────

// contains：字符串包含检查（Go 标准库没有 strings.Contains 的时候用）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ─── 指令相关方法（MockStore 实现） ─────────────────────────

// MockStore 的指令存储
// 注意：这只是内存实现，重启后数据丢失
type mockCommandStore struct {
	commands []device.Command
}

var mockCmdStore = &mockCommandStore{}

func (s *MockStore) CreateCommand(cmd device.Command) error {
	mockCmdStore.commands = append(mockCmdStore.commands, cmd)
	return nil
}

func (s *MockStore) GetCommand(id string) (*device.Command, error) {
	for _, cmd := range mockCmdStore.commands {
		if cmd.ID == id {
			return &cmd, nil
		}
	}
	return nil, fmt.Errorf("指令不存在: %s", id)
}

func (s *MockStore) GetCommandByUUID(commandID string) (*device.Command, error) {
	for _, cmd := range mockCmdStore.commands {
		if cmd.CommandID == commandID {
			return &cmd, nil
		}
	}
	return nil, fmt.Errorf("指令不存在: %s", commandID)
}

func (s *MockStore) ListCommands(f device.CommandFilters) ([]device.Command, error) {
	var result []device.Command
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	for _, cmd := range mockCmdStore.commands {
		if f.DeviceID != "" && cmd.DeviceID != f.DeviceID {
			continue
		}
		if f.Status != "" && cmd.Status != f.Status {
			continue
		}
		if f.Type != "" && cmd.CommandType != f.Type {
			continue
		}
		result = append(result, cmd)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (s *MockStore) GetPendingCommands(deviceID string) ([]device.Command, error) {
	return s.ListCommands(device.CommandFilters{
		DeviceID: deviceID,
		Status:   device.CommandStatusPending,
	})
}

func (s *MockStore) UpdateCommandStatus(id string, status string, resultMessage string) error {
	for i, cmd := range mockCmdStore.commands {
		if cmd.ID == id {
			mockCmdStore.commands[i].Status = status
			mockCmdStore.commands[i].ResultMessage = resultMessage
			now := time.Now()
			if status == device.CommandStatusSent {
				mockCmdStore.commands[i].SentAt = &now
			} else {
				mockCmdStore.commands[i].ExecutedAt = &now
			}
			mockCmdStore.commands[i].UpdatedAt = now
			return nil
		}
	}
	return fmt.Errorf("指令不存在: %s", id)
}

func (s *MockStore) UpdateCommandResultByUUID(commandID string, status string, message string) error {
	for i, cmd := range mockCmdStore.commands {
		if cmd.CommandID == commandID {
			mockCmdStore.commands[i].Status = status
			mockCmdStore.commands[i].ResultMessage = message
			now := time.Now()
			mockCmdStore.commands[i].ExecutedAt = &now
			mockCmdStore.commands[i].UpdatedAt = now
			return nil
		}
	}
	return fmt.Errorf("指令不存在: %s", commandID)
}

func (s *MockStore) ExpireTimedOutCommands() (int64, error) {
	var count int64
	now := time.Now()
	for i, cmd := range mockCmdStore.commands {
		if cmd.Status == device.CommandStatusSent && cmd.SentAt != nil {
			deadline := cmd.SentAt.Add(time.Duration(cmd.TimeoutSeconds) * time.Second)
			if now.After(deadline) {
				mockCmdStore.commands[i].Status = device.CommandStatusTimeout
				mockCmdStore.commands[i].ResultMessage = "服务端检测超时"
				mockCmdStore.commands[i].UpdatedAt = now
				count++
			}
		}
	}
	return count, nil
}

func (s *MockStore) UpdateDeviceStatus(id string, status string) error {
	for i, d := range s.devices {
		if d.ID == id {
			s.devices[i].Status = status
			s.devices[i].LastHeartbeat = time.Now()
			return nil
		}
	}
	return fmt.Errorf("设备不存在: %s", id)
}

func (s *MockStore) UpdateDeviceCapabilities(id string, cap device.DeviceCapability) error {
	for i, d := range s.devices {
		if d.ID == id {
			s.devices[i].Capabilities = cap
			return nil
		}
	}
	return fmt.Errorf("设备不存在: %s", id)
}

func (s *MockStore) CreateFaultLog(log device.FaultLog) error {
	s.faultLogs = append(s.faultLogs, log)
	return nil
}
