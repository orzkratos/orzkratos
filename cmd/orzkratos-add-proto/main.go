// orzkratos-add-proto: Kratos proto file addition CLI
// Adds new proto files to Kratos projects with ease
//
// Usage modes:
//  1. Position arg: orzkratos-add-proto demo.proto
//  2. Flag: orzkratos-add-proto -name demo.proto
//  3. No arg: uses current DIR name as proto filename
//
// orzkratos-add-proto: Kratos proto 文件添加命令行
// 简化向 Kratos 项目添加新 proto 文件的流程
//
// 使用方式：
//  1. 位置参数: orzkratos-add-proto demo.proto
//  2. flag 参数: orzkratos-add-proto -name demo.proto
//  3. 无参数: 使用当前 DIR 名作为 proto 文件名
package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/orzkratos/orzkratos/internal/utils"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/osexec"
	"github.com/yyle88/rese"
	"github.com/yyle88/tern"
	"github.com/yyle88/tern/zerotern"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

func main() {
	// Get current working DIR to analyze project structure and set default proto name
	// 获取当前工作 DIR，用于确定项目结构和默认 proto 名称
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
	flag.StringVar(&protoName, "name", "", "proto-file-name. example: demo.proto / demo")
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

	// When no proto-name is given, use current DIR name as default
	// 当没有传 proto-name 时，使用当前 DIR 名作为默认值
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

	// Build the kratos command to be executed and show to get confirmation
	// 构建将要执行的 kratos 命令并显示给用户确认
	zaplog.LOG.Debug("command to run", zap.String("root", projectPath), zap.String("cmd", "kratos proto add "+protoPath))
	// Ask to confirm command execution
	// 确认命令执行
	if !chooseConfirm("execute kratos proto add?") {
		return
	}

	// Execute kratos proto add command in project root DIR
	// Example: "kratos proto add api/helloworld/demo.proto"
	//
	// 在项目根 DIR 执行 kratos proto add 命令
	// 示例: "kratos proto add api/helloworld/demo.proto"
	output := rese.V1(osexec.ExecInPath(projectPath, "kratos", "proto", "add", protoPath))
	zaplog.SUG.Debugln(string(output))
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
