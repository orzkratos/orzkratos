package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yyle88/runpath"
)

// TestGetProjectPath tests project root path detection via go.mod
// TestGetProjectPath 测试通过 go.mod 检测项目根路径
func TestGetProjectPath(t *testing.T) {
	path := runpath.PARENT.Path()
	t.Log(path)

	projectPath, shortMiddle := GetProjectPath(path)
	t.Log(projectPath)
	t.Log(shortMiddle)
}

// TestHasFiles tests file existence check in DIR
// TestHasFiles 测试 DIR 中的文件存在性检查
func TestHasFiles(t *testing.T) {
	exist, err := HasFiles(runpath.PARENT.Path())
	require.NoError(t, err)
	require.True(t, exist)
}
