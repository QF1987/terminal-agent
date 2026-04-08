// ============================================================
// serve.go - gRPC 服务启动命令
// ============================================================
// device-ctl serve --port 9090
// 启动 gRPC 服务，监听设备心跳/状态/事件上报和指令推送
// ============================================================

package cmd

import (
	"fmt"       // 格式化输出
	"log"       // 日志
	"net"       // 网络监听
	"os"        // 系统调用
	"os/signal" // 信号处理
	"syscall"   // 系统信号

	"github.com/QF1987/terminal-agent-go/internal/server" // gRPC 服务端实现
	"github.com/spf13/cobra"                              // CLI 框架
	"google.golang.org/grpc"                              // gRPC 框架
)

var (
	servePort int // 监听端口，默认 9090
)

// serveCmd：gRPC 服务启动命令
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 gRPC 服务",
	Long: `启动 terminal-agent gRPC 服务，监听设备连接。

设备通过 gRPC 协议连接后可以：
  - 发送心跳（Heartbeat）
  - 上报状态（ReportStatus）
  - 上报事件（ReportEvent）
  - 接收指令（CommandStream）
  - 回报指令执行结果（ReportCommandResult）

示例：
  device-ctl serve                    # 默认监听 :9090
  device-ctl serve --port 50051       # 指定端口`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建 TCP 监听器
		addr := fmt.Sprintf(":%d", servePort)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("监听 %s 失败: %v", addr, err)
		}

		// 创建 gRPC server
		grpcServer := grpc.NewServer()

		// 注册 DeviceService + CommandService
		server.RegisterServices(grpcServer, Store)

		// 优雅关停：监听 SIGINT/SIGTERM
		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			sig := <-sigCh
			log.Printf("收到信号 %v，正在关闭 gRPC 服务...", sig)
			grpcServer.GracefulStop()
		}()

		// 启动服务
		log.Printf("🚀 terminal-agent gRPC 服务已启动，监听 %s", addr)
		log.Printf("📡 DeviceService: Heartbeat / ReportStatus / ReportEvent / ReportCommandResult")
		log.Printf("📡 CommandService: CommandStream (server streaming)")
		log.Printf("按 Ctrl+C 停止服务")

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC 服务异常: %v", err)
		}
	},
}

func init() {
	// 注册 serve 子命令到根命令
	rootCmd.AddCommand(serveCmd)

	// 定义 --port 参数
	// Flags().IntVarP(&变量, "长名", "短名", 默认值, "帮助说明")
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 9090, "gRPC 监听端口")
}
