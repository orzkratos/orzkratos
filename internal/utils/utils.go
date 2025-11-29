// Package utils provides common utilities used across the orzkratos project
// Contains file operations, string manipulation, and path utilities
//
// utils 包提供 orzkratos 项目中使用的通用工具函数
// 包含文件操作、字符串处理和路径工具
package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"unicode"

	"github.com/yyle88/erero"
	"github.com/yyle88/formatgo"
	"github.com/yyle88/must"
	"github.com/yyle88/osexistpath/osomitexist"
)

// HasFiles checks if DIR contains files in depth
// Returns true on first file found, stops walk at once
//
// HasFiles 深度检查 DIR 中是否包含文件
// 找到第一个文件时返回 true 并立即停止遍历
func HasFiles(root string) (bool, error) {
	var exist bool
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			exist = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return false, erero.Wro(err)
	}
	return exist, nil
}

// IsFirstCharUpper checks if first rune of string is uppercase
// 检查字符串的第一个字符是否为大写
func IsFirstCharUpper(s string) bool {
	runes := []rune(s)
	if len(runes) > 0 {
		return unicode.IsUpper(runes[0])
	}
	return false
}

// LowerFirstChar converts the first rune of string to lowercase
// 将字符串的第一个字符转换为小写
func LowerFirstChar(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}

// CopyBytes creates a clone of byte slice
// 创建字节切片的克隆
func CopyBytes(src []byte) []byte {
	dst := make([]byte, len(src)) // Allocate space before copying // 复制前需要分配空间
	copy(dst, src)
	return dst
}

// FormatAndWriteCode formats raw Go code and writes to file
// 对原始 Go 代码进行格式化并写入文件
func FormatAndWriteCode(path string, data []byte) {
	code, _ := formatgo.FormatBytes(data)
	must.Have(code)
	must.Done(os.WriteFile(path, code, 0644))
}

// GetProjectPath finds project root via go.mod file location
// Returns project root path and relative path from current to root
//
// GetProjectPath 通过定位 go.mod 文件找到项目根路径
// 返回项目根路径和从当前位置到根路径的相对路径
func GetProjectPath(currentPath string) (string, string) {
	projectPath := currentPath
	shortMiddle := ""
	for !osomitexist.IsFile(filepath.Join(projectPath, "go.mod")) {
		subName := filepath.Base(projectPath) // Extract current DIR name // 提取当前 DIR 名称

		prePath := filepath.Dir(projectPath)
		must.Different(prePath, projectPath) // Ensure not stuck at root // 确保没有卡在根路径

		projectPath = prePath
		shortMiddle = filepath.Join(subName, shortMiddle) // Build relative path // 构建相对路径
	}
	return projectPath, shortMiddle
}
