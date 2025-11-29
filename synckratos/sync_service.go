// Package synckratos provides service code sync with proto definitions
// Auto syncs service methods when proto files change: add new, unexport removed, sort existing
//
// synckratos 包提供服务代码与 proto 定义的同步功能
// proto 文件变更时自动同步服务方法：新增、非导出已删除、排序现有方法
package synckratos

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/orzkratos/astkratos"
	"github.com/orzkratos/orzkratos/internal/utils"
	"github.com/yyle88/eroticgo"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/osexec"
	"github.com/yyle88/osexistpath/osmustexist"
	"github.com/yyle88/osexistpath/ossoftexist"
	"github.com/yyle88/printgo"
	"github.com/yyle88/rese"
	"github.com/yyle88/sortx"
	"github.com/yyle88/syntaxgo/syntaxgo_ast"
	"github.com/yyle88/syntaxgo/syntaxgo_astnode"
	"github.com/yyle88/syntaxgo/syntaxgo_search"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

// SyncOptions defines options used in service synchronization
// SyncOptions 定义服务同步中使用的选项
type SyncOptions struct {
	MaskMode bool // Match via Unimplemented*Server type instead of filename // 按 Unimplemented*Server 类型匹配而非文件名
}

// GenServicesCode syncs each service file in project with proto definitions
// Scans api/ DIR, generates missing services, and syncs existing ones
//
// GenServicesCode 将项目中的所有服务文件与 proto 定义同步
// 扫描 api/ DIR，生成缺失的服务，并同步现有服务
func GenServicesCode(projectRoot string, options *SyncOptions) {
	zaplog.LOG.Debug("sync all services", zap.String("project", projectRoot), zap.Bool("mask-mode", options.MaskMode))

	protoVolume := filepath.Join(projectRoot, "api")
	serviceTypes := astkratos.ListGrpcServices(protoVolume)
	zaplog.SUG.Debugln("found gRPC services:", eroticgo.BLUE.Sprint(neatjsons.S(serviceTypes)))

	oldServiceRoot := filepath.Join(projectRoot, "internal/service")
	newServiceTemp := filepath.Join(oldServiceRoot, "tmp")
	newServiceRoot := filepath.Join(newServiceTemp, time.Now().Format("20060102150405"))

	must.Done(utils.WalkFiles(protoVolume, utils.NewSuffixPattern([]string{".proto"}), func(protoPath string, info os.FileInfo) error {
		createNewService(&createNewServiceParam{
			projectRoot:    projectRoot,
			protoPath:      protoPath,
			serviceTypes:   serviceTypes,
			oldServiceRoot: oldServiceRoot,
			newServiceRoot: newServiceRoot,
			syncOptions:    options,
		})
		return nil
	}))

	writeServiceCode(oldServiceRoot, newServiceRoot, options)

	if path := newServiceTemp; ossoftexist.IsRoot(path) {
		exist := rese.V1(utils.HasFiles(path))
		if !exist {
			must.Done(os.RemoveAll(path)) // Complete, remove redundant DIR // 完成时删除多余的 DIR
		}
	}

	zaplog.LOG.Debug("sync all done")
	eroticgo.GREEN.ShowMessage("SUCCESS")
}

// GenServicesOnce syncs service files with a single proto file
// Generates missing service and syncs existing one based on specified proto
//
// GenServicesOnce 将服务文件与单个 proto 文件同步
// 根据指定的 proto 生成缺失的服务并同步现有服务
func GenServicesOnce(projectRoot string, protoPath string, options *SyncOptions) {
	zaplog.LOG.Debug("sync single proto", zap.String("project", projectRoot), zap.String("proto", protoPath), zap.Bool("mask-mode", options.MaskMode))

	osmustexist.MustRoot(projectRoot)
	osmustexist.MustFile(protoPath)
	protoVolume := filepath.Dir(protoPath)
	serviceTypes := astkratos.ListGrpcServices(protoVolume)
	zaplog.SUG.Debugln("found gRPC services:", eroticgo.BLUE.Sprint(neatjsons.S(serviceTypes)))

	oldServiceRoot := filepath.Join(projectRoot, "internal/service")
	newServiceRoot := filepath.Join(oldServiceRoot, "tmp", time.Now().Format("20060102150405"))

	createNewService(&createNewServiceParam{
		projectRoot:    projectRoot,
		protoPath:      protoPath,
		serviceTypes:   serviceTypes,
		oldServiceRoot: oldServiceRoot,
		newServiceRoot: newServiceRoot,
		syncOptions:    options,
	})

	writeServiceCode(oldServiceRoot, newServiceRoot, options)

	zaplog.LOG.Debug("sync single done")
	eroticgo.GREEN.ShowMessage("SUCCESS")
}

