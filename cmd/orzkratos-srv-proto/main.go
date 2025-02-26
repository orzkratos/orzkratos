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
	currentPath := rese.C1(os.Getwd())
	zaplog.LOG.Debug("current:", zap.String("path", currentPath))

	executePath := rese.C1(os.Executable())
	zaplog.LOG.Debug("execute:", zap.String("path", executePath))

	projectPath, shortMiddle := utils.GetProjectPath(currentPath)
	zaplog.LOG.Debug("project:", zap.String("path", projectPath))

	var protoName string
	flag.StringVar(&protoName, "name", "", "proto-file-name. example: demo.proto / demo")
	flag.Parse()

	if protoName != "" {
		protoName = zerotern.VF(protoName, func() string {
			return filepath.Base(currentPath)
		})
		must.Nice(protoName)
		zaplog.LOG.Debug("protoName:", zap.String("protoName", protoName))

		protoPath := tern.BVF(strings.HasSuffix(protoName, ".proto"), protoName, func() string {
			return protoName + ".proto"
		})
		protoPath = filepath.Join(shortMiddle, protoPath)
		zaplog.LOG.Debug("protoPath:", zap.String("protoPath", protoPath))

		if !chooseConfirm("execute sync kratos service once?") {
			return
		}
		synckratos.GenServicesOnce(projectPath, protoPath)
	} else {
		if !chooseConfirm("execute sync kratos service code?") {
			return
		}
		synckratos.GenServicesCode(projectPath)
	}
}

func chooseConfirm(msg string) bool {
	// 用于存储用户的回答
	var input bool

	// 定义确认问题
	prompt := &survey.Confirm{
		Message: msg,
		Default: true, // 默认值，如果用户直接按回车
	}

	// 运行提示并捕获用户输入的内容
	done.Done(survey.AskOne(prompt, &input))

	// 输出用户的回答
	if input {
		fmt.Println("You chose Yes")
		return true
	} else {
		fmt.Println("You chose Not")
		return false
	}
}
