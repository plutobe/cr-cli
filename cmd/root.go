package cmd

import (
	"cr-cli/internal/i18n"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "cr-cli",
	Short: i18n.T("AI 驱动的代码审查工具", "AI-powered code review tool"),
	Long:  i18n.T("CR-CLI 是一个 AI 驱动的代码审查命令行工具。", "CR-CLI is an AI-powered code review CLI tool."),
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", i18n.T("配置文件路径 (默认查找 cr-cli.yaml)", "Config file path (default: cr-cli.yaml)"))
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, i18n.T("打印调试信息（请求/响应报文）", "Print debug info (request/response)"))
}

func Execute() error {
	return rootCmd.Execute()
}