// createNewServiceParam holds params needed to create and regenerate service files
// createNewServiceParam 保存创建和重新生成服务文件所需的参数
type createNewServiceParam struct {
	projectRoot    string                          // Project root path // 项目根路径
	protoPath      string                          // Proto file path // Proto 文件路径
	serviceTypes   []*astkratos.GrpcTypeDefinition // gRPC service type definitions // gRPC 服务类型定义
	oldServiceRoot string                          // Existing service DIR // 现有服务 DIR
	newServiceRoot string                          // Staging DIR to regenerate services // 重新生成服务的暂存 DIR
	syncOptions    *SyncOptions                    // Sync options // 同步选项
}

// createNewService creates and regenerates service based on proto definition
// createNewService 根据 proto 定义创建和重新生成服务
func createNewService(param *createNewServiceParam) {
	zaplog.LOG.Debug("processing proto file", zap.String("proto", param.protoPath), zap.Bool("mask-mode", param.syncOptions.MaskMode))
	anyMissing := false
	anyPresent := false

	// In mask mode, build mask type map to check service existence
	// 在 mask 模式下，构建嵌入类型映射来检查服务是否存在
	var maskMap map[string]string
	if param.syncOptions.MaskMode {
		maskMap = buildMaskTypeMap(param.oldServiceRoot)
	}

	protoCode := string(rese.V1(os.ReadFile(param.protoPath)))
	for _, serviceType := range param.serviceTypes {
		must.OK(serviceType.Name)
		zaplog.LOG.Debug("checking service", zap.String("name", serviceType.Name))

		if !strings.Contains(protoCode, fmt.Sprintf("service %s {", serviceType.Name)) {
			zaplog.LOG.Debug("service not in this proto, skip")
			continue
		}
		zaplog.LOG.Debug("service defined in proto", zap.String("name", serviceType.Name))

		// Check if service exists
		// 检查服务是否存在
		var serviceExists bool
		if param.syncOptions.MaskMode {
			// Mask mode: check via mask type (Unimplemented*Server, without package prefix)
			// Mask 模式：按嵌入类型检查（不带包前缀）
			maskTypeName := fmt.Sprintf("Unimplemented%sServer", serviceType.Name)
			_, serviceExists = maskMap[maskTypeName]
			zaplog.LOG.Debug("mask mode check", zap.String("type", maskTypeName), zap.Bool("exists", serviceExists))
		} else {
			// Default mode: check via filename
			// 默认模式：按文件名检查
			serviceFileName := strings.ToLower(serviceType.Name) + ".go"
			serviceFilePath := filepath.Join(param.projectRoot, "internal/service", serviceFileName)
			serviceExists = ossoftexist.IsFile(serviceFilePath)
		}

		if !serviceExists {
			zaplog.LOG.Debug("service not found", zap.String("name", serviceType.Name))
			anyMissing = true
		} else {
			zaplog.LOG.Debug("service exists", zap.String("name", serviceType.Name))
			anyPresent = true
		}
	}

	zaplog.LOG.Debug("check result", zap.Bool("any-missing", anyMissing), zap.Bool("any-present", anyPresent))
	if anyMissing {
		// Create new service when at least one service is missing
		// 只要有1个 service 缺失就新建服务
		zaplog.LOG.Debug("creating new service", zap.String("path", param.oldServiceRoot))
		out := rese.V1(osexec.ExecInPath(param.projectRoot, "kratos", "proto", "server", param.protoPath, "-t", param.oldServiceRoot))
		zaplog.SUG.Debugln("kratos output:", string(out))
	}

	if anyPresent {
		// Regenerate to staging DIR when at least one service exists
		// 只要有1个 service 已存在就重建到暂存 DIR 以便对比
		zaplog.LOG.Debug("regenerate to temp", zap.String("path", param.newServiceRoot))
		must.Done(os.MkdirAll(param.newServiceRoot, 0755))
		out := rese.V1(osexec.ExecInPath(param.projectRoot, "kratos", "proto", "server", param.protoPath, "-t", param.newServiceRoot))
		zaplog.SUG.Debugln("kratos output:", string(out))
	}
	zaplog.LOG.Debug("proto processing done")
}

