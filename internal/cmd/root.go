// ============================================================
// root.go - 根命令定义
// ============================================================
// 这里定义了 CLI 的根命令（device-ctl）
// 以及全局变量（比如数据存储 Store）
// ============================================================

package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/QF1987/terminal-agent-go/internal/store"
	"github.com/spf13/cobra"
)

// 全局变量：数据存储
// 所有子命令都可以通过 Store 访问数据
// 优先使用 PostgreSQL，失败则 fallback 到 MockStore
var (
	Store store.Store

	// rootCmd：根命令，相当于 "device-ctl" 这个命令本身
	// Use: 命令名称
	// Short: 简短描述（help 时显示）
	// Long: 详细描述
	rootCmd = &cobra.Command{
		Use:   "device-ctl",
		Short: "自助购药机终端管理工具",
		Long: `Terminal Agent - AI 驱动的终端管理助手
自助购药机终端管理命令行工具，支持设备查询、监控、控制等功能。`,
	}
)

// init() 函数：Go 的特殊函数，在 main() 之前自动执行
// 这里用来初始化全局变量
func init() {
	// 优先尝试连接 PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://deviceops:deviceops123@localhost:5432/deviceops?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err == nil {
		// 测试连接
		if err = db.Ping(); err == nil {
			// 建表
			if err = store.InitSchema(db); err == nil {
				// Seed 数据（如果表为空）
				if err = store.SeedData(db); err == nil {
					Store = store.NewPGStore(db)
					return
				}
			}
		}
		db.Close()
	}

	// PostgreSQL 连接失败，使用 MockStore
	Store = store.NewMockStore()
}

// Execute()：执行根命令
// 这是程序的真正入口，main() 只是调用它
func Execute() {
	// rootCmd.Execute()：执行 CLI 解析和命令分发
	// 如果出错，打印错误信息并退出（exit code 1）
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
