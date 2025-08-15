// orzkratos-srv-proto: Kratos service-proto synchronization tool
// This tool automatically synchronizes service implementations with proto interface changes
// Supports automatic addition of missing methods, conversion of deleted methods to unexported,
// and reordering methods based on proto definition order
//
// Usage modes:
//  1. Sync specific proto: orzkratos-srv-proto -name demo.proto
//  2. Sync specific proto: orzkratos-srv-proto demo.proto
//  3. Sync all protos: orzkratos-srv-proto
//  4. Auto-confirm mode: orzkratos-srv-proto -auto
//
// orzkratos-srv-proto: Kratos 服务-proto 同步工具
// 此工具自动同步服务实现与 proto 接口变更
// 支持自动添加缺失方法、将删除的方法转换为非导出、
// 以及根据 proto 定义顺序重新排列方法
//
// 使用方式：
//  1. 同步特定 proto: orzkratos-srv-proto -name demo.proto
//  2. 同步特定 proto: orzkratos-srv-proto demo.proto
//  3. 同步所有 proto: orzkratos-srv-proto
//  4. 自动确认模式: orzkratos-srv-proto -auto
package main

import (
	"flag"
	"fmt"
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
	// Get current working DIR for project structure analysis
	// 获取当前工作 DIR，用于分析项目结构
	currentPath := rese.C1(os.Getwd())
	zaplog.LOG.Debug("current:", zap.String("path", currentPath))

	// Get current executable path for debugging information
	// 获取当前可执行文件路径，用于调试信息
	executePath := rese.C1(os.Executable())
	zaplog.LOG.Debug("execute:", zap.String("path", executePath))

	// Analyze project structure, get project root and relative path
	// projectPath: project root DIR
	// shortMiddle: relative path from project root to current DIR
	//
	// 分析项目结构，获取项目根目录和相对路径
	// projectPath: 项目根 DIR
	// shortMiddle: 从项目根 DIR 到当前 DIR 的相对路径
	projectPath, shortMiddle := utils.GetProjectPath(currentPath)
	zaplog.LOG.Debug("project:", zap.String("path", projectPath))

	// Define command line parameters
	// 定义命令行参数
	var protoName string
	flag.StringVar(&protoName, "name", "", "proto-filename. example: demo.proto / demo")
	var autoConfirm bool
	flag.BoolVar(&autoConfirm, "auto", false, "auto-confirm")
	flag.Parse()

	// Support positional arguments: prioritize the first positional argument from command line
	// 支持位置参数：优先使用命令行的第一个位置参数
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

	// Branch execution based on whether specific proto file is specified
	// 根据是否指定了特定的 proto 文件来分支执行
	if protoName != "" {
		// Sync specific proto file mode
		// 同步特定 proto 文件模式
		protoName = zerotern.VF(protoName, func() string {
			return filepath.Base(currentPath)
		})
		must.Nice(protoName)
		zaplog.LOG.Debug("protoName:", zap.String("protoName", protoName))

		// Build complete proto file path, auto add .proto suffix if needed
		// 构建完整的 proto 文件路径，如果需要则自动添加 .proto 后缀
		protoPath := tern.BVF(strings.HasSuffix(protoName, ".proto"), protoName, func() string {
			return protoName + ".proto"
		})
		// Join relative path and filename into complete path
		// 将相对路径和文件名拼接成完整路径
		protoPath = filepath.Join(shortMiddle, protoPath)
		zaplog.LOG.Debug("protoPath:", zap.String("protoPath", protoPath))

		// Ask user to confirm single proto sync (unless auto-confirm enabled)
		// 询问用户确认单个 proto 同步（除非启用自动确认）
		if !autoConfirm && !chooseConfirm("execute sync kratos service once?") {
			return
		}
		// Sync services for specific proto file
		// 为特定 proto 文件同步服务
		synckratos.GenServicesOnce(projectPath, protoPath)
	} else {
		// Sync all proto files mode. Ask user to confirm full service sync (unless auto-confirm enabled)
		// 同步所有 proto 文件模式。需要询问用户确认完整服务同步（除非启用自动确认）
		if !autoConfirm && !chooseConfirm("execute sync kratos service code?") {
			return
		}
		// Sync all services in the project
		// 同步项目中的所有服务
		synckratos.GenServicesCode(projectPath)
	}
}

// chooseConfirm displays a confirmation dialog for user to choose whether to continue execution
// Uses survey library to provide interactive Y/N selection interface
//
// chooseConfirm 显示确认对话框，让用户选择是否继续执行
// 使用 survey 库提供交互式的 Y/N 选择界面
func chooseConfirm(msg string) bool {
	// Variable to store user's answer
	// 用于存储用户的回答
	var input bool

	// Define confirmation question with default choice as Yes (true)
	// 定义确认问题，默认选择为 Yes (true)
	prompt := &survey.Confirm{
		Message: msg,
		Default: true, // Default value if user directly presses Enter // 默认值，如果用户直接按回车
	}

	// Run prompt and capture user input
	// 运行提示并捕获用户输入的内容
	done.Done(survey.AskOne(prompt, &input))

	// Output user's answer and return result
	// 输出用户的回答并返回结果
	if input {
		fmt.Println("You chose Yes")
		return true
	} else {
		fmt.Println("You chose Not")
		return false
	}
}