// writeServiceCode writes synced service code back to source location
// writeServiceCode 将同步后的服务代码写回源位置
func writeServiceCode(oldServiceRoot string, newServiceRoot string, options *SyncOptions) {
	zaplog.LOG.Debug("writing service code", zap.String("old", oldServiceRoot), zap.String("new", newServiceRoot))
	if path := newServiceRoot; ossoftexist.IsRoot(path) {
		// Replace proto imports
		// 替换 proto 引用
		replaceProtoImports(path)

		// Sync service code
		// 同步服务代码
		syncServicesCode(oldServiceRoot, path, options)

		// Remove temp DIR when done
		// 完成后删除临时 DIR
		must.Done(os.RemoveAll(path))
	}
}

// replaceProtoImports fixes generated proto imports to use standard protobuf types
// replaceProtoImports 修复生成的 proto 引用以使用标准 protobuf 类型
func replaceProtoImports(newServiceRoot string) {
	zaplog.LOG.Debug("replacing proto imports", zap.String("root", newServiceRoot))
	rep := strings.NewReplacer(
		"pb.google_protobuf_StringValue", "wrapperspb.StringValue", //pb.google_protobuf_StringValue -> wrapperspb.StringValue
		"pb.google_protobuf_Empty", "emptypb.Empty", //pb.google_protobuf_Empty -> emptypb.Empty
	)

	must.Done(utils.WalkFiles(newServiceRoot, utils.NewSuffixPattern([]string{".go"}), func(path string, info os.FileInfo) error {
		srcContent := string(rese.V1(os.ReadFile(path)))
		newContent := rep.Replace(srcContent)
		if newContent != srcContent {
			newSource := syntaxgo_ast.InjectImports([]byte(newContent), []string{
				"google.golang.org/protobuf/types/known/wrapperspb",
				"google.golang.org/protobuf/types/known/emptypb",
			})
			utils.FormatAndWriteCode(path, newSource)
		}
		return nil
	}))
}

