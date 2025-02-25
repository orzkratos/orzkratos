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
	"github.com/yyle88/sortslice"
	"github.com/yyle88/syntaxgo/syntaxgo_ast"
	"github.com/yyle88/syntaxgo/syntaxgo_astnode"
	"github.com/yyle88/syntaxgo/syntaxgo_search"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

func GenServicesCode(projectRoot string) {
	protoVolume := filepath.Join(projectRoot, "api")
	defineServiceTypes := astkratos.ListGrpcServices(protoVolume)
	if len(defineServiceTypes) > 0 {
		zaplog.SUG.Debugln(neatjsons.S(defineServiceTypes))
	} else {
		zaplog.SUG.Debugln("maybe no service in protos")
	}

	oldServiceRoot := filepath.Join(projectRoot, "internal/service")
	newServiceTemp := filepath.Join(oldServiceRoot, "tmp")
	newServiceRoot := filepath.Join(newServiceTemp, time.Now().Format("20060102150405"))

	must.Done(astkratos.WalkFiles(protoVolume, astkratos.NewSuffixMatcher([]string{".proto"}), func(path string, info os.FileInfo) error {
		zaplog.SUG.Debugln("proto:", path)
		createServicesOnce(path, defineServiceTypes, projectRoot, oldServiceRoot, newServiceRoot)
		return nil
	}))

	if path := newServiceRoot; ossoftexist.IsRoot(path) {
		zaplog.SUG.Debugln(path)

		replaceProtoTypes(path)

		zaplog.SUG.Debugln(path)

		rewriteServicesCode(oldServiceRoot, path) //根据接口定义重新调整service代码内容，把缺少的补全出来，把多余的变成小写的

		must.Done(os.RemoveAll(path)) //结束以后要删除这个多余的目录
	}

	if path := newServiceTemp; ossoftexist.IsRoot(newServiceTemp) {
		if rese.V1(utils.CntFileNum(path)) == 0 {
			must.Done(os.RemoveAll(path)) //结束以后要删除这个多余的目录
		}
	}
}

func GenServicesOnce(projectRoot string, protoPath string) {
	osmustexist.MustRoot(projectRoot)
	osmustexist.MustFile(protoPath)
	protoVolume := filepath.Dir(protoPath)
	defineTypes := astkratos.ListGrpcServices(protoVolume)
	zaplog.SUG.Debugln(neatjsons.S(defineTypes))

	oldServiceRoot := filepath.Join(projectRoot, "internal/service")
	newServiceRoot := filepath.Join(oldServiceRoot, "tmp", time.Now().Format("20060102150405"))

	createServicesOnce(protoPath, defineTypes, projectRoot, oldServiceRoot, newServiceRoot)

	if ossoftexist.IsRoot(newServiceRoot) {
		zaplog.SUG.Debugln(newServiceRoot)

		replaceProtoTypes(newServiceRoot)

		zaplog.SUG.Debugln(newServiceRoot)

		rewriteServicesCode(oldServiceRoot, newServiceRoot)
		//结束以后要删除这个多余的目录
		must.Done(os.RemoveAll(newServiceRoot))
	}
}

func createServicesOnce(protoPath string, defineTypes []*astkratos.GrpcTypeDefinition, projectRoot string, oldServiceRoot string, newServiceRoot string) {
	zaplog.LOG.Debug("once")
	var miss = false
	var meet = false

	protoCode := string(rese.V1(os.ReadFile(protoPath)))
	for _, c := range defineTypes {
		must.OK(c.Name)
		defineService := fmt.Sprintf("service %s {", c.Name)
		zaplog.SUG.Debugln(defineService)
		if strings.Contains(protoCode, defineService) {
			if !ossoftexist.IsFile(filepath.Join(projectRoot, "internal/service", strings.ToLower(c.Name)+".go")) {
				zaplog.SUG.Debugln("meet:", defineService, "miss:", c.Name)
				miss = true
			} else {
				zaplog.SUG.Debugln("meet:", defineService, "have:", c.Name)
				meet = true
			}
		}
	}
	//只要有1个 service 是 miss 的就新建服务
	if miss {
		zaplog.LOG.Debug("miss")
		startTime := time.Now()

		out := rese.V1(osexec.ExecInPath(projectRoot, "kratos", "proto", "server", protoPath, "-t", oldServiceRoot))
		zaplog.SUG.Debugln(string(out))
		zaplog.LOG.Debug("miss-done", zap.Duration("duration", time.Since(startTime)))
	}
	//只要有1个 service 是 meet 的就重建服务和检查服务，看看是不是需要更新代码
	//注意，这块逻辑要放在新建逻辑的后面，以确保重写时服务已经存在
	if meet {
		zaplog.LOG.Debug("meet")
		startTime := time.Now()

		must.Done(os.MkdirAll(newServiceRoot, 0755))
		out := rese.V1(osexec.ExecInPath(projectRoot, "kratos", "proto", "server", protoPath, "-t", newServiceRoot))
		zaplog.SUG.Debugln(string(out))
		zaplog.LOG.Debug("meet-done", zap.Duration("duration", time.Since(startTime)))
	}
}

