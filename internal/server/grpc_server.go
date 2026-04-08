// ============================================================
// grpc_server.go - gRPC 服务端实现
// ============================================================
// 实现 DeviceService 和 CommandService 两个 gRPC 服务
// 把 gRPC 请求转换成对 Store 的操作
// ============================================================

package server

import (
	"context"     // 上下文，用于请求生命周期管理
	"fmt"         // 格式化输出
	"log"         // 日志
	"sync"        // 同步原语（互斥锁）
	"time"        // 时间处理

	"google.golang.org/grpc"  // gRPC 框架

	v1 "github.com/QF1987/terminal-agent-go/gen/terminal_agent/v1"  // protoc 生成的代码
	"github.com/QF1987/terminal-agent-go/internal/device"          // 设备数据模型
	"github.com/QF1987/terminal-agent-go/internal/store"            // 数据存储接口
)

// ─── 设备连接管理 ─────────────────────────────────────────

// DeviceStream：记录一个设备的 CommandStream 连接
type DeviceStream struct {
	DeviceID string
	Stream   grpc.ServerStreamingServer[v1.Command] // server streaming 接口
}

// ConnectionManager：管理所有在线设备的 CommandStream 长连接
// 线程安全，用 RWMutex 保护内部 map
type ConnectionManager struct {
	mu      sync.RWMutex               // 读写锁
	streams map[string]*DeviceStream    // device_id → DeviceStream
}

// NewConnectionManager：创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		streams: make(map[string]*DeviceStream),
	}
}

// Register：注册一个设备的 stream 连接
func (cm *ConnectionManager) Register(deviceID string, stream grpc.ServerStreamingServer[v1.Command]) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.streams[deviceID] = &DeviceStream{
		DeviceID: deviceID,
		Stream:   stream,
	}
	log.Printf("[ConnMgr] 设备 %s CommandStream 已连接", deviceID)
}

// Unregister：注销一个设备的 stream 连接
func (cm *ConnectionManager) Unregister(deviceID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.streams, deviceID)
	log.Printf("[ConnMgr] 设备 %s CommandStream 已断开", deviceID)
}

// Send：向指定设备推送指令
// 返回 true 表示推送成功（设备在线且 stream 可用），false 表示设备不在线
func (cm *ConnectionManager) Send(deviceID string, cmd *v1.Command) bool {
	cm.mu.RLock()
	ds, ok := cm.streams[deviceID]
	cm.mu.RUnlock()
	if !ok {
		return false // 设备不在线，没有 stream 连接
	}

	err := ds.Stream.Send(cmd)
	if err != nil {
		log.Printf("[ConnMgr] 向设备 %s 推送指令失败: %v", deviceID, err)
		// 发送失败说明连接已断开，清理掉
		cm.Unregister(deviceID)
		return false
	}
	return true
}

// OnlineDevices：返回当前在线的设备 ID 列表
func (cm *ConnectionManager) OnlineDevices() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ids := make([]string, 0, len(cm.streams))
	for id := range cm.streams {
		ids = append(ids, id)
	}
	return ids
}

// Count：返回当前在线设备数
func (cm *ConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.streams)
}

// ─── gRPC 服务实现 ─────────────────────────────────────────

// DeviceGRPCServer：实现 DeviceServiceServer 接口
type DeviceGRPCServer struct {
	v1.UnimplementedDeviceServiceServer  // 嵌入未实现的基类，保证向前兼容
	store       store.Store                              // 数据存储
	connMgr     *ConnectionManager                       // 连接管理器
}

// CommandGRPCServer：实现 CommandServiceServer 接口
type CommandGRPCServer struct {
	v1.UnimplementedCommandServiceServer
	store   store.Store
	connMgr *ConnectionManager
}

// NewDeviceGRPCServer：创建设备服务端
func NewDeviceGRPCServer(s store.Store, cm *ConnectionManager) *DeviceGRPCServer {
	return &DeviceGRPCServer{
		store:   s,
		connMgr: cm,
	}
}

// NewCommandGRPCServer：创建指令服务端
func NewCommandGRPCServer(s store.Store, cm *ConnectionManager) *CommandGRPCServer {
	return &CommandGRPCServer{
		store:   s,
		connMgr: cm,
	}
}

// ─── DeviceService 实现 ───────────────────────────────────