// syncServicesCode syncs old service code with new generated service code
// Adds missing methods, unexports removed methods, and sorts existing methods
//
// syncServicesCode 将旧服务代码与新生成的服务代码同步
// 添加缺失的方法、非导出已删除的方法、排序现有方法
func syncServicesCode(oldServiceRoot string, newServiceRoot string, options *SyncOptions) {
	zaplog.LOG.Debug("syncing service code", zap.String("old", oldServiceRoot), zap.String("new", newServiceRoot), zap.Bool("mask-mode", options.MaskMode))

	// In mask mode, build mask type to file path map based on old service files
	// 在 mask 模式下，根据旧服务文件构建嵌入类型到文件路径的映射
	var maskMap map[string]string
	if options.MaskMode {
		maskMap = buildMaskTypeMap(oldServiceRoot)
		zaplog.SUG.Debugln("mask type map:", neatjsons.S(maskMap))
	}

	must.Done(utils.WalkFiles(newServiceRoot, utils.NewSuffixPattern([]string{".go"}), func(path string, info os.FileInfo) error {
		zaplog.SUG.Debugln("---")

		// Parse new service file first
		// 首先解析新服务文件
		zaplog.LOG.Debug("parsing new service file", zap.String("file", info.Name()))
		vNew := parseServiceFile(path)
		zaplog.SUG.Debugln("---")

		// Find old service file path
		// 查找旧服务文件路径
		var oldFilePath string
		if options.MaskMode {
			// Mask mode: match via Unimplemented*Server type
			// Mask 模式：按嵌入的 Unimplemented*Server 类型匹配
			maskTypes := extractMaskTypes(vNew)
			if len(maskTypes) > 0 {
				maskType := maskTypes[0]
				if foundPath, ok := maskMap[maskType]; ok {
					oldFilePath = foundPath
					zaplog.LOG.Debug("mask mode matched", zap.String("type", maskType), zap.String("path", oldFilePath))
				}
			}
			if oldFilePath == "" {
				// Fallback to filename match if mask type not found
				// 如果找不到嵌入类型，回退到文件名匹配
				oldFilePath = filepath.Join(oldServiceRoot, info.Name())
				zaplog.LOG.Debug("mask mode fallback to filename", zap.String("path", oldFilePath))
			}
		} else {
			// Default mode: match via filename
			// 默认模式：按文件名匹配
			oldFilePath = filepath.Join(oldServiceRoot, info.Name())
		}

		zaplog.LOG.Debug("parsing old service file", zap.String("file", filepath.Base(oldFilePath)))
		vOld := parseServiceFile(oldFilePath)
		zaplog.SUG.Debugln("---")

		if missingCode := searchMissingMethods(vOld, vNew); len(missingCode) > 0 {
			changedCode := []byte(string(vOld.code) + "\n" + missingCode)
			utils.FormatAndWriteCode(vOld.path, changedCode)
			vOld = parseServiceFile(vOld.path)
			zaplog.LOG.Debug("added missing methods", zap.String("file", filepath.Base(vOld.path)))
		}

		if changedCode := unexportMethods(vOld, vNew); len(changedCode) > 0 {
			utils.FormatAndWriteCode(vOld.path, changedCode)
			vOld = parseServiceFile(vOld.path)
			zaplog.LOG.Debug("unexported removed methods", zap.String("file", filepath.Base(vOld.path)))
		}

		sortServiceMethods(vOld, vNew)
		return nil
	}))
}

// buildMaskTypeMap scans DIR and builds map from mask type to file path
// Skips tmp/ sub-DIR to avoid scanning temp files
// Supports multiple mask types in one file
//
// buildMaskTypeMap 扫描 DIR 并构建嵌入类型到文件路径的映射
// 跳过 tmp/ 子 DIR 以避免扫描临时文件
// 支持单个文件有多个嵌入类型
func buildMaskTypeMap(serviceRoot string) map[string]string {
	zaplog.LOG.Debug("building mask type map", zap.String("root", serviceRoot))
	maskMap := make(map[string]string)
	_ = filepath.Walk(serviceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip tmp DIR
		// 跳过 tmp DIR
		if info.IsDir() && info.Name() == "tmp" {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		svcFile := parseServiceFile(path)
		maskTypes := extractMaskTypes(svcFile)
		for _, maskType := range maskTypes {
			maskMap[maskType] = path
			zaplog.LOG.Debug("found mask type", zap.String("type", maskType), zap.String("file", info.Name()))
		}
		return nil
	})
	return maskMap
}

// extractMaskTypes extracts Unimplemented*Server mask types from ServiceFile
// Returns type names without package prefix (e.g., "UnimplementedGreeterServer")
//
// extractMaskTypes 从 ServiceFile 中提取所有 Unimplemented*Server 嵌入类型
// 只返回类型名，不包含包前缀（例如 "UnimplementedGreeterServer"）
func extractMaskTypes(svcFile *ServiceFile) []string {
	var types []string
	for _, serviceStruct := range svcFile.serviceStructMap {
		if serviceStruct.structType == nil || serviceStruct.structType.Fields == nil {
			continue
		}
		for _, field := range serviceStruct.structType.Fields.List {
			// Check embedded field (mask type)
			// 检查嵌入字段（mask 类型）
			if len(field.Names) == 0 {
				typeName := getTypeName(svcFile.code, field.Type)
				if strings.Contains(typeName, "Unimplemented") && strings.HasSuffix(typeName, "Server") {
					// Remove package prefix (e.g., "v1.UnimplementedGreeterServer" -> "UnimplementedGreeterServer")
					// 移除包前缀（例如 "v1.UnimplementedGreeterServer" -> "UnimplementedGreeterServer"）
					if idx := strings.LastIndex(typeName, "."); idx != -1 {
						typeName = typeName[idx+1:]
					}
					types = append(types, typeName)
				}
			}
		}
	}
	return types
}