func replaceProtoTypes(newServiceRoot string) {
	rep := strings.NewReplacer(
		"pb.google_protobuf_StringValue", "wrapperspb.StringValue", //pb.google_protobuf_StringValue -> wrapperspb.StringValue
		"pb.google_protobuf_Empty", "emptypb.Empty", //pb.google_protobuf_Empty -> emptypb.Empty
	)

	must.Done(astkratos.WalkFiles(newServiceRoot, astkratos.NewSuffixMatcher([]string{".go"}), func(path string, info os.FileInfo) error {
		srcContent := string(rese.V1(os.ReadFile(path)))
		newContent := rep.Replace(srcContent)
		if newContent != srcContent {
			newSource := syntaxgo_ast.InjectImports([]byte(newContent), []string{
				"google.golang.org/protobuf/types/known/wrapperspb",
				"google.golang.org/protobuf/types/known/emptypb",
			})
			utils.WriteFormatBytes(newSource, path)
		}
		return nil
	}))
}

func rewriteServicesCode(oldServiceRoot string, newServiceRoot string) {
	zaplog.SUG.Debugln(oldServiceRoot)
	zaplog.SUG.Debugln(newServiceRoot)

	must.Done(astkratos.WalkFiles(newServiceRoot, astkratos.NewSuffixMatcher([]string{".go"}), func(path string, info os.FileInfo) error {
		vOld := parseServiceFile(filepath.Join(oldServiceRoot, info.Name()))
		vNew := parseServiceFile(path)

		if missingCode := searchMissingMethods(vOld, vNew); len(missingCode) > 0 {
			changedCode := []byte((string(vOld.code) + "\n" + missingCode))

			utils.WriteFormatBytes(changedCode, vOld.path)
			vOld = parseServiceFile(vOld.path)
		}

		if changedCode := notExportUselessMethods(vOld, vNew); len(changedCode) > 0 {
			utils.WriteFormatBytes(changedCode, vOld.path)
			vOld = parseServiceFile(vOld.path)
		}

		sortServiceMethods(vOld, vNew)
		return nil
	}))
}

// 当开发者在 proto 里增加函数时，这个函数能够识别服务代码中缺失的方法代码
func searchMissingMethods(old *ServiceFile, new *ServiceFile) string {
	ptx := printgo.NewPTX()
	for structName, temp := range new.serviceStructsMap {
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(new.GetNode(temp.structType))
		zaplog.SUG.Debugln("-")

		serviceStruct, ok := old.serviceStructsMap[structName]
		if !ok {
			ptx.Println(new.GetStructNode(structName))
			methods := new.serviceStructsMap[structName].methods
			for _, method := range methods {
				ptx.Println(new.GetNode(method))
			}
			continue
		}
		methods := new.serviceStructsMap[structName].methods
		for _, method := range methods {
			oldMethod, ok := serviceStruct.methodsMap[method.Name.Name]
			if !ok {
				zaplog.SUG.Debugln("missing", method.Name.Name)
				ptx.Println(new.GetNode(method))
				continue
			}
			zaplog.SUG.Debugln("existing", oldMethod.Name.Name)
		}
	}
	missingCode := strings.TrimSpace(ptx.String())
	return missingCode
}

