// orzkratos-srv-proto: Kratos service-proto sync CLI
// Auto syncs service code with proto changes: add missing methods, unexport deleted, sort methods
// Default uses mask mode (match via embedded Unimplemented*Server type)
//
// Usage modes:
//  1. Sync one proto: orzkratos-srv-proto -name demo.proto
//  2. Sync one proto: orzkratos-srv-proto demo.proto
//  3. Sync each proto: orzkratos-srv-proto
//  4. Auto-confirm mode: orzkratos-srv-proto -auto
//  5. Mask mode (default): orzkratos-srv-proto -mask
//  6. Disable mask mode: orzkratos-srv-proto -mask=false
//
// orzkratos-srv-proto: Kratos 服务-proto 同步命令行
// 自动同步服务代码与 proto 变更：添加缺失方法、非导出已删除方法、排序方法
// 默认使用 mask 模式（按嵌入的 Unimplemented*Server 类型匹配）
//
// 使用方式：
//  1. 同步单个 proto: orzkratos-srv-proto -name demo.proto
//  2. 同步单个 proto: orzkratos-srv-proto demo.proto
//  3. 同步所有 proto: orzkratos-srv-proto
//  4. 自动确认模式: orzkratos-srv-proto -auto
//  5. Mask 模式（默认）: orzkratos-srv-proto -mask
//  6. 禁用 mask 模式: orzkratos-srv-proto -mask=false
package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/orzkratos/orzkratos/internal/utils"
	"github.com/orzkratos/orzkratos/synckratos"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"github.com/yyle88/tern"
	"github.com/yyle88/tern/zerotern"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

func main() {
	// Get current working DIR to analyze project structure
	// 获取当前工作 DIR，用于分析项目结构
	currentPath := rese.C1(os.Getwd())
	zaplog.LOG.Debug("current path", zap.String("path", currentPath))

	// Get current executable path as debug info
	// 获取当前可执行文件路径，用于调试信息
	executePath := rese.C1(os.Executable())
	zaplog.LOG.Debug("execute path", zap.String("path", executePath))

	// Analyze project structure, get project root and relative path
	// projectPath: project root DIR
	// shortMiddle: relative path from project root to current DIR
	//
	// 分析项目结构，获取项目根目录和相对路径
	// projectPath: 项目根 DIR
	// shortMiddle: 从项目根 DIR 到当前 DIR 的相对路径
	projectPath, shortMiddle := utils.GetProjectPath(currentPath)
	zaplog.LOG.Debug("project path", zap.String("path", projectPath))

	// Define command line parameters
	// 定义命令行参数
	var protoName string
	flag.StringVar(&protoName, "name", "", "proto-filename. example: demo.proto / demo")
	var autoConfirm bool
	flag.BoolVar(&autoConfirm, "auto", false, "auto-confirm")
	var maskMode bool
	flag.BoolVar(&maskMode, "mask", true, "mask mode: match via embedded Unimplemented*Server type")
	flag.Parse()

	// Handle position args: use the first arg from command line
	// 处理位置参数：使用命令行的第一个参数
	if args := flag.Args(); len(args) > 0 {
		// Cannot use both methods to pass proto-name as it would cause confusion
		// 两种方式传 proto-name 但不能同时传否则会造成混乱的
		if protoName != "" {
			zaplog.LOG.Panic("duplicate proto-name: cannot use both -name flag and args name")
		}
		if len(args) > 1 {
			zaplog.LOG.Panic("multiple proto-names: cannot use more than one args proto name")
		}
		protoName = args[0]
	}

	// Execute based on proto file specification
	// 根据是否指定 proto 文件来执行
	if protoName != "" {
		// Sync specific proto file mode
		// 同步特定 proto 文件模式
		protoName = zerotern.VF(protoName, func() string {
			return filepath.Base(currentPath)
		})
		must.Nice(protoName)
		zaplog.LOG.Debug("proto name", zap.String("name", protoName))

		// Build complete proto file path, auto add .proto suffix if needed
		// 构建完整的 proto 文件路径，如果需要则自动添加 .proto 后缀
		protoPath := tern.BVF(strings.HasSuffix(protoName, ".proto"), protoName, func() string {
			return protoName + ".proto"
		})
		// Join relative path and filename into complete path
		// 将相对路径和文件名拼接成完整路径
		protoPath = filepath.Join(shortMiddle, protoPath)
		zaplog.LOG.Debug("proto path", zap.String("path", protoPath))

		// Ask to confirm single proto sync (unless auto-confirm enabled)
		// 确认单个 proto 同步（除非启用自动确认）
		if !autoConfirm && !chooseConfirm("execute sync kratos service once?") {
			return
		}
		// Sync services with the specific proto file
		// 同步特定 proto 文件的服务
		synckratos.GenServicesOnce(projectPath, protoPath, &synckratos.SyncOptions{MaskMode: maskMode})
	} else {
		// Sync each proto file mode. Ask to confirm service sync (unless auto-confirm enabled)
		// 同步所有 proto 文件模式。确认服务同步（除非启用自动确认）
		if !autoConfirm && !chooseConfirm("execute sync kratos service code?") {
			return
		}
		// Sync each service in the project
		// 同步项目中的所有服务
		synckratos.GenServicesCode(projectPath, &synckratos.SyncOptions{MaskMode: maskMode})
	}
}

// chooseConfirm shows a confirmation prompt with Y/N selection
// chooseConfirm 显示确认提示，提供 Y/N 选择
func chooseConfirm(msg string) bool {
	// Save the input response
	// 保存输入的响应
	var input bool

	// Define confirmation question with default choice as Yes (true)
	// 定义确认问题，默认选择为 Yes (true)
	prompt := &survey.Confirm{
		Message: msg,
		Default: true, // Default when user just presses Enter // 默认值，如果用户直接按回车
	}

	// Run prompt and capture input
	// 运行提示并捕获输入
	done.Done(survey.AskOne(prompt, &input))

	// Output the choice and return
	// 输出选择并返回
	if input {
		zaplog.SUG.Infoln("You chose Yes")
		return true
	}
	zaplog.SUG.Infoln("You chose Not")
	return false
}
