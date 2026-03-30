// ============================================================
// main.go - CLI 入口文件
// ============================================================
// 这是整个 Go CLI 程序的起点
// 类似 package.json 里的 "bin" 字段
// ============================================================

package main

import (
	"github.com/QF1987/terminal-agent-go/internal/cmd"
)

func main() {
	// 调用 cmd 包的 Execute() 函数，启动 CLI
	// cmd 包里定义了所有子命令（list, info, stats 等）
	cmd.Execute()
}
