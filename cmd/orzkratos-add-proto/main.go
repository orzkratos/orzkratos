// orzkratos-add-proto: Kratos proto file addition tool
// This tool simplifies the process of adding new proto files to Kratos projects
// Supports three usage methods:
//  1. Positional argument: orzkratos-add-proto demo.proto
//  2. Flag parameter: orzkratos-add-proto -name demo.proto
//  3. If neither specified, uses current DIR name as proto filename
//
// orzkratos-add-proto: Kratos 项目 proto 文件添加工具
// 这个工具简化了在 Kratos 项目中添加新 proto 文件的过程
// 支持三种使用方式：
//  1. 位置参数: orzkratos-add-proto demo.proto
//  2. flag 参数: orzkratos-add-proto -name demo.proto
//  3. 如果都不指定，则使用当前 DIR 名作为 proto 文件名
package main

import (
	"flag"
	"fmt"
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
	// Get current working DIR for project structure analysis and default proto name
	// 获取当前工作 DIR，用于确定项目结构和默认 proto 名称
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
	flag.StringVar(&protoName, "name", "", "proto-file-name. example: demo.proto / demo")
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

	// When no proto-name is provided (neither -name flag nor positional args), use current DIR name as default
	// 当没有传 proto-name 时，即，既没有 -name 传的参数也没有 args 传的参数，则使用当前 DIR 名作为默认值
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

	// Build the kratos command to be executed and show it to user for confirmation
	// 构建将要执行的 kratos 命令并显示给用户确认
	msg := fmt.Sprintf("cd %s && kratos proto add %s", projectPath, protoPath)
	zaplog.SUG.Debugln(msg)
	// Ask user to confirm command execution
	// 询问用户是否确认执行命令
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