// Heartbeat：处理设备心跳
// 1. 更新 devices 表的 last_heartbeat
// 2. 如果带了 capability，更新设备能力
// 3. 检查是否有待下发指令，通过响应告知设备
func (s *DeviceGRPCServer) Heartbeat(ctx context.Context, req *v1.HeartbeatRequest) (*v1.HeartbeatResponse, error) {
	log.Printf("[Heartbeat] 设备 %s 心跳到达", req.DeviceId)

	// 检查设备是否存在
	_, err := s.store.GetDevice(req.DeviceId)
	if err != nil {
		log.Printf("[Heartbeat] 设备 %s 不存在: %v", req.DeviceId, err)
		return &v1.HeartbeatResponse{
			HasPendingCommand: false,
			ServerTime:        time.Now().UnixMilli(),
		}, nil // 不返回错误，避免设备端重试风暴
	}

	// 如果心跳带了 capability，更新设备能力
	if req.Capability != nil {
		cap := device.DeviceCapability{
			FirmwareVersion:   req.Capability.FirmwareVersion,
			ProtoVersion:      req.Capability.ProtoVersion,
			SupportedFeatures: req.Capability.SupportedFeatures,
		}
		if err := s.store.UpdateDeviceCapabilities(req.DeviceId, cap); err != nil {
			log.Printf("[Heartbeat] 更新设备 %s 能力失败: %v", req.DeviceId, err)
		} else {
			log.Printf("[Heartbeat] 设备 %s 更新能力: firmware=%s features=%v",
				req.DeviceId, req.Capability.FirmwareVersion, req.Capability.SupportedFeatures)
		}
	}

	// 检查是否有待下发指令
	pending, err := s.store.GetPendingCommands(req.DeviceId)
	hasPending := err == nil && len(pending) > 0

	return &v1.HeartbeatResponse{
		HasPendingCommand: hasPending,
		ServerTime:        time.Now().UnixMilli(),
	}, nil
}

// ReportStatus：处理设备状态上报
// 更新 devices 表的状态、固件版本、配置、性能指标
func (s *DeviceGRPCServer) ReportStatus(ctx context.Context, req *v1.StatusReport) (*v1.StatusReportResponse, error) {
	log.Printf("[Status] 设备 %s 状态上报: status=%s firmware=%s",
		req.DeviceId, req.Status, req.FirmwareVersion)

	// 检查设备是否存在
	_, err := s.store.GetDevice(req.DeviceId)
	if err != nil {
		return &v1.StatusReportResponse{
			Accepted: false,
			Message:  fmt.Sprintf("设备不存在: %s", req.DeviceId),
		}, nil
	}

	// 如果有配置信息，更新配置
	if req.Config != nil {
		config := device.DeviceConfig{
			ScreenBrightness:  int(req.Config.ScreenBrightness),
			VolumeLevel:       int(req.Config.VolumeLevel),
			AutoRebootEnabled: req.Config.AutoRebootEnabled,
			AutoRebootTime:    req.Config.AutoRebootTime,
			CustomConfigJSON:  req.Config.CustomConfigJson,
		}
		if err := s.store.UpdateDeviceConfig(req.DeviceId, config); err != nil {
			log.Printf("[Status] 更新设备 %s 配置失败: %v", req.DeviceId, err)
		}
	}

	// 更新设备状态
	if req.Status != "" {
		if err := s.store.UpdateDeviceStatus(req.DeviceId, req.Status); err != nil {
			log.Printf("[Status] 更新设备 %s 状态失败: %v", req.DeviceId, err)
		}
	}

	return &v1.StatusReportResponse{
		Accepted: true,
		Message:  "ok",
	}, nil
}

// ReportEvent：处理设备事件上报
// 写入 fault_logs 表
func (s *DeviceGRPCServer) ReportEvent(ctx context.Context, req *v1.EventReport) (*v1.EventReportResponse, error) {
	log.Printf("[Event] 设备 %s 事件: type=%s severity=%s msg=%s",
		req.DeviceId, req.EventType, req.Severity, req.Message)

	// 检查设备是否存在
	_, err := s.store.GetDevice(req.DeviceId)
	if err != nil {
		return &v1.EventReportResponse{
			Accepted: false,
			Message:  fmt.Sprintf("设备不存在: %s", req.DeviceId),
		}, nil
	}

	// 生成日志 ID 和写入 fault_logs
	logEntry := device.FaultLog{
		ID:        fmt.Sprintf("LOG-%d", time.Now().UnixNano()),
		DeviceID:  req.DeviceId,
		Timestamp: time.UnixMilli(req.Timestamp),
		Type:      req.EventType,
		Severity:  req.Severity,
		Message:   req.Message,
		Resolved:  false,
	}
	if err := s.store.CreateFaultLog(logEntry); err != nil {
		log.Printf("[Event] 写入 fault_logs 失败: %v", err)
		return &v1.EventReportResponse{
			Accepted: false,
			Message:  fmt.Sprintf("写入日志失败: %v", err),
		}, nil
	}

	return &v1.EventReportResponse{
		Accepted: true,
		Message:  "ok",
	}, nil
}