// getTypeName extracts type name string from ast.Expr
// getTypeName 从 ast.Expr 中提取类型名称字符串
func getTypeName(code []byte, expr ast.Expr) string {
	return strings.TrimSpace(syntaxgo_astnode.GetText(code, expr))
}

// searchMissingMethods detects missing methods when proto adds new functions
// In mask mode, match structs via mask type and swap the method's struct name
//
// searchMissingMethods 检测 proto 增加函数时服务代码中缺失的方法
// 在 mask 模式下，按嵌入类型匹配 struct 并替换方法的 struct 名
func searchMissingMethods(oldFile *ServiceFile, newFile *ServiceFile) string {
	ptx := printgo.NewPTX()

	// Build mask type to struct name map based on old file
	// 根据旧文件构建嵌入类型到 struct 名的映射
	oldMaskToStruct := buildStructMaskMap(oldFile)
	newMaskToStruct := buildStructMaskMap(newFile)

	for structName, newServiceStruct := range newFile.serviceStructMap {
		zaplog.SUG.Debugln("---")
		zaplog.LOG.Debug("checking struct", zap.String("name", structName))

		// Find matching struct via name first
		// 首先按名字查找匹配的 struct
		serviceStruct, ok := oldFile.serviceStructMap[structName]
		oldStructName := structName

		// If not found via name, find via mask type
		// 如果按名字找不到，按嵌入类型查找
		if !ok {
			newMaskType := newMaskToStruct[structName]
			if newMaskType != "" {
				for oldName, oldMask := range oldMaskToStruct {
					if oldMask == newMaskType {
						serviceStruct = oldFile.serviceStructMap[oldName]
						oldStructName = oldName
						zaplog.LOG.Debug("mask mode struct match", zap.String("new", structName), zap.String("old", oldStructName), zap.String("via", newMaskType))
						break
					}
				}
			}
		}

		if serviceStruct == nil {
			ptx.Println("type", structName, newFile.GetNode(newServiceStruct.structType))
			for _, method := range newServiceStruct.methods {
				ptx.Println(newFile.GetNode(method))
			}
			continue
		}

		methods := newServiceStruct.methods
		for _, method := range methods {
			oldMethod, ok := serviceStruct.methodsMap[method.Name.Name]
			if !ok {
				zaplog.LOG.Debug("to add", zap.String("method", method.Name.Name))
				// Swap struct name in method if names mismatch
				// 如果 struct 名不同，替换方法中的 struct 名
				methodCode := newFile.GetNode(method)
				if structName != oldStructName {
					methodCode = strings.Replace(methodCode, "*"+structName, "*"+oldStructName, 1)
				}
				ptx.Println(methodCode)
				continue
			}
			zaplog.LOG.Debug("exists", zap.String("method", oldMethod.Name.Name))
		}
	}
	missingCode := strings.TrimSpace(ptx.String())
	return missingCode
}

// buildStructMaskMap builds struct name to mask type map
// buildStructMaskMap 构建 struct 名到嵌入类型的映射
func buildStructMaskMap(svcFile *ServiceFile) map[string]string {
	result := make(map[string]string)
	for structName, serviceStruct := range svcFile.serviceStructMap {
		if serviceStruct.structType == nil || serviceStruct.structType.Fields == nil {
			continue
		}
		for _, field := range serviceStruct.structType.Fields.List {
			if len(field.Names) == 0 {
				typeName := getTypeName(svcFile.code, field.Type)
				if strings.Contains(typeName, "Unimplemented") && strings.HasSuffix(typeName, "Server") {
					if idx := strings.LastIndex(typeName, "."); idx != -1 {
						typeName = typeName[idx+1:]
					}
					result[structName] = typeName
					break
				}
			}
		}
	}
	return result
}

