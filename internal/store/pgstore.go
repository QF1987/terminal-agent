// ============================================================
// pgstore.go - PostgreSQL 持久化存储
// ============================================================
// 实现了 Store 接口，使用 PostgreSQL 数据库存储设备数据和故障日志。
// Go 语言基础概念速查：
//   - struct：类似其他语言的"类"，用字段存储数据
//   - method (func (s *PGStore) xxx)：绑定到 struct 上的方法，类似类的成员函数
//   - interface：定义一组方法签名，实现这些方法就自动满足接口（无需显式声明 implements）
//   - error：Go 没有异常机制，函数通过返回 error 来表示出错，nil 表示成功
//   - *：指针，指向内存地址。*sql.DB 表示"指向 sql.DB 的指针"
//   - defer：延迟执行，在函数返回前自动调用（常用于释放资源、关闭连接）
//   - _：空白标识符，表示"忽略这个返回值"
// ============================================================

package store

import (
	"database/sql"    // Go 标准库的数据库接口，支持各种数据库驱动
	"encoding/json"   // JSON 序列化/反序列化
	"fmt"             // 格式化输出，类似 C 语言的 printf
	"strings"         // 字符串操作工具
	"time"            // 时间处理

	"github.com/QF1987/terminal-agent-go/internal/device"  // 项目内部的设备数据模型
	_ "github.com/lib/pq"  // PostgreSQL 驱动，导入时用 _ 表示"只注册驱动，不直接调用"
)

// PGStore：PostgreSQL 存储实现，实现了 Store 接口
// Go 的接口满足是隐式的：只要 PGStore 实现了 Store 接口定义的所有方法，就自动满足
// 结构体字段：db 是数据库连接池（*sql.DB 是并发安全的，可以全局复用）
type PGStore struct {
	db *sql.DB  // 数据库连接池，由外部传入，不需要自己创建
}

// NewPGStore：创建 PostgreSQL 存储实例的工厂函数
// Go 没有构造函数，惯例是用 NewXxx 函数来创建实例
// 参数 db：已建立的数据库连接，通常通过 sql.Open() + db.Ping() 创建
// 返回值：*PGStore 指针（Go 惯例：返回指针而不是值，避免大结构体的复制开销）
func NewPGStore(db *sql.DB) *PGStore {
	return &PGStore{db: db}  // & 取地址，创建指向新 PGStore 的指针
}

// ─── 辅助方法（内部使用，不暴露给外部） ─────────────────────────

// scanDevice：从数据库查询结果中读取一行数据，组装成 Device 结构体
// 参数 row：*sql.Row 表示查询返回的单行结果（QueryRow 返回的）
// 返回值：成功返回 Device 指针，失败返回 error
// 注意：sql.Row.Scan() 会把数据库列的值按顺序填入传入的变量地址
//       如果列的值为 NULL 或类型不匹配，Scan 会返回 error
func (s *PGStore) scanDevice(row *sql.Row) (*device.Device, error) {
	var d device.Device    // 声明一个零值 Device（Go 声明变量会自动初始化为零值）
	var configJSON string  // 数据库中 config 字段是 JSONB 类型，读出来是字符串
	var statsJSON string   // 同上，stats 也是 JSONB
	var lastHeartbeat, installedAt time.Time  // 数据库 TIMESTAMP 类型，读出来是 time.Time
	var token, deviceSecret sql.NullString    // 可为 NULL 的字符串
	var capabilitiesJSON sql.NullString       // 可为 NULL 的 JSONB

	// Scan 按 SELECT 列的顺序，逐个填入变量地址（& 取地址）
	// 必须保证 SELECT 的列顺序和 Scan 的参数顺序完全一致
	err := row.Scan(
		&d.ID, &d.Name, &d.Type, &d.Region, &d.Address,   // 基本字段直接映射
		&d.Status, &lastHeartbeat, &d.Firmware,
		&configJSON, &installedAt, &statsJSON,
		&token, &deviceSecret, &capabilitiesJSON,  // 新增字段
	)
	if err != nil {
		return nil, err  // 查询失败（比如设备不存在时会返回 sql.ErrNoRows）
	}

	d.LastHeartbeat = lastHeartbeat  // 把临时变量赋值给结构体字段
	d.InstalledAt = installedAt

	// 处理可为 NULL 的字段
	if token.Valid {
		d.Token = token.String
	}
	if deviceSecret.Valid {
		d.DeviceSecret = deviceSecret.String
	}
	if capabilitiesJSON.Valid {
		json.Unmarshal([]byte(capabilitiesJSON.String), &d.Capabilities)
	}

	// json.Unmarshal：把 JSON 字符串解析成 Go 结构体
	// []byte(configJSON)：把 string 转成 []byte（字节切片），这是 Unmarshal 要求的参数类型
	// &d.Config：传入字段的地址，Unmarshal 会直接修改这个字段的值
	if err := json.Unmarshal([]byte(configJSON), &d.Config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)  // %w 包装原始错误，保留错误链
	}
	if err := json.Unmarshal([]byte(statsJSON), &d.Stats); err != nil {
		return nil, fmt.Errorf("解析统计失败: %w", err)
	}

	return &d, nil  // 返回设备指针，nil 表示没有错误
}