// ReportCommandResult：处理设备指令执行结果回报
func (s *DeviceGRPCServer) ReportCommandResult(ctx context.Context, req *v1.CommandResult) (*v1.CommandResultResponse, error) {
	log.Printf("[CmdResult] 设备 %s 指令 %s 结果: status=%s msg=%s",
		req.DeviceId, req.CommandId, req.Status, req.Message)

	// 通过 proto UUID 更新指令状态
	err := s.store.UpdateCommandResultByUUID(req.CommandId, req.Status, req.Message)
	if err != nil {
		return &v1.CommandResultResponse{
			Accepted: false,
			Message:  fmt.Sprintf("更新指令结果失败: %v", err),
		}, nil
	}

	return &v1.CommandResultResponse{
		Accepted: true,
		Message:  "ok",
	}, nil
}

// ─── CommandService 实现 ──────────────────────────────────

// CommandStream：处理设备的指令流长连接
// 设备调用后建立 server streaming 连接
// 服务端有待下发指令时通过 stream 推送给设备
func (s *CommandGRPCServer) CommandStream(req *v1.CommandStreamRequest, stream grpc.ServerStreamingServer[v1.Command]) error {
	deviceID := req.DeviceId
	log.Printf("[CmdStream] 设备 %s 建立指令流连接", deviceID)

	// 注册到连接管理器
	s.connMgr.Register(deviceID, stream)
	defer s.connMgr.Unregister(deviceID)

	// 连接建立时，检查并推送积压的待下发指令
	pending, err := s.store.GetPendingCommands(deviceID)
	if err == nil && len(pending) > 0 {
		log.Printf("[CmdStream] 设备 %s 有 %d 条积压指令，推送中", deviceID, len(pending))
		for _, cmd := range pending {
			if !s.pushCommand(deviceID, cmd, stream) {
				return nil // 推送失败，连接已断
			}
		}
	}

	// 保持连接，等待新指令
	// 用 context 控制连接生命周期
	ctx := stream.Context()
	ticker := time.NewTicker(5 * time.Second) // 每 5 秒检查一次是否有新指令
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[CmdStream] 设备 %s 连接关闭: %v", deviceID, ctx.Err())
			return ctx.Err()

		case <-ticker.C:
			// 定期检查是否有新的待下发指令
			pending, err := s.store.GetPendingCommands(deviceID)
			if err != nil || len(pending) == 0 {
				continue
			}
			for _, cmd := range pending {
				if !s.pushCommand(deviceID, cmd, stream) {
					return nil
				}
			}
		}
	}
}

// pushCommand：将一条指令推送给设备，并更新数据库状态
func (s *CommandGRPCServer) pushCommand(deviceID string, cmd device.Command, stream grpc.ServerStreamingServer[v1.Command]) bool {
	// 转换成 proto 消息
	protoCmd := &v1.Command{
		CommandId:      cmd.CommandID,
		CommandType:    cmd.CommandType,
		PayloadJson:    cmd.PayloadJSON,
		IssuedAt:       cmd.IssuedAt,
		TimeoutSeconds: cmd.TimeoutSeconds,
	}

	// 通过 stream 推送
	err := stream.Send(protoCmd)
	if err != nil {
		log.Printf("[CmdStream] 向设备 %s 推送指令 %s 失败: %v", deviceID, cmd.CommandID, err)
		return false
	}

	// 更新数据库状态：pending → sent
	if err := s.store.UpdateCommandStatus(cmd.ID, device.CommandStatusSent, ""); err != nil {
		log.Printf("[CmdStream] 更新指令 %s 状态失败: %v", cmd.ID, err)
	}

	log.Printf("[CmdStream] 指令 %s 已推送给设备 %s (type=%s)", cmd.CommandID, deviceID, cmd.CommandType)
	return true
}

// ─── 注册到 gRPC Server ──────────────────────────────────

// RegisterServices：将两个服务注册到 gRPC server
func RegisterServices(s *grpc.Server, store store.Store) {
	connMgr := NewConnectionManager()

	v1.RegisterDeviceServiceServer(s, NewDeviceGRPCServer(store, connMgr))
	v1.RegisterCommandServiceServer(s, NewCommandGRPCServer(store, connMgr))

	log.Printf("[gRPC] DeviceService + CommandService 已注册，在线设备数: %d", connMgr.Count())
}