// unexportMethods converts deleted proto functions to unexported
// In mask mode, match structs via mask type
//
// unexportMethods 当 proto 删除函数时自动转换为非导出
// 在 mask 模式下，按嵌入类型匹配 struct
func unexportMethods(oldFile *ServiceFile, newFile *ServiceFile) []byte {
	var removedMethods []*ast.FuncDecl

	// Build mask type to struct name map
	// 构建嵌入类型到 struct 名的映射
	oldMaskToStruct := buildStructMaskMap(oldFile)
	newMaskToStruct := buildStructMaskMap(newFile)

	// Build map: mask type -> new struct
	// 构建映射：嵌入类型 -> 新 struct
	maskToNewStruct := make(map[string]*ServiceStruct)
	for structName, maskType := range newMaskToStruct {
		maskToNewStruct[maskType] = newFile.serviceStructMap[structName]
	}

	for structName := range oldFile.serviceStructMap {
		zaplog.SUG.Debugln("---")
		zaplog.LOG.Debug("check removed methods", zap.String("struct", structName))
		oldMethods := oldFile.serviceStructMap[structName].methods

		// Find matching struct via name first
		// 首先按名字查找匹配的 struct
		serviceStruct, ok := newFile.serviceStructMap[structName]

		// If not found via name, find via mask type
		// 如果按名字找不到，按嵌入类型查找
		oldMaskType := oldMaskToStruct[structName]
		if !ok {
			if oldMaskType != "" {
				serviceStruct = maskToNewStruct[oldMaskType]
			}
		}

		// If struct not found in new file, check if it should be skipped
		// 如果在新文件中找不到 struct，检查是否应该跳过
		if serviceStruct == nil {
			// In multi-service scenario: skip if old struct's mask type is not in new file
			// 多服务场景：如果旧 struct 的 mask type 不在新文件中，跳过（不属于当前处理的服务）
			if oldMaskType != "" {
				if _, exists := maskToNewStruct[oldMaskType]; !exists {
					// This struct belongs to a different service, skip it
					// 这个 struct 属于不同的服务，跳过
					continue
				}
			}
			// Struct's mask type is in new file but struct not found -> each method should be unexported
			// Struct 的 mask type 在新文件中但找不到 struct -> 所有方法都应变为非导出
			removedMethods = append(removedMethods, oldMethods...)
			continue
		}

		for _, method := range oldMethods {
			newMethod, ok := serviceStruct.methodsMap[method.Name.Name]
			if !ok {
				zaplog.LOG.Debug("to unexport", zap.String("method", method.Name.Name))
				removedMethods = append(removedMethods, method) // Method not in new service should be unexported // 新服务里没有此方法则应非导出
				continue
			}
			zaplog.LOG.Debug("retained", zap.String("method", newMethod.Name.Name))
		}
	}
	if len(removedMethods) == 0 {
		return []byte{}
	}

	source := utils.CopyBytes(oldFile.code)
	changed := false // Track if any change made, skip write if unchanged // 跟踪是否有改动，无改动则跳过写入
	for _, method := range removedMethods {
		name := method.Name.Name
		zaplog.LOG.Debug("convert to unexported", zap.String("name", name))
		if utils.IsFirstCharUpper(name) {
			newName := []byte(utils.LowerFirstChar(name))
			oldName := syntaxgo_astnode.GetCode(source, method.Name)
			must.Same(len(newName), len(oldName))
			copy(oldName, newName) // Same length allows in-place update // 长度相同可直接原地替换
			changed = true
		}
	}
	if changed {
		return source
	}
	return []byte{}
}