// 当开发者删除 proto 中某个函数时，这个函数能自动把被删除的函数转换为非导出的
func notExportUselessMethods(old *ServiceFile, new *ServiceFile) []byte {
	var uselessMethods []*ast.FuncDecl
	for structName, temp := range old.serviceStructsMap {
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(old.GetNode(temp.structType))
		zaplog.SUG.Debugln("-")
		oldMethods := old.serviceStructsMap[structName].methods

		//假如新服务里没有某个类型，则这个类型下的所有方法都应该是非导出的
		serviceStruct, ok := new.serviceStructsMap[structName]
		if !ok {
			uselessMethods = append(uselessMethods, oldMethods...)
			continue
		}

		for _, method := range oldMethods {
			newMethod, ok := serviceStruct.methodsMap[method.Name.Name]
			if !ok {
				zaplog.SUG.Debugln("useless", method.Name.Name)
				uselessMethods = append(uselessMethods, method) //假如新服务里没有这个方法则它也该是非导出的
				continue
			}
			zaplog.SUG.Debugln("matching", newMethod.Name.Name)
		}
	}
	if len(uselessMethods) == 0 {
		return []byte{}
	}

	var source = utils.Clone(old.code)
	var change = false //结果是否改变，假如没有替换的就不用写回文件，能提升性能
	for _, method := range uselessMethods {
		name := method.Name.Name
		zaplog.SUG.Debugln("useless", name)
		if utils.C0IsUpper(name) {
			newName := []byte(utils.CvtC0Lower(name))
			oldName := syntaxgo_astnode.GetCode(source, method.Name)
			must.Same(len(newName), len(oldName))
			copy(oldName, newName) //这里由于长度相同，因此可以直接复制在相同的位置，就算是大功告成啦
			change = true
		}
	}
	if change {
		return source
	}
	return []byte{}
}

// 这个函数的目的是：当proto文件里调整了函数的顺序，则在service代码里也自动调整函数顺序，这个符合开发者的预期
// 思路基本是这样的：首先拿到所有的函数，其次过滤出有用的函数
// 把有用的函数以及它下面的无用部分看做整块代码，直到下个有用的函数为止的范围都是代码块
// 这个代码块，有1个函数名称，按照这个名称排序即可
func sortServiceMethods(old *ServiceFile, new *ServiceFile) {
	for structName, newServiceStruct := range new.serviceStructsMap {
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(new.GetNode(newServiceStruct.structType))
		zaplog.SUG.Debugln("-")

		//找到旧服务的结构体，由于新服务是个壳，旧服务是实现了业务的代码，因此旧服务的同名结构体一定存在
		oldStructParam, ok := old.serviceStructsMap[structName]
		must.True(ok)

		//这里是旧服务的方法列表，需要根据新文件的索引，把旧文件重新编排，因此首先是收集有效的旧方法列表
		var methods []*ast.FuncDecl
		for _, fun := range oldStructParam.methods {
			_, ok := newServiceStruct.methodsMap[fun.Name.Name]
			if ok {
				methods = append(methods, fun)
			}
		}
		if len(methods) == 0 {
			return //假如啥都没有，也就不用排序啦，其实也可更严格些比如 len < 2 时，也是不用排序的，但不要这么做以免影响调试效果
		}

		ptx := printgo.NewPTX()

		//首先是第一个方法前的代码块，即头部块，比如包名+引用+定义结构体+初始化逻辑等等
		headNode := syntaxgo_astnode.NewNode(1, methods[0].Pos())
		if methods[0].Doc != nil {
			checkDocPos(methods[0])
			headNode.SetEnd(methods[0].Doc.Pos()) //假如方法有注释，就截取到注释以前的部分
		}
		ptx.Println(syntaxgo_astnode.GetText(old.code, headNode))

		type MethodBlock struct {
			Node *syntaxgo_astnode.Node
			Name string
		}

		//接下来是每个方法的代码块，这个方法块可能不止单个方法的代码，还有可能会包含其它自定义的代码
		//由于只是筛选出新服务中有的方法，因此相邻两个方法中间还有其它代码，让代码归属于前面的方法块
		//例如，新服务只有 A 和 B 两个方法，而旧服务有 A F B 三个方法，则 A和F 是一块， B 是一块
		var methodBlocks = make([]*MethodBlock, 0, len(methods))
		for idx, method := range methods {
			node := syntaxgo_astnode.NewNode(method.Pos(), token.Pos(1+len(old.code)))
			if method.Doc != nil {
				checkDocPos(method)
				node.SetPos(method.Doc.Pos()) //假如有注释，则代码块的起始位置应该是注释的位置
			}

			if edx := idx + 1; edx < len(methods) {
				nextFn := methods[edx]
				node.SetEnd(nextFn.Pos()) //第二个函数的起始位置，就是第一个函数的结束位置
				if nextFn.Doc != nil {
					checkDocPos(nextFn)
					node.SetEnd(nextFn.Doc.Pos()) //第二个函数的注释起始位置，就是第一个函数的结束位置
				}
			}

			zaplog.SUG.Debugln(eroticgo.BLUE.Sprint(syntaxgo_astnode.GetText(old.code, node)))

			methodBlocks = append(methodBlocks, &MethodBlock{
				Node: node,
				Name: method.Name.Name,
			})
		}
		must.Same(len(methods), len(methodBlocks))

		compareLess := func(i, j int) bool {
			a := methodBlocks[i]
			b := methodBlocks[j]
			idxA := newServiceStruct.methodsIdx[a.Name] //在新文件中的序号-假如用户调整proto的函数顺序，则新文件序号会自动调整
			idxB := newServiceStruct.methodsIdx[b.Name] //在新文件中的序号
			return idxA < idxB
		}
		if sort.SliceIsSorted(methodBlocks, compareLess) {
			return //假如已经是有序的，就没必要再排序，这样也能提升性能
		}

		sortslice.SortByIndex(methodBlocks, compareLess)

		//把排序后的代码拼接起来
		for _, methodBlock := range methodBlocks {
			ptx.Println(syntaxgo_astnode.GetText(old.code, methodBlock.Node))
		}

		//把代码格式化再写回文件
		utils.WriteFormatBytes(ptx.Bytes(), old.path)
	}
}