// buildDeviceFilters：根据过滤条件动态构建 SQL WHERE 子句
// 参数 f：DeviceFilters 结构体，包含 Region/Status/Type/Keyword 等可选过滤字段
// 返回值：
//   - where：WHERE 条件字符串切片，比如 ["region = $1", "status = $2"]
//   - args：对应的参数值切片，比如 ["华东", "在线"]
// 注意：这里有个 bug —— $ 后面没有跟参数编号（$1, $2...），
//       在 PostgreSQL 中 $ 后面必须有数字，但这里依赖调用方的处理
func (s *PGStore) buildDeviceFilters(f device.DeviceFilters) (where []string, args []interface{}) {
	argIdx := 1
	if f.Region != "" {
		where = append(where, fmt.Sprintf("region = $%d", argIdx))
		args = append(args, f.Region)
		argIdx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, f.Type)
		argIdx++
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR address ILIKE $%d OR id ILIKE $%d)", argIdx, argIdx+1, argIdx+2))
		args = append(args, kw, kw, kw)
		argIdx += 3
	}
	return
}

// ─── Store 接口实现 ─────────────────────────────────────────
// 以下方法必须全部实现才能满足 Store 接口

// ListDevices：列出所有设备，支持按区域/状态/类型/关键字过滤
// 参数 f：过滤条件，字段为空则不过滤
// 返回值：设备列表和可能的错误
func (s *PGStore) ListDevices(f device.DeviceFilters) ([]device.Device, error) {
	where, args := s.buildDeviceFilters(f)  // 构建 WHERE 条件

	// 动态拼接 SQL 查询语句
	// 注意：生产环境应该用参数化查询防止 SQL 注入，这里的 keyword 部分已经用了 $ 参数
	query := "SELECT id, name, type, region, address, status, last_heartbeat, firmware, config, installed_at, stats, token, device_secret, capabilities FROM devices"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")  // 用 AND 连接所有条件
	}
	query += " ORDER BY id"  // 按设备 ID 排序

	// Query 执行查询，返回多行结果
	// args... 是 Go 的可变参数展开语法，把切片展开成独立的参数传给函数
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()  // defer：函数返回前自动关闭结果集，防止数据库连接泄漏

	var devices []device.Device  // 声明切片（类似动态数组），初始值为 nil

	// rows.Next() 指向下一行，返回 false 表示没有更多行了
	for rows.Next() {
		var d device.Device
		var configJSON, statsJSON string
		var lastHeartbeat, installedAt time.Time
		var token, deviceSecret sql.NullString
		var capabilitiesJSON sql.NullString

		// 每一行都 Scan 一次，把列值读入变量
		err := rows.Scan(
			&d.ID, &d.Name, &d.Type, &d.Region, &d.Address,
			&d.Status, &lastHeartbeat, &d.Firmware,
			&configJSON, &installedAt, &statsJSON,
			&token, &deviceSecret, &capabilitiesJSON,
		)
		if err != nil {
			return nil, err  // Scan 失败，直接返回错误
		}

		d.LastHeartbeat = lastHeartbeat
		d.InstalledAt = installedAt
		if token.Valid {
			d.Token = token.String
		}
		if deviceSecret.Valid {
			d.DeviceSecret = deviceSecret.String
		}
		if capabilitiesJSON.Valid {
			json.Unmarshal([]byte(capabilitiesJSON.String), &d.Capabilities)
		}

		// JSON 反序列化（同 scanDevice 中的逻辑）
		if err := json.Unmarshal([]byte(configJSON), &d.Config); err != nil {
			return nil, fmt.Errorf("解析配置失败: %w", err)
		}
		if err := json.Unmarshal([]byte(statsJSON), &d.Stats); err != nil {
			return nil, fmt.Errorf("解析统计失败: %w", err)
		}

		devices = append(devices, d)  // 把解析好的设备追加到切片中
	}
	// rows.Err()：检查遍历过程中是否有错误（比如网络中断）
	return devices, rows.Err()
}