// sortServiceMethods sorts methods to match proto definition sequence
// Treats each method with its post code as a block, sorts via method name
// In mask mode, match structs via mask type
//
// sortServiceMethods 按 proto 定义顺序排序服务方法
// 把每个方法及其后续代码作为代码块，按方法名排序
// 在 mask 模式下，按嵌入类型匹配 struct
func sortServiceMethods(oldFile *ServiceFile, newFile *ServiceFile) {
	// Build mask type to struct name map
	// 构建嵌入类型到 struct 名的映射
	oldMaskToStruct := buildStructMaskMap(oldFile)
	newMaskToStruct := buildStructMaskMap(newFile)

	for structName, newServiceStruct := range newFile.serviceStructMap {
		zaplog.SUG.Debugln("---")
		zaplog.LOG.Debug("sort methods", zap.String("struct", structName))

		// Find matching struct via name first, then via mask type
		// 首先按名字查找，然后按嵌入类型查找
		oldServiceStruct, ok := oldFile.serviceStructMap[structName]
		if !ok {
			newMaskType := newMaskToStruct[structName]
			if newMaskType != "" {
				for oldName, oldMask := range oldMaskToStruct {
					if oldMask == newMaskType {
						oldServiceStruct = oldFile.serviceStructMap[oldName]
						zaplog.LOG.Debug("mask mode sort match", zap.String("new", structName), zap.String("old", oldName), zap.String("via", newMaskType))
						break
					}
				}
			}
		}
		if oldServiceStruct == nil {
			continue
		}

		// Collect valid old methods to sort based on new file index
		// 收集有效的旧方法列表，根据新文件索引排序
		var methods []*ast.FuncDecl
		for _, method := range oldServiceStruct.methods {
			_, ok := newServiceStruct.methodsMap[method.Name.Name]
			if ok {
				methods = append(methods, method)
			}
		}
		for _, method := range methods {
			zaplog.LOG.Debug("to sort", zap.String("method", method.Name.Name))
		}
		zaplog.SUG.Debugln("---")

		if len(methods) == 0 {
			// Use return instead of continue: single struct file is the common case
			// 使用 return 而非 continue：单 struct 文件是常见情况
			return // No methods to sort // 没有方法需要排序
		}

		ptx := printgo.NewPTX()

		// Head block: code before first method (package, imports, struct, init)
		// 头部块：第一个方法前的代码（包名、引用、结构体、初始化）
		headNode := syntaxgo_astnode.NewNode(1, methods[0].Pos())
		if methods[0].Doc != nil {
			checkDocPos(methods[0])
			headNode.SetEnd(methods[0].Doc.Pos()) // End at doc comment if exists // 有注释则截取到注释前
		}
		ptx.Println(syntaxgo_astnode.GetText(oldFile.code, headNode))

		type MethodBlock struct {
			Node *syntaxgo_astnode.Node
			Name string
		}

		// Method blocks: each method with its code up to next method
		// E.g., new has A/B and old has A/F/B, then A+F is one block, B is one block
		//
		// 方法代码块：每个方法及其代码直到下个方法
		// 例如：新服务有 A/B，旧服务有 A/F/B，则 A+F 是一块，B 是一块
		methodBlocks := make([]*MethodBlock, 0, len(methods))
		for idx, method := range methods {
			node := syntaxgo_astnode.NewNode(method.Pos(), token.Pos(1+len(oldFile.code)))
			if method.Doc != nil {
				checkDocPos(method)
				node.SetPos(method.Doc.Pos()) // Start from doc comment if exists // 有注释则从注释位置开始
			}

			if nextIdx := idx + 1; nextIdx < len(methods) {
				nextMethod := methods[nextIdx]
				node.SetEnd(nextMethod.Pos()) // Next method start is current method end // 下个方法的起始位置即当前方法的结束位置
				if nextMethod.Doc != nil {
					checkDocPos(nextMethod)
					node.SetEnd(nextMethod.Doc.Pos()) // Or next doc comment start // 或下个方法注释的起始位置
				}
			}

			zaplog.SUG.Debugln(eroticgo.BLUE.Sprint(syntaxgo_astnode.GetText(oldFile.code, node)))

			methodBlocks = append(methodBlocks, &MethodBlock{
				Node: node,
				Name: method.Name.Name,
			})
		}
		must.Same(len(methods), len(methodBlocks))

		compareLess := func(i, j int) bool {
			a := methodBlocks[i]
			b := methodBlocks[j]
			idxA := newServiceStruct.methodsIdx[a.Name] // Index in new file (proto order) // 在新文件中的序号（proto 顺序）
			idxB := newServiceStruct.methodsIdx[b.Name] // Index in new file // 在新文件中的序号
			return idxA < idxB
		}
		if sort.SliceIsSorted(methodBlocks, compareLess) {
			return // Skip if sorted // 已排序则跳过
		}

		sortx.SortByIndex(methodBlocks, compareLess)

		// Concatenate sorted blocks
		// 拼接排序后的代码块
		for _, methodBlock := range methodBlocks {
			ptx.Println(syntaxgo_astnode.GetText(oldFile.code, methodBlock.Node))
		}

		// Format and write back
		// 格式化并写回文件
		utils.FormatAndWriteCode(oldFile.path, ptx.Bytes())
	}
}

