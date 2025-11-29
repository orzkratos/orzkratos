package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yyle88/runpath"
)

// TestWalkFiles tests path walking with suffix pattern matching
// TestWalkFiles 测试带后缀模式匹配的路径遍历
func TestWalkFiles(t *testing.T) {
	require.NoError(t, WalkFiles(runpath.PARENT.Path(), NewSuffixPattern([]string{".go"}), func(path string, info os.FileInfo) error {
		t.Log(path)
		return nil
	}))
}
