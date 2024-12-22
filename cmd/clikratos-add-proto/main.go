package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/yyle88/done"
	"github.com/yyle88/osexec"
	"github.com/yyle88/osexistpath/osomitexist"
	"github.com/yyle88/rese"
	"github.com/yyle88/tern"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

func main() {
	currentPath := rese.C1(os.Getwd())
	zaplog.LOG.Debug("apipath:", zap.String("path", currentPath))

	executePath := rese.C1(os.Executable())
	zaplog.LOG.Debug("execute:", zap.String("path", executePath))

	var name string
	flag.StringVar(&name, "name", "", "proto-file-name. example: demo.proto / demo")
	flag.Parse()

	protoPath := tern.BVF(strings.HasSuffix(name, ".proto"), name, func() string {
		return name + ".proto"
	})

	projectPath := currentPath
	for {
		if osomitexist.IsFile(filepath.Join(projectPath, "Makefile")) {
			break
		}
		subName := filepath.Base(projectPath)

		prePath := filepath.Dir(projectPath)
		if prePath == projectPath {
			zaplog.SUG.Errorln("wrong")
			return
		}
		projectPath = prePath

		protoPath = filepath.Join(subName, protoPath)
	}

	msg := fmt.Sprintf("cd %s && kratos proto add %s", projectPath, protoPath)
	zaplog.SUG.Debugln(msg)
	if !chooseConfirm("execute kratos proto add?") {
		return
	}

	// "kratos proto add api/helloworld/demo.proto"
	output := rese.V1(osexec.ExecInPath(projectPath, "kratos", "proto", "add", protoPath))
	zaplog.SUG.Debugln(string(output))
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