// checkDocPos validates doc comment position is before function declaration
// checkDocPos 验证文档注释位置在函数声明之前
func checkDocPos(method *ast.FuncDecl) {
	if method.Doc != nil {
		must.True(method.Doc.Pos() < method.Pos())
		must.True(method.Doc.End() < method.Pos())
	}
}

// parseServiceFile parses Go service file and extracts struct and method info
// parseServiceFile 解析 Go 服务文件并提取结构体和方法信息
func parseServiceFile(path string) *ServiceFile {
	code := rese.V1(os.ReadFile(path))
	astBundle := rese.P1(syntaxgo_ast.NewAstBundleV1(code))
	astFile, _ := astBundle.GetBundle()
	structTypes := syntaxgo_search.MapStructTypesByName(astFile)

	serviceStructMap := make(map[string]*ServiceStruct, len(structTypes))
	for structName, structType := range structTypes {
		zaplog.SUG.Debugln("---")
		methods := syntaxgo_search.FindFunctionsByReceiverName(astFile, structName, true)
		zaplog.LOG.Debug("found struct", zap.String("name", structName), zap.Int("methods", len(methods)))
		for _, method := range methods {
			zaplog.LOG.Debug("  method", zap.String("name", method.Name.Name))
		}
		zaplog.SUG.Debugln("---")

		methodsMap := make(map[string]*ast.FuncDecl, len(methods))
		methodsIdx := make(map[string]int, len(methods))
		for idx, method := range methods {
			methodsMap[method.Name.Name] = method
			methodsIdx[method.Name.Name] = idx
		}

		serviceStructMap[structName] = &ServiceStruct{
			structType: structType,
			methods:    methods,
			methodsMap: methodsMap,
			methodsIdx: methodsIdx,
		}
	}
	return &ServiceFile{
		path:             path,
		code:             code,
		serviceStructMap: serviceStructMap,
	}
}

// ServiceFile represents a parsed Go service file with its structs and methods
// ServiceFile 表示已解析的 Go 服务文件及其结构体和方法
type ServiceFile struct {
	path             string                    // File path // 文件路径
	code             []byte                    // Source code content // 源代码内容
	serviceStructMap map[string]*ServiceStruct // Struct name to ServiceStruct map // 结构体名到 ServiceStruct 的映射
}

// GetNode extracts source code text of an AST node
// GetNode 提取 AST 节点的源代码文本
func (sf *ServiceFile) GetNode(astNode ast.Node) string {
	return syntaxgo_astnode.GetText(sf.code, astNode)
}

// ServiceStruct represents a service struct with its methods
// ServiceStruct 表示服务结构体及其方法
type ServiceStruct struct {
	structType *ast.StructType          // AST struct type // AST 结构体类型
	methods    []*ast.FuncDecl          // Methods in declaration sequence // 按声明顺序排列的方法
	methodsMap map[string]*ast.FuncDecl // Method name to FuncDecl map // 方法名到 FuncDecl 的映射
	methodsIdx map[string]int           // Method name to index map // 方法名到索引的映射
}
