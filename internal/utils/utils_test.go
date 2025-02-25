package utils

import (
	"testing"

	"github.com/yyle88/runpath"
)

func TestGetProjectPath(t *testing.T) {
	path := runpath.PARENT.Path()
	t.Log(path)

	projectPath, shortMiddle := GetProjectPath(path)
	t.Log(projectPath)
	t.Log(shortMiddle)
}