// GetDevice：根据设备 ID 获取单个设备详情
// 参数 id：设备 ID，比如 "DEV-001"
// 返回值：找到返回 Device 指针，未找到返回 sql.ErrNoRows 错误
func (s *PGStore) GetDevice(id string) (*device.Device, error) {
	// QueryRow：查询单行，返回 *sql.Row
	// $1 是 PostgreSQL 的参数占位符，对应第二个参数 id
	// 用 $1 而不是字符串拼接，是为了防止 SQL 注入攻击
	row := s.db.QueryRow(
		"SELECT id, name, type, region, address, status, last_heartbeat, firmware, config, installed_at, stats, token, device_secret, capabilities FROM devices WHERE id = $1",
		id,
	)
	return s.scanDevice(row)  // 复用 scanDevice 辅助方法解析结果
}

// GetDeviceStats：获取设备的统计信息
// 参数 id：设备 ID
// 参数 days：天数（当前实现中未使用，预留用于查询历史统计）
// 注意：当前实现只是从设备记录中读取 stats 字段，没有按天聚合
func (s *PGStore) GetDeviceStats(id string, days int) (*device.DeviceStats, error) {
	dev, err := s.GetDevice(id)  // 先获取设备
	if err != nil {
		return nil, err
	}
	return &dev.Stats, nil  // 返回设备中存储的统计信息
}

// GetFaultLogs：获取故障日志，支持按设备ID/严重程度/类型/时间范围过滤
// 参数 f：日志过滤条件
// 返回值：故障日志列表和可能的错误
func (s *PGStore) GetFaultLogs(f device.LogFilters) ([]device.FaultLog, error) {
	where := []string{}           // WHERE 条件列表
	args := []interface{}{}       // 参数值列表
	argIdx := 1                   // PostgreSQL 参数编号从 $1 开始

	// 动态构建过滤条件，每个条件对应一个 $N 参数
	if f.DeviceID != "" {
		where = append(where, fmt.Sprintf("device_id = $%d", argIdx))  // fmt.Sprintf 格式化字符串
		args = append(args, f.DeviceID)
		argIdx++  // 自增参数编号
	}
	if f.Severity != "" {
		where = append(where, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, f.Severity)
		argIdx++
	}
	if f.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, f.Type)
		argIdx++
	}
	if f.Days > 0 {
		// PostgreSQL 的 INTERVAL 语法：NOW() - INTERVAL '30 days' 表示"30天前"
		// 注意：这里用 fmt.Sprintf 直接拼接了整数到 SQL 中，不是参数化查询
		// 但由于 f.Days 是 int 类型，不存在 SQL 注入风险
		where = append(where, fmt.Sprintf("timestamp >= NOW() - INTERVAL '%d days'", f.Days))
	}

	// 设置默认返回条数限制
	limit := f.Limit
	if limit <= 0 {
		limit = 20  // 默认返回最近 20 条
	}

	// 拼接完整 SQL
	query := "SELECT id, device_id, timestamp, type, severity, message, resolved, resolved_at FROM fault_logs"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIdx)  // DESC：降序，最新的在前
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []device.FaultLog

	for rows.Next() {
		var log device.FaultLog
		var resolvedAt sql.NullTime  // sql.NullTime：表示"可能是 NULL 的时间"
		// 数据库中 resolved_at 允许为 NULL（未解决的故障没有解决时间）
		// sql.NullTime 有两个字段：Valid（是否非NULL）和 Time（实际值）

		err := rows.Scan(&log.ID, &log.DeviceID, &log.Timestamp, &log.Type, &log.Severity, &log.Message, &log.Resolved, &resolvedAt)
		if err != nil {
			return nil, err
		}

		// 检查 resolved_at 是否有值（数据库中是否为 NULL）
		if resolvedAt.Valid {
			log.ResolvedAt = &resolvedAt.Time  // 有值：取地址赋给指针字段
		}
		// 如果 resolvedAt.Valid == false（即数据库值为 NULL），
		// log.ResolvedAt 保持为 nil（Go 指针零值就是 nil）

		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// UpdateDeviceConfig：更新设备的配置信息
// 参数 id：设备 ID
// 参数 config：新的配置内容（DeviceConfig 结构体）
// 返回值：成功返回 nil，设备不存在返回错误
func (s *PGStore) UpdateDeviceConfig(id string, config device.DeviceConfig) error {
	// json.Marshal：把 Go 结构体序列化成 JSON 字节切片
	// 与 json.Unmarshal（反序列化）是互逆操作
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// Exec：执行 UPDATE 语句，返回 sql.Result
	// 用 $1, $2 参数化查询，防止 SQL 注入
	result, err := s.db.Exec("UPDATE devices SET config = $1 WHERE id = $2", configJSON, id)
	if err != nil {
		return err
	}

	// RowsAffected()：返回受影响的行数
	// 如果 WHERE id = $2 没有匹配到任何行，说明设备不存在
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("设备不存在: %s", id)  // 返回描述性错误
	}
	return nil  // nil 表示成功（Go 惯例：nil error = 无错误）
}

