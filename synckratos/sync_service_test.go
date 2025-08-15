package synckratos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
)

func TestParseServiceFile(t *testing.T) {
	// Create temp DIR using modern os.MkdirTemp
	// 使用现代化的 os.MkdirTemp 创建临时 DIR
	tempRoot := rese.C1(os.MkdirTemp("", "orzkratos_test_*"))
	defer func() {
		must.Done(os.RemoveAll(tempRoot))
	}()

	testContent := `
package service

type GreeterService struct{}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {}
func (s *GreeterService) SayWorld(ctx context.Context, in *v1.HelloRequest) (*v1.WorldReply, error) {}
`
	testFile := filepath.Join(tempRoot, "greeter.go")
	must.Done(os.WriteFile(testFile, []byte(testContent), 0644))

	// Test parseServiceFile
	// 测试 parseServiceFile
	serviceFile := parseServiceFile(testFile)
	require.NotNil(t, serviceFile)
	require.Equal(t, testFile, serviceFile.path)
	require.NotEmpty(t, serviceFile.serviceStructsMap)

	// Check if GreeterService exists
	// 检查 GreeterService 是否存在
	greeterStruct, exists := serviceFile.serviceStructsMap["GreeterService"]
	require.True(t, exists)
	require.Len(t, greeterStruct.methods, 2)
}

func TestSearchMissingMethods(t *testing.T) {
	// Create temp DIR using modern os.MkdirTemp
	// 使用现代化的 os.MkdirTemp 创建临时 DIR
	tempRoot := rese.C1(os.MkdirTemp("", "orzkratos_missing_*"))
	defer func(path string) {
		must.Done(os.RemoveAll(path))
	}(tempRoot)

	// Create old service with fewer methods
	// 创建方法较少的旧服务
	oldContent := `package service

type GreeterService struct{}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {}
`
	oldFile := filepath.Join(tempRoot, "old_greeter.go")
	must.Done(os.WriteFile(oldFile, []byte(oldContent), 0644))

	// Create new service with more methods
	// 创建方法较多的新服务
	newContent := `package service

type GreeterService struct{}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {}
func (s *GreeterService) SayWorld(ctx context.Context, in *v1.HelloRequest) (*v1.WorldReply, error) {}
`
	newFile := filepath.Join(tempRoot, "new_greeter.go")
	must.Done(os.WriteFile(newFile, []byte(newContent), 0644))

	// Parse both files
	// 解析两个文件
	oldService := parseServiceFile(oldFile)
	newService := parseServiceFile(newFile)

	// Search for missing methods
	// 搜索缺失的方法
	missingCode := searchMissingMethods(oldService, newService)
	require.NotEmpty(t, missingCode)

	// Should contain SayWorld method
	// 应该包含 SayWorld 方法
	t.Log("Missing methods found:", missingCode)
}
