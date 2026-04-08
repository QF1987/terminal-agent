// ============================================================
// schema.go - 数据库 Schema 管理
// ============================================================
// 提供建表（DDL）和数据初始化（Seed）功能。
// Go 语言基础概念速查（续）：
//   - const：常量，编译时确定值，不能修改
//   - 切片（slice）：类似动态数组，用 []T 表示，可以 append 扩容
//   - map：键值对集合，类似其他语言的 dict/hashmap/字典
//   - range：用于遍历切片、map、字符串等
//   - tx（事务）：要么全部成功，要么全部回滚，保证数据一致性
//   - Prepare：预编译 SQL 语句，提高批量插入性能
// ============================================================

package store

import (
	cryptorand "crypto/rand" // 密码学安全的随机数生成
	"database/sql"            // 数据库接口
	"encoding/hex"            // 十六进制编码
	"encoding/json"           // JSON 序列化
	"fmt"                     // 格式化输出
	"math/rand"               // 随机数生成（非安全场景）
	"time"                    // 时间处理

	"github.com/QF1987/terminal-agent-go/internal/device" // 设备数据模型
)

// schemaSQL：建表 DDL（数据定义语言），用 Go 的多行字符串（反引号 ``）定义
// CREATE TABLE IF NOT EXISTS：如果表不存在才创建，幂等操作，可以安全重复执行
// CREATE INDEX IF NOT EXISTS：同理，创建索引加速查询
const schemaSQL = `
-- devices 表：存储自助购药机设备的基本信息
-- 每台设备一行记录，用 id 作为主键
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(50) PRIMARY KEY,        -- 设备唯一标识，格式：DEV-001
    name VARCHAR(200) NOT NULL,         -- 设备名称（含位置信息），如"上海浦东-张江药房-01"
    type VARCHAR(100) NOT NULL,         -- 设备类型，如"智能售药机A型"
    region VARCHAR(50) NOT NULL,        -- 所属区域：华东/华南/华北/西南/华中
    address VARCHAR(300) NOT NULL,      -- 详细地址
    status VARCHAR(50) NOT NULL,        -- 当前状态：online/offline/error/maintenance
    last_heartbeat TIMESTAMP NOT NULL,  -- 最后心跳时间（设备定期上报），用于判断设备是否在线
    firmware VARCHAR(50) NOT NULL,      -- 固件版本号，如"2.1.5"
    config JSONB NOT NULL,              -- 设备配置（JSON 格式），包含超时、亮度、音量等设置
    installed_at TIMESTAMP NOT NULL,    -- 安装部署日期
    stats JSONB NOT NULL,               -- 运行统计（JSON 格式），包含交易量、在线时长、故障次数
    token VARCHAR(64),                  -- 设备令牌（第一阶段认证），注册时服务端生成
    device_secret VARCHAR(128),         -- 设备密钥（预留给第二阶段 HMAC-SHA256 签名认证）
    capabilities JSONB                  -- 设备能力声明（对应 proto DeviceCapability）
);

-- fault_logs 表：存储设备的故障和事件日志
-- 一台设备可以有多条故障记录（一对多关系）
CREATE TABLE IF NOT EXISTS fault_logs (
    id VARCHAR(50) PRIMARY KEY,           -- 日志唯一标识，格式：LOG-001
    device_id VARCHAR(50) NOT NULL,        -- 关联的设备 ID（外键，指向 devices.id）
    timestamp TIMESTAMP NOT NULL,          -- 故障发生时间
    type VARCHAR(50) NOT NULL,             -- 故障类型：hardware/software/network/medicine_stock
    severity VARCHAR(50) NOT NULL,         -- 严重程度：low/medium/high/critical
    message TEXT NOT NULL,                 -- 故障描述信息
    resolved BOOLEAN NOT NULL,             -- 是否已解决
    resolved_at TIMESTAMP                  -- 解决时间，允许为 NULL（未解决时）
);

-- commands 表：指令队列，存储服务端下发给设备的指令
-- 对应 proto 的 Command + CommandResult
CREATE TABLE IF NOT EXISTS commands (
    id VARCHAR(50) PRIMARY KEY,              -- 内部编号 CMD-001
    command_id UUID NOT NULL UNIQUE,          -- proto 用的 UUID，设备回报带回来
    device_id VARCHAR(50) NOT NULL,           -- 目标设备
    command_type VARCHAR(50) NOT NULL,        -- reboot / update_config / upgrade_firmware / custom
    payload_json TEXT NOT NULL DEFAULT '{}',   -- 指令参数（JSON 字符串）
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/sent/executing/completed/failed/timeout/rejected
    timeout_seconds INT NOT NULL DEFAULT 30,  -- 执行超时（秒）
    issued_at TIMESTAMP NOT NULL,             -- 服务端下发时间
    sent_at TIMESTAMP,                        -- 实际推送给设备的时间
    executed_at TIMESTAMP,                    -- 设备回报执行完成时间
    result_message TEXT,                       -- 设备回报的执行结果描述
    created_by VARCHAR(100),                  -- 谁发起的（操作员/API）
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 索引：加速常见查询
-- 没有索引时，查询需要扫描全表（慢）
-- 有了索引后，数据库可以直接定位到目标行（快）
CREATE INDEX IF NOT EXISTS idx_devices_region ON devices(region);       -- 按区域查设备
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);       -- 按状态查设备
CREATE INDEX IF NOT EXISTS idx_fault_logs_device_id ON fault_logs(device_id);  -- 按设备查日志
CREATE INDEX IF NOT EXISTS idx_fault_logs_timestamp ON fault_logs(timestamp);  -- 按时间查日志
CREATE INDEX IF NOT EXISTS idx_commands_device_id ON commands(device_id);      -- 按设备查指令
CREATE INDEX IF NOT EXISTS idx_commands_status ON commands(status);            -- 按状态查指令
CREATE INDEX IF NOT EXISTS idx_commands_device_status ON commands(device_id, status);  -- 组合查询
`