// RebootDevice：重启设备（当前实现只是更新心跳时间，模拟重启效果）
// 参数 id：设备 ID
// 参数 force：是否强制重启（当前实现中未使用）
// 注意：实际项目中这里应该调用设备的重启 API，目前只是更新 last_heartbeat 字段
func (s *PGStore) RebootDevice(id string, force bool) error {
	result, err := s.db.Exec("UPDATE devices SET last_heartbeat = NOW() WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("设备不存在: %s", id)
	}
	return nil
}

// ─── 指令相关方法 ─────────────────────────────────────────

// CreateCommand：创建新指令并写入 commands 表
// 返回值：成功返回完整 Command（含数据库生成的字段）
func (s *PGStore) CreateCommand(cmd device.Command) error {
	_, err := s.db.Exec(`
		INSERT INTO commands (id, command_id, device_id, command_type, payload_json, status, timeout_seconds, issued_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, cmd.ID, cmd.CommandID, cmd.DeviceID, cmd.CommandType, cmd.PayloadJSON,
		cmd.Status, cmd.TimeoutSeconds, cmd.IssuedAt, cmd.CreatedBy)
	return err
}

// GetCommand：根据内部编号获取指令
func (s *PGStore) GetCommand(id string) (*device.Command, error) {
	row := s.db.QueryRow(
		"SELECT id, command_id, device_id, command_type, payload_json, status, timeout_seconds, issued_at, sent_at, executed_at, result_message, created_by, created_at, updated_at FROM commands WHERE id = $1",
		id,
	)
	return s.scanCommand(row)
}

// GetCommandByUUID：根据 proto UUID 获取指令
func (s *PGStore) GetCommandByUUID(commandID string) (*device.Command, error) {
	row := s.db.QueryRow(
		"SELECT id, command_id, device_id, command_type, payload_json, status, timeout_seconds, issued_at, sent_at, executed_at, result_message, created_by, created_at, updated_at FROM commands WHERE command_id = $1",
		commandID,
	)
	return s.scanCommand(row)
}

// ListCommands：列出指令，支持按设备/状态/类型过滤
func (s *PGStore) ListCommands(f device.CommandFilters) ([]device.Command, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if f.DeviceID != "" {
		where = append(where, fmt.Sprintf("device_id = $%d", argIdx))
		args = append(args, f.DeviceID)
		argIdx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.Type != "" {
		where = append(where, fmt.Sprintf("command_type = $%d", argIdx))
		args = append(args, f.Type)
		argIdx++
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	query := "SELECT id, command_id, device_id, command_type, payload_json, status, timeout_seconds, issued_at, sent_at, executed_at, result_message, created_by, created_at, updated_at FROM commands"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY issued_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []device.Command
	for rows.Next() {
		var cmd device.Command
		var sentAt, executedAt sql.NullTime
		var resultMessage sql.NullString

		err := rows.Scan(
			&cmd.ID, &cmd.CommandID, &cmd.DeviceID, &cmd.CommandType, &cmd.PayloadJSON,
			&cmd.Status, &cmd.TimeoutSeconds, &cmd.IssuedAt,
			&sentAt, &executedAt, &resultMessage,
			&cmd.CreatedBy, &cmd.CreatedAt, &cmd.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if sentAt.Valid {
			cmd.SentAt = &sentAt.Time
		}
		if executedAt.Valid {
			cmd.ExecutedAt = &executedAt.Time
		}
		if resultMessage.Valid {
			cmd.ResultMessage = resultMessage.String
		}
		cmds = append(cmds, cmd)
	}
	return cmds, rows.Err()
}

// GetPendingCommands：获取某设备的待下发指令
func (s *PGStore) GetPendingCommands(deviceID string) ([]device.Command, error) {
	return s.ListCommands(device.CommandFilters{
		DeviceID: deviceID,
		Status:   device.CommandStatusPending,
	})
}

// UpdateCommandStatus：更新指令状态（sent → completed/failed/timeout/rejected）
func (s *PGStore) UpdateCommandStatus(id string, status string, resultMessage string) error {
	var result sql.Result
	var err error

	now := time.Now()
	switch status {
	case device.CommandStatusSent:
		result, err = s.db.Exec(
			"UPDATE commands SET status = $1, sent_at = $2, updated_at = $2 WHERE id = $3",
			status, now, id,
		)
	case device.CommandStatusCompleted, device.CommandStatusFailed, device.CommandStatusTimeout, device.CommandStatusRejected:
		result, err = s.db.Exec(
			"UPDATE commands SET status = $1, executed_at = $2, result_message = $3, updated_at = $2 WHERE id = $4",
			status, now, resultMessage, id,
		)
	default:
		result, err = s.db.Exec(
			"UPDATE commands SET status = $1, updated_at = NOW() WHERE id = $2",
			status, id,
		)
	}
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("指令不存在: %s", id)
	}
	return nil
}

// UpdateCommandResultByUUID：设备通过 proto UUID 回报指令执行结果
func (s *PGStore) UpdateCommandResultByUUID(commandID string, status string, message string) error {
	now := time.Now()
	result, err := s.db.Exec(
		"UPDATE commands SET status = $1, executed_at = $2, result_message = $3, updated_at = $2 WHERE command_id = $4",
		status, now, message, commandID,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("指令不存在: %s", commandID)
	}
	return nil
}

// ExpireTimedOutCommands：将已超时的 sent 指令标记为 timeout
// 建议定期调用（比如每分钟一次）
func (s *PGStore) ExpireTimedOutCommands() (int64, error) {
	result, err := s.db.Exec(`
		UPDATE commands SET status = 'timeout', result_message = '服务端检测超时', updated_at = NOW()
		WHERE status = 'sent' AND sent_at + (timeout_seconds || ' seconds')::interval < NOW()
	`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanCommand：从数据库行扫描出 Command 结构体
func (s *PGStore) scanCommand(row *sql.Row) (*device.Command, error) {
	var cmd device.Command
	var sentAt, executedAt sql.NullTime
	var resultMessage sql.NullString

	err := row.Scan(
		&cmd.ID, &cmd.CommandID, &cmd.DeviceID, &cmd.CommandType, &cmd.PayloadJSON,
		&cmd.Status, &cmd.TimeoutSeconds, &cmd.IssuedAt,
		&sentAt, &executedAt, &resultMessage,
		&cmd.CreatedBy, &cmd.CreatedAt, &cmd.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if sentAt.Valid {
		cmd.SentAt = &sentAt.Time
	}
	if executedAt.Valid {
		cmd.ExecutedAt = &executedAt.Time
	}
	if resultMessage.Valid {
		cmd.ResultMessage = resultMessage.String
	}
	return &cmd, nil
}

// ─── 设备状态/能力更新 ───────────────────────────────────

// UpdateDeviceStatus：更新设备状态和最后心跳时间
func (s *PGStore) UpdateDeviceStatus(id string, status string) error {
	result, err := s.db.Exec(
		"UPDATE devices SET status = $1, last_heartbeat = NOW() WHERE id = $2",
		status, id,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("设备不存在: %s", id)
	}
	return nil
}

// UpdateDeviceCapabilities：更新设备能力声明
func (s *PGStore) UpdateDeviceCapabilities(id string, cap device.DeviceCapability) error {
	capJSON, err := json.Marshal(cap)
	if err != nil {
		return fmt.Errorf("序列化能力声明失败: %w", err)
	}
	result, err := s.db.Exec(
		"UPDATE devices SET capabilities = $1 WHERE id = $2",
		capJSON, id,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("设备不存在: %s", id)
	}
	return nil
}

// CreateFaultLog：创建故障/事件日志
func (s *PGStore) CreateFaultLog(log device.FaultLog) error {
	var resolvedAt interface{}
	if log.ResolvedAt != nil {
		resolvedAt = *log.ResolvedAt
	}
	_, err := s.db.Exec(`
		INSERT INTO fault_logs (id, device_id, timestamp, type, severity, message, resolved, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, log.ID, log.DeviceID, log.Timestamp, log.Type, log.Severity, log.Message, log.Resolved, resolvedAt)
	return err
}
