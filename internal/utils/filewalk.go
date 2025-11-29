// Package utils file walk utilities: Smart file system navigation in code scan
// Provides pattern-based file matching and path walk functions
// Features customizable suffix matching and callback-based file processing
// Optimized in Kratos project structure scan with flexible options
//
// utils 文件遍历工具：代码扫描中的智能文件系统导航
// 提供基于模式的文件匹配和路径遍历功能
// 具有可定制的后缀匹配和基于回调的文件处理
// 针对 Kratos 项目结构扫描优化，具有灵活的选项
package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yyle88/erero"
)

// SuffixPattern provides intelligent file extension matching in selective file processing
// Contains configurable suffix patterns with flexible file filtering operations
// Supports multiple suffix matching with optimized string comparison algorithms
//
// SuffixPattern 提供智能的文件扩展名匹配，用于选择性文件处理
// 包含可配置的后缀模式，支持灵活的文件过滤操作
// 支持多后缀匹配，具有优化的字符串比较算法
type SuffixPattern struct {
	suffixes []string // List of file suffixes used in matching // 用于匹配的文件后缀列表
}

// NewSuffixPattern creates a new SuffixPattern with specified suffix patterns
// Initializes pattern with custom suffix list to select target files
// Returns configured pattern set to file pattern matching operations
//
// NewSuffixPattern 创建一个具有指定后缀模式的新 SuffixPattern
// 使用自定义后缀列表初始化模式，用于目标文件选择
// 返回配置好的模式，准备进行文件模式匹配操作
func NewSuffixPattern(suffixes []string) *SuffixPattern {
	return &SuffixPattern{
		suffixes: suffixes,
	}
}

// Match performs suffix-based string matching against configured patterns
// Tests if input string ends with one of the predefined suffixes
// Returns true when match found, false otherwise to enable efficient filtering
//
// Match 对配置的模式执行基于后缀的字符串匹配
// 测试输入字符串是否以任何预定义后缀结尾
// 如果找到匹配则返回 true，否则返回 false 以实现高效过滤
func (sp *SuffixPattern) Match(s string) bool {
	for _, suffix := range sp.suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// WalkFiles performs path walk with intelligent file filtering
// Applies callback function to files matching the specified suffix patterns
// Handles issues and skips non-matching files in walk process
// Returns aggregated issue from walk process and callback execution
//
// WalkFiles 执行带有智能文件过滤的路径遍历
// 对匹配指定后缀模式的文件应用回调函数
// 提供全面的错误处理并自动跳过不匹配的文件
// 返回来自遍历或回调执行失败的聚合错误
func WalkFiles(root string, suffixPattern *SuffixPattern, run func(path string, info os.FileInfo) error) error {
	if err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return erero.Wro(err)
			}
			if info.IsDir() {
				return nil
			}
			if suffixPattern.Match(path) {
				return run(path, info)
			}
			return nil
		},
	); err != nil {
		return erero.Wro(err)
	}
	return nil
}
