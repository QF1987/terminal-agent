package store

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/QF1987/terminal-agent-go/internal/device"
)

// Store 数据存储接口
type Store interface {
	ListDevices(f device.DeviceFilters) ([]device.Device, error)
	GetDevice(id string) (*device.Device, error)
	GetDeviceStats(id string, days int) (*device.DeviceStats, error)
	GetFaultLogs(f device.LogFilters) ([]device.FaultLog, error)
	UpdateDeviceConfig(id string, config device.DeviceConfig) error
	RebootDevice(id string, force bool) error
}

// MockStore 模拟数据存储
type MockStore struct {
	devices   []device.Device
	faultLogs []device.FaultLog
}

func NewMockStore() *MockStore {
	s := &MockStore{}
	s.generateData()
	return s
}

func (s *MockStore) generateData() {
	now := time.Now()

	// 生成 50 台设备
	deviceTemplates := []struct {
		name    string
		region  string
		address string
	}{
		// 华东
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
		// 华南
		{"广州天河-体育西路药房-01", "华南", "广州市天河区体育西路100号"},
		{"广州天河-体育西路药房-02", "华南", "广州市天河区体育西路100号"},
		{"深圳南山-科技园药房-01", "华南", "深圳市南山区科技园南路50号"},
		{"深圳南山-科技园药房-02", "华南", "深圳市南山区科技园南路50号"},
		{"深圳福田-华强北路药房-01", "华南", "深圳市福田区华强北路200号"},
		{"珠海香洲-凤凰路药房-01", "华南", "珠海市香洲区凤凰路100号"},
		{"东莞南城-鸿福路药房-01", "华南", "东莞市南城区鸿福路50号"},
		{"佛山禅城-汾江路药房-01", "华南", "佛山市禅城区汾江路100号"},
		// 华北
		{"北京朝阳-望京药房-01", "华北", "北京市朝阳区望京西路100号"},
		{"北京朝阳-望京药房-02", "华北", "北京市朝阳区望京西路100号"},
		{"北京海淀-中关村药房-01", "华北", "北京市海淀区中关村大街50号"},
		{"北京海淀-中关村药房-02", "华北", "北京市海淀区中关村大街50号"},
		{"天津河西-友谊路药房-01", "华北", "天津市河西区友谊路100号"},
		{"石家庄长安-中山东路药房-01", "华北", "石家庄市长安区中山东路200号"},
		{"济南历下-泉城路药房-01", "华北", "济南市历下区泉城路100号"},
		{"青岛崂山-海尔路药房-01", "华北", "青岛市崂山区海尔路50号"},
		// 西南
		{"成都锦江-春熙路药房-01", "西南", "成都市锦江区春熙路100号"},
		{"成都锦江-春熙路药房-02", "西南", "成都市锦江区春熙路100号"},
		{"重庆渝中-解放碑药房-01", "西南", "重庆市渝中区解放碑步行街50号"},
		{"重庆渝中-解放碑药房-02", "西南", "重庆市渝中区解放碑步行街50号"},
		{"昆明五华-东风西路药房-01", "西南", "昆明市五华区东风西路100号"},
		{"贵阳云岩-中华北路药房-01", "西南", "贵阳市云岩区中华北路50号"},
		// 华中
		{"武汉武昌-光谷药房-01", "华中", "武汉市武昌区光谷大道100号"},
		{"武汉武昌-光谷药房-02", "华中", "武汉市武昌区光谷大道100号"},
		{"武汉洪山-街道口药房-01", "华中", "武汉市洪山区街道口50号"},
		{"长沙岳麓-麓山路药房-01", "华中", "长沙市岳麓区麓山路100号"},
		{"长沙岳麓-麓山路药房-02", "华中", "长沙市岳麓区麓山路100号"},
		{"郑州金水-花园路药房-01", "华中", "郑州市金水区花园路200号"},
		{"南昌红谷滩-红谷中大道药房-01", "华中", "南昌市红谷滩区红谷中大道50号"},
	}

	statuses := []string{device.StatusOnline, device.StatusOffline, device.StatusError, device.StatusMaintenance}
	types := device.DeviceTypes

	for i, tmpl := range deviceTemplates {
		status := statuses[rand.Intn(len(statuses))]
		if i < 35 {
			status = device.StatusOnline // 大部分在线
		}

		s.devices = append(s.devices, device.Device{
			ID:   fmt.Sprintf("DEV-%03d", i+1),
			Name: tmpl.name,
			Type: types[rand.Intn(len(types))],
			Region: tmpl.region,
			Address: tmpl.address,
			Status: status,
			LastHeartbeat: now.Add(-time.Duration(rand.Intn(3600)) * time.Second),
			Firmware: fmt.Sprintf("2.%d.%d", rand.Intn(3), rand.Intn(10)),
			Config: device.DeviceConfig{
				TransactionTimeout: 30,
				ScreenBrightness:   60 + rand.Intn(40),
				VolumeLevel:        50 + rand.Intn(50),
				AutoRebootEnabled:  rand.Intn(2) == 1,
				AutoRebootTime:     fmt.Sprintf("%02d:00", rand.Intn(6)),
				MedicineCategory:   []string{"处方药", "OTC", "保健品"},
			},
			InstalledAt: now.AddDate(-1, -rand.Intn(12), 0),
			Stats: device.DeviceStats{
				TotalTransactions: 1000 + rand.Intn(9000),
				TodayTransactions: rand.Intn(200),
				Uptime:            100 + rand.Intn(8000),
				FaultCount:        rand.Intn(20),
			},
		})
	}

	// 生成故障日志
	for i := 0; i < 100; i++ {
		dev := s.devices[rand.Intn(len(s.devices))]
		severities := []string{device.SeverityLow, device.SeverityMedium, device.SeverityHigh, device.SeverityCritical}
		logTypes := []string{device.LogHardware, device.LogSoftware, device.LogNetwork, device.LogMedicineStock}

		messages := map[string][]string{
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

		resolved := rand.Intn(3) > 0
		var resolvedAt *time.Time
		if resolved {
			t := now.Add(-time.Duration(rand.Intn(24)) * time.Hour)
			resolvedAt = &t
		}

		s.faultLogs = append(s.faultLogs, device.FaultLog{
			ID:         fmt.Sprintf("LOG-%03d", i+1),
			DeviceID:   dev.ID,
			Timestamp:  now.Add(-time.Duration(rand.Intn(720)) * time.Hour),
			Type:       logType,
			Severity:   severities[rand.Intn(len(severities))],
			Message:    msg,
			Resolved:   resolved,
			ResolvedAt: resolvedAt,
		})
	}
}

func (s *MockStore) ListDevices(f device.DeviceFilters) ([]device.Device, error) {
	var result []device.Device
	for _, d := range s.devices {
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
	return result, nil
}

func (s *MockStore) GetDevice(id string) (*device.Device, error) {
	for _, d := range s.devices {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("设备不存在: %s", id)
}

func (s *MockStore) GetDeviceStats(id string, days int) (*device.DeviceStats, error) {
	dev, err := s.GetDevice(id)
	if err != nil {
		return nil, err
	}
	return &dev.Stats, nil
}

func (s *MockStore) GetFaultLogs(f device.LogFilters) ([]device.FaultLog, error) {
	var result []device.FaultLog
	limit := f.Limit
	if limit <= 0 {
		limit = 20
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

func (s *MockStore) UpdateDeviceConfig(id string, config device.DeviceConfig) error {
	for i, d := range s.devices {
		if d.ID == id {
			s.devices[i].Config = config
			return nil
		}
	}
	return fmt.Errorf("设备不存在: %s", id)
}

func (s *MockStore) RebootDevice(id string, force bool) error {
	_, err := s.GetDevice(id)
	if err != nil {
		return err
	}
	// 模拟重启
	return nil
}

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
