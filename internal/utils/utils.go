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

func CloneBytes(org []byte) (dst []byte) {
	dst = make([]byte, len(org)) //看来得提前分配空间否则不能拷贝
	copy(dst, org)
	return dst
}

func WriteFormatBytes(data []byte, path string) {
	code, _ := formatgo.FormatBytes(data)
	must.Have(code)
	must.Done(os.WriteFile(path, code, 0644))
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