func checkDocPos(fun *ast.FuncDecl) {
	if fun.Doc != nil {
		must.True(fun.Doc.Pos() < fun.Pos())
		must.True(fun.Doc.End() < fun.Pos())
	}
}

func parseServiceFile(path string) *ServiceFile {
	code := rese.V1(os.ReadFile(path))
	astBundle := rese.P1(syntaxgo_ast.NewAstBundleV1(code))
	astFile, _ := astBundle.GetBundle()

	structTypes := syntaxgo_search.MapStructTypesByName(astFile)
	zaplog.SUG.Debugln(len(structTypes))

	var serviceStructsMap = make(map[string]*ServiceStruct, len(structTypes))
	for structName, structType := range structTypes {
		zaplog.SUG.Debugln(structName)
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(syntaxgo_astnode.GetText(code, structType))
		zaplog.SUG.Debugln("-")

		methods := syntaxgo_search.FindFunctionsByReceiverName(astFile, structName, true)
		zaplog.SUG.Debugln("-")
		for _, fun := range methods {
			zaplog.SUG.Debugln(fun.Name.Name)
		}
		zaplog.SUG.Debugln("-")

		var methodsMap = make(map[string]*ast.FuncDecl, len(methods))
		var methodsIdx = make(map[string]int, len(methods))
		for idx, fun := range methods {
			methodsMap[fun.Name.Name] = fun
			methodsIdx[fun.Name.Name] = idx
		}

		serviceStructsMap[structName] = &ServiceStruct{
			structType: structType,
			methods:    methods,
			methodsMap: methodsMap,
			methodsIdx: methodsIdx,
		}
	}
	return &ServiceFile{
		path:              path,
		code:              code,
		serviceStructsMap: serviceStructsMap,
	}
}

type ServiceFile struct {
	path              string
	code              []byte
	serviceStructsMap map[string]*ServiceStruct
}

func (x *ServiceFile) GetNode(astNode ast.Node) string {
	return syntaxgo_astnode.GetText(x.code, astNode)
}

func (x *ServiceFile) GetStructNode(structName string) string {
	return syntaxgo_astnode.GetText(x.code, x.serviceStructsMap[structName].structType)
}

type ServiceStruct struct {
	structType *ast.StructType
	methods    []*ast.FuncDecl
	methodsMap map[string]*ast.FuncDecl
	methodsIdx map[string]int
}
