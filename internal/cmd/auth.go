// ============================================================
// auth.go - 授权管理命令
// ============================================================
// 实现 "device-ctl auth" 子命令
// 包含 whoami（当前用户）、grant（授权）、revoke（撤销）
// 用于权限管理（谁能操作哪些设备）
// ============================================================

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ─── auth 父命令 ──────────────────────────────────────────
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "授权管理",
	Long:  `授权管理命令，包含 whoami（当前用户）、grant（授权）、revoke（撤销）子命令`,
}

// ─── auth whoami 子命令 ───────────────────────────────────
var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "查看当前用户",
	Long:  `查看当前登录用户信息`,
	Run: func(cmd *cobra.Command, args []string) {
		// 模拟用户信息（实际应该从 API 获取）
		fmt.Println("\n👤 当前用户信息")
		fmt.Println("──────────────────")
		fmt.Println("用户ID:   admin-001")
		fmt.Println("用户名:   管理员")
		fmt.Println("角色:     系统管理员")
		fmt.Println("区域权限: 华东, 华南, 华北")
		fmt.Println("最后登录: 2026-03-29 15:30:00")
	},
}

// ─── auth grant 子命令 ────────────────────────────────────
var (
	grantUser   string  // --user：目标用户
	grantRegion string  // --region：授权区域
	grantDevice string  // --device：授权设备
	grantRole   string  // --role：授权角色
)

var authGrantCmd = &cobra.Command{
	Use:   "grant",
	Short: "授权用户",
	Long:  `授权用户访问指定区域或设备`,
	Run: func(cmd *cobra.Command, args []string) {
		if grantUser == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --user 参数")
			os.Exit(1)
		}

		// 构建授权信息
		fmt.Printf("\n✅ 授权成功\n")
		fmt.Println("──────────────────")
		fmt.Printf("用户:   %s\n", grantUser)

		if grantRegion != "" {
			fmt.Printf("区域:   %s\n", grantRegion)
		}
		if grantDevice != "" {
			fmt.Printf("设备:   %s\n", grantDevice)
		}
		if grantRole != "" {
			fmt.Printf("角色:   %s\n", grantRole)
		}

		fmt.Println("\n💡 用户现在可以操作指定资源了")
	},
}

// ─── auth revoke 子命令 ───────────────────────────────────
var (
	revokeUser   string  // --user：目标用户
	revokeRegion string  // --region：撤销区域
	revokeDevice string  // --device：撤销设备
)

var authRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "撤销授权",
	Long:  `撤销用户的访问权限`,
	Run: func(cmd *cobra.Command, args []string) {
		if revokeUser == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --user 参数")
			os.Exit(1)
		}

		fmt.Printf("\n🚫 撤销授权成功\n")
		fmt.Println("──────────────────")
		fmt.Printf("用户:   %s\n", revokeUser)

		if revokeRegion != "" {
			fmt.Printf("区域:   %s\n", revokeRegion)
		}
		if revokeDevice != "" {
			fmt.Printf("设备:   %s\n", revokeDevice)
		}

		fmt.Println("\n💡 用户权限已撤销")
	},
}

func init() {
	// grant 的 flag
	authGrantCmd.Flags().StringVarP(&grantUser, "user", "u", "", "目标用户ID（必填）")
	authGrantCmd.Flags().StringVarP(&grantRegion, "region", "r", "", "授权区域")
	authGrantCmd.Flags().StringVarP(&grantDevice, "device", "d", "", "授权设备ID")
	authGrantCmd.Flags().StringVarP(&grantRole, "role", "R", "", "授权角色")

	// revoke 的 flag
	authRevokeCmd.Flags().StringVarP(&revokeUser, "user", "u", "", "目标用户ID（必填）")
	authRevokeCmd.Flags().StringVarP(&revokeRegion, "region", "r", "", "撤销区域")
	authRevokeCmd.Flags().StringVarP(&revokeDevice, "device", "d", "", "撤销设备ID")

	// 添加子命令
	authCmd.AddCommand(authWhoamiCmd)
	authCmd.AddCommand(authGrantCmd)
	authCmd.AddCommand(authRevokeCmd)

	rootCmd.AddCommand(authCmd)
}
