package utils

import (
	"os"
	"path/filepath"
	"unicode"

	"github.com/yyle88/erero"
	"github.com/yyle88/formatgo"
	"github.com/yyle88/must"
	"github.com/yyle88/osexistpath/osomitexist"
)

// CountFiles counts all files in the specified DIR recursively
// 递归统计指定 DIR 中的所有文件数量
func CountFiles(root string) (count int64, err error) {
	err = WalkFiles(root, func(path string, info os.FileInfo) error {
		count++
		return nil
	})
	if err != nil {
		return 0, erero.Wro(err)
	}
	return count, nil
}

// WalkFiles walks through all files in DIR and applies function to each file
// 遍历 DIR 中的所有文件并对每个文件执行函数
func WalkFiles(root string, run func(path string, info os.FileInfo) error) (err error) {
	err = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return erero.Wro(err)
			}
			if info.IsDir() {
				return nil
			}
			return run(path, info)
		},
	)
	if err != nil {
		return erero.Wro(err)
	}
	return nil
}

// IsFirstCharUpper checks if the first character of string is uppercase
// 检查字符串的第一个字符是否为大写
func IsFirstCharUpper(s string) bool {
	runes := []rune(s)
	if len(runes) > 0 {
		return unicode.IsUpper(runes[0])
	}
	return false
}

// LowerFirstChar converts the first character of string to lowercase
// 将字符串的第一个字符转换为小写
func LowerFirstChar(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}

// CopyBytes creates a deep copy of byte slice
// 创建字节切片的深度拷贝
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

// GetProjectPath finds project root by locating go.mod file
// 通过定位 go.mod 文件找到项目根路径
func GetProjectPath(currentPath string) (string, string) {
	projectPath := currentPath
	shortMiddle := ""
	for !osomitexist.IsFile(filepath.Join(projectPath, "go.mod")) {
		subName := filepath.Base(projectPath)

		prePath := filepath.Dir(projectPath)
		must.Different(prePath, projectPath)

		projectPath = prePath
		shortMiddle = filepath.Join(subName, shortMiddle)
	}
	return projectPath, shortMiddle
}
