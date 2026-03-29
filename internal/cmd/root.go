package cmd

import (
	"fmt"
	"os"

	"github.com/QF1987/terminal-agent-go/internal/store"
	"github.com/spf13/cobra"
)

var (
	Store store.Store
	rootCmd = &cobra.Command{
		Use:   "device-ctl",
		Short: "自助购药机终端管理工具",
		Long: `Terminal Agent - AI 驱动的终端管理助手
自助购药机终端管理命令行工具，支持设备查询、监控、控制等功能。`,
	}
)

func init() {
	Store = store.NewMockStore()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