// InitSchema：执行建表 SQL，在数据库中创建所有需要的表和索引
// 参数 db：数据库连接
// 返回值：成功返回 nil，失败返回错误
// 通常在应用启动时调用一次，确保表结构存在
func InitSchema(db *sql.DB) error {
	// Exec：执行不需要返回结果的 SQL（DDL/DML 都可以）
	// schemaSQL 是一个 const 字符串，包含多条 SQL 语句
	_, err := db.Exec(schemaSQL)
	return err  // err == nil 表示成功
}

// SeedData：将模拟数据写入数据库（仅当表为空时才插入）
// 参数 db：数据库连接
// 设计意图：开发/演示时自动生成测试数据，生产环境已有数据时自动跳过
func SeedData(db *sql.DB) error {
	// 第一步：检查表中是否已有数据
	// SELECT COUNT(*) 返回满足条件的行数，即使没有数据也会返回 0
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count)
	if err != nil {
		return fmt.Errorf("检查数据失败: %w", err)
	}
	if count > 0 {
		return nil // 已有数据，不需要重复插入，直接返回 nil（成功）
	}

	now := time.Now()  // 获取当前时间，用于生成相对时间

	// 设备模板数据：每台设备的基础信息（名称、区域、地址）
	// 使用结构体切片存储，便于遍历
	// 分为 5 个区域，模拟真实的分布式部署场景
	deviceTemplates := []struct {
		name    string  // 设备名称（含位置）
		region  string  // 所属区域
		address string  // 详细地址
	}{
		// ─── 华东（10 台）─────────────────────
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
		// ─── 华南（8 台）─────────────────────
		{"广州天河-体育西路药房-01", "华南", "广州市天河区体育西路100号"},
		{"广州天河-体育西路药房-02", "华南", "广州市天河区体育西路100号"},
		{"深圳南山-科技园药房-01", "华南", "深圳市南山区科技园南路50号"},
		{"深圳南山-科技园药房-02", "华南", "深圳市南山区科技园南路50号"},
		{"深圳福田-华强北路药房-01", "华南", "深圳市福田区华强北路200号"},
		{"珠海香洲-凤凰路药房-01", "华南", "珠海市香洲区凤凰路100号"},
		{"东莞南城-鸿福路药房-01", "华南", "东莞市南城区鸿福路50号"},
		{"佛山禅城-汾江路药房-01", "华南", "佛山市禅城区汾江路100号"},
		// ─── 华北（8 台）─────────────────────
		{"北京朝阳-望京药房-01", "华北", "北京市朝阳区望京西路100号"},
		{"北京朝阳-望京药房-02", "华北", "北京市朝阳区望京西路100号"},
		{"北京海淀-中关村药房-01", "华北", "北京市海淀区中关村大街50号"},
		{"北京海淀-中关村药房-02", "华北", "北京市海淀区中关村大街50号"},
		{"天津河西-友谊路药房-01", "华北", "天津市河西区友谊路100号"},
		{"石家庄长安-中山东路药房-01", "华北", "石家庄市长安区中山东路200号"},
		{"济南历下-泉城路药房-01", "华北", "济南市历下区泉城路100号"},
		{"青岛崂山-海尔路药房-01", "华北", "青岛市崂山区海尔路50号"},
		// ─── 西南（6 台）─────────────────────
		{"成都锦江-春熙路药房-01", "西南", "成都市锦江区春熙路100号"},
		{"成都锦江-春熙路药房-02", "西南", "成都市锦江区春熙路100号"},
		{"重庆渝中-解放碑药房-01", "西南", "重庆市渝中区解放碑步行街50号"},
		{"重庆渝中-解放碑药房-02", "西南", "重庆市渝中区解放碑步行街50号"},
		{"昆明五华-东风西路药房-01", "西南", "昆明市五华区东风西路100号"},
		{"贵阳云岩-中华北路药房-01", "西南", "贵阳市云岩区中华北路50号"},
		// ─── 华中（7 台）─────────────────────
		{"武汉武昌-光谷药房-01", "华中", "武汉市武昌区光谷大道100号"},
		{"武汉武昌-光谷药房-02", "华中", "武汉市武昌区光谷大道100号"},
		{"武汉洪山-街道口药房-01", "华中", "武汉市洪山区街道口50号"},
		{"长沙岳麓-麓山路药房-01", "华中", "长沙市岳麓区麓山路100号"},
		{"长沙岳麓-麓山路药房-02", "华中", "长沙市岳麓区麓山路100号"},
		{"郑州金水-花园路药房-01", "华中", "郑州市金水区花园路200号"},
		{"南昌红谷滩-红谷中大道药房-01", "华中", "南昌市红谷滩区红谷中大道50号"},
	}

	// 可选值列表，用于随机生成数据
	statuses := []string{device.StatusOnline, device.StatusOffline, device.StatusError, device.StatusMaintenance}
	types := device.DeviceTypes  // 设备类型列表（从 device 包导入）

	// ─── 使用事务批量插入设备数据 ──────────────────────
	// 事务（Transaction）：保证"要么全部成功，要么全部回滚"
	// 比如插入到第 30 台设备时失败了，前面 29 台也会全部回滚，不会留下脏数据
	tx, err := db.Begin()  // 开始事务
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()  // defer：如果函数提前 return（比如出错），自动回滚事务
	// 注意：如果最后执行了 tx.Commit()，Rollback 就不会执行（因为事务已提交）

	// Prepare：预编译 SQL 语句
	// 批量插入时，Prepare 比每次拼 SQL 快很多（数据库只需解析一次）
	// $1-$11 是参数占位符，执行时传入实际值
	stmt, err := tx.Prepare(`
		INSERT INTO devices (id, name, type, region, address, status, last_heartbeat, firmware, config, installed_at, stats, token, device_secret, capabilities)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`)
	if err != nil {
		return fmt.Errorf("预处理设备语句失败: %w", err)
	}
	defer stmt.Close()  // 函数结束时关闭预编译语句，释放数据库资源

	// 遍历设备模板，逐个生成并插入设备数据
	for i, tmpl := range deviceTemplates {
		// 随机选择状态
		status := statuses[rand.Intn(len(statuses))]  // rand.Intn(n) 返回 [0, n) 的随机整数
		if i < 35 {
			status = device.StatusOnline  // 前 35 台设备强制设为在线，模拟大部分设备正常运行
		}

		// 构建设备配置（随机生成参数）
		config := device.DeviceConfig{
			TransactionTimeout: 30,                        // 交易超时时间（秒）
			ScreenBrightness:   60 + rand.Intn(40),        // 屏幕亮度 60-99
			VolumeLevel:        50 + rand.Intn(50),        // 音量 50-99
			AutoRebootEnabled:  rand.Intn(2) == 1,         // 是否开启自动重启（50% 概率）
			AutoRebootTime:     fmt.Sprintf("%02d:00", rand.Intn(6)),  // 自动重启时间，如 "03:00"
			MedicineCategory:   []string{"处方药", "OTC", "保健品"},   // 售卖的药品类别
		}
		configJSON, _ := json.Marshal(config)  // 序列化成 JSON，_ 忽略错误（config 结构简单不会失败）

		// 构建设备统计（随机生成）
		stats := device.DeviceStats{
			TotalTransactions: 1000 + rand.Intn(9000),  // 总交易量 1000-9999
			TodayTransactions: rand.Intn(200),           // 今日交易量 0-199
			Uptime:            100 + rand.Intn(8000),    // 在线时长（小时）100-8099
			FaultCount:        rand.Intn(20),            // 故障次数 0-19
		}
		statsJSON, _ := json.Marshal(stats)

		// 执行预编译的 INSERT 语句，传入 14 个参数
		_, err = stmt.Exec(
			fmt.Sprintf("DEV-%03d", i+1),  // 设备 ID：DEV-001, DEV-002, ...
			tmpl.name,                       // 从模板读取名称
			types[rand.Intn(len(types))],    // 随机设备类型
			tmpl.region,                     // 从模板读取区域
			tmpl.address,                    // 从模板读取地址
			status,                          // 随机/固定状态
			// 心跳时间：过去 0-3600 秒内随机（模拟设备最后活跃时间）
			now.Add(-time.Duration(rand.Intn(3600))*time.Second),
			// 固件版本：2.x.y 格式
			fmt.Sprintf("2.%d.%d", rand.Intn(3), rand.Intn(10)),
			configJSON,  // JSON 格式的配置
			// 安装时间：过去 1-12 个月内随机
			now.AddDate(-1, -rand.Intn(12), 0),
			statsJSON,   // JSON 格式的统计
			generateToken(),  // 设备令牌
			nil,              // device_secret（第一阶段不填）
			generateCapabilities(fmt.Sprintf("2.%d.%d", rand.Intn(3), rand.Intn(10))),  // 设备能力声明
		)
		if err != nil {
			return fmt.Errorf("插入设备失败: %w", err)
		}
	}

	// ─── 插入故障日志 ──────────────────────────────
	// 同样使用 Prepare + 批量插入的方式
	logStmt, err := tx.Prepare(`
		INSERT INTO fault_logs (id, device_id, timestamp, type, severity, message, resolved, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("预处理日志语句失败: %w", err)
	}
	defer logStmt.Close()

	// 按故障类型组织的模拟消息
	// map[key_type][]value_type：Go 的字典/哈希表
	// 键是故障类型，值是该类型下的消息列表
	messages := map[string][]string{
		device.LogHardware: {       // 硬件故障
			"打印机卡纸，已自动恢复",
			"扫码器响应超时",
			"触摸屏校准偏移",
			"出药机构卡顿",
		},
		device.LogSoftware: {       // 软件故障
			"应用崩溃，已自动重启",
			"数据库连接超时",
			"内存使用率过高(85%)",
			"系统更新失败",
		},
		device.LogNetwork: {        // 网络故障
			"网络模块故障",
			"4G信号弱，切换至备用网络",
			"VPN连接中断",
			"心跳包丢失超过阈值",
		},
		device.LogMedicineStock: {  // 药品库存告警
			"布洛芬库存低于警戒线",
			"板蓝根即将过期",
			"感冒灵库存为0",
			"药品补货延迟",
		},
	}

	severities := []string{device.SeverityLow, device.SeverityMedium, device.SeverityHigh, device.SeverityCritical}
	logTypes := []string{device.LogHardware, device.LogSoftware, device.LogNetwork, device.LogMedicineStock}

	// 生成 100 条随机故障日志
	for i := 0; i < 100; i++ {
		// 随机选择一个设备
		devID := fmt.Sprintf("DEV-%03d", rand.Intn(len(deviceTemplates))+1)
		// 随机选择故障类型
		logType := logTypes[rand.Intn(len(logTypes))]
		// 随机选择严重程度
		severity := severities[rand.Intn(len(severities))]
		// 从对应类型的模板中随机选一条消息
		msg := messages[logType][rand.Intn(len(messages[logType]))]

		// 2/3 的概率为已解决
		resolved := rand.Intn(3) > 0
		// interface{}：Go 的"任意类型"，类似 TypeScript 的 any
		// 用于 Exec 参数时可以传入任何值（包括 nil）
		var resolvedAt interface{}
		if resolved {
			t := now.Add(-time.Duration(rand.Intn(24)) * time.Hour)  // 解决时间为过去 0-24 小时内
			resolvedAt = t
		}
		// 如果未解决，resolvedAt 保持为 nil，数据库中会存为 NULL

		_, err = logStmt.Exec(
			fmt.Sprintf("LOG-%03d", i+1),  // 日志 ID：LOG-001, LOG-002, ...
			devID,                            // 关联的设备 ID
			// 故障时间：过去 0-720 小时内（30 天）
			now.Add(-time.Duration(rand.Intn(720))*time.Hour),
			logType,      // 故障类型
			severity,     // 严重程度
			msg,          // 故障描述
			resolved,     // 是否已解决
			resolvedAt,   // 解决时间（未解决时为 nil → 数据库存 NULL）
		)
		if err != nil {
			return fmt.Errorf("插入日志失败: %w", err)
		}
	}

	// Commit：提交事务，所有 INSERT 操作真正生效
	// 如果这里之前有 stmt.Exec 失败，函数会提前 return，defer tx.Rollback() 会自动回滚
	return tx.Commit()
}

// ─── 辅助函数 ─────────────────────────────────────────────

// generateToken：生成 64 字符的随机设备令牌（密码学安全）
func generateToken() string {
	b := make([]byte, 32)        // 32 字节 = 256 位熵
	cryptorand.Read(b)           // crypto/rand.Read 填充随机字节
	return hex.EncodeToString(b) // 转成 64 字符的十六进制字符串
}

// generateCapabilities：根据固件版本生成设备能力声明
func generateCapabilities(firmwareVersion string) []byte {
	allFeatures := []string{
		"heartbeat_basic",
		"heartbeat_metrics",
		"event_report",
		"status_report",
		"config_push",
		"command_reboot",
		"command_firmware",
	}
	// 随机选 3-6 个功能
	n := 3 + rand.Intn(4)
	features := make([]string, 0, n)
	used := make(map[int]bool)
	for len(features) < n {
		idx := rand.Intn(len(allFeatures))
		if !used[idx] {
			used[idx] = true
			features = append(features, allFeatures[idx])
		}
	}

	cap := device.DeviceCapability{
		FirmwareVersion:   firmwareVersion,
		ProtoVersion:      1,
		SupportedFeatures: features,
	}
	b, _ := json.Marshal(cap)
	return b
}
