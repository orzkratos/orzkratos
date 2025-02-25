package utils

import (
	"os"
	"path/filepath"
	"unicode"

	"github.com/yyle88/must"
	"github.com/yyle88/osexistpath/osomitexist"
)

func CntFileNum(root string) (fileNum int64, err error) {
	err = Files(root, func(path string, info os.FileInfo) error {
		fileNum++
		return nil
	})
	if err != nil {
		return 0, err
	}
	return fileNum, nil
}

func Files(root string, run func(path string, info os.FileInfo) error) (err error) {
	err = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			return run(path, info)
		},
	)
	return err
}

func C0IsUpper(s string) bool {
	runes := []rune(s)
	if len(runes) > 0 {
		return unicode.IsUpper(runes[0])
	}
	return false
}

func CvtC0Lower(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}

func Clone[V any](org []V) (dst []V) {
	dst = make([]V, len(org)) //看来得提前分配空间否则不能拷贝
	copy(dst, org)
	return dst
}

func SoftLast[V any](a []V) (v V) {
	if n := len(a); n > 0 {
		return a[n-1]
	}
	return
}

func GetProjectPath(currentPath string) (string, string) {
	projectPath := currentPath
	shortMiddle := ""
	for {
		if osomitexist.IsFile(filepath.Join(projectPath, "go.mod")) {
			break
		}
		subName := filepath.Base(projectPath)

		prePath := filepath.Dir(projectPath)
		must.Different(prePath, projectPath)

		projectPath = prePath
		shortMiddle = filepath.Join(subName, shortMiddle)
	}
	return projectPath, shortMiddle
}
