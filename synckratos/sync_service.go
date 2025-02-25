package synckratos

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/barweiss/go-tuple"
	"github.com/orzkratos/astkratos"
	"github.com/orzkratos/orzkratos/internal/utils"
	"github.com/yyle88/eroticgo"
	"github.com/yyle88/formatgo"
	"github.com/yyle88/must"
	"github.com/yyle88/must/mustslice"
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
		genServicesOnce(path, defineServiceTypes, projectRoot, oldServiceRoot, newServiceRoot)
		return nil
	}))

	if path := newServiceRoot; ossoftexist.IsRoot(path) {
		zaplog.SUG.Debugln(path)

		replaceSomeProtoGoTypes(path)

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

	genServicesOnce(protoPath, defineTypes, projectRoot, oldServiceRoot, newServiceRoot)

	if ossoftexist.IsRoot(newServiceRoot) {
		zaplog.SUG.Debugln(newServiceRoot)

		replaceSomeProtoGoTypes(newServiceRoot)

		zaplog.SUG.Debugln(newServiceRoot)

		rewriteServicesCode(oldServiceRoot, newServiceRoot)
		//结束以后要删除这个多余的目录
		must.Done(os.RemoveAll(newServiceRoot))
	}
}

func genServicesOnce(protoPath string, defineTypes []*astkratos.GrpcTypeDefinition, projectRoot string, oldServiceRoot string, newServiceRoot string) {
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

func replaceSomeProtoGoTypes(newServiceRoot string) {
	rep := strings.NewReplacer(
		"pb.google_protobuf_StringValue", "wrapperspb.StringValue", //pb.google_protobuf_StringValue -> wrapperspb.StringValue
		"pb.google_protobuf_Empty", "emptypb.Empty", //pb.google_protobuf_Empty -> emptypb.Empty
	)

	must.Done(astkratos.WalkFiles(newServiceRoot, astkratos.NewSuffixMatcher([]string{".go"}), func(path string, info os.FileInfo) error {
		content := string(rese.V1(os.ReadFile(path)))
		content = rep.Replace(content)

		newSource := syntaxgo_ast.InjectImports([]byte(content), []string{
			"google.golang.org/protobuf/types/known/wrapperspb",
			"google.golang.org/protobuf/types/known/emptypb",
		})

		//zaplog.SUG.Debugln(string(newSource))

		newSource, _ = formatgo.FormatBytes(newSource)
		must.Done(os.WriteFile(path, newSource, 0644))
		return nil
	}))
}

func rewriteServicesCode(oldServiceRoot string, newServiceRoot string) {
	zaplog.SUG.Debugln(oldServiceRoot)
	zaplog.SUG.Debugln(newServiceRoot)

	must.Done(astkratos.WalkFiles(newServiceRoot, astkratos.NewSuffixMatcher([]string{".go"}), func(path string, info os.FileInfo) error {
		oldX := collectAstParam(filepath.Join(oldServiceRoot, info.Name()))
		newX := collectAstParam(path)

		if missCode := calcMissCode(oldX, newX); len(missCode) > 0 {
			code := []byte((string(oldX.content) + "\n" + missCode))

			newCode, _ := formatgo.FormatBytes(code)
			must.Done(os.WriteFile(oldX.path, newCode, 0644))
			oldX = collectAstParam(oldX.path)
		}

		if code, change := cleanMoreCode(oldX, newX); change {
			newCode, _ := formatgo.FormatBytes(code)
			must.Done(os.WriteFile(oldX.path, newCode, 0644))
			oldX = collectAstParam(oldX.path)
		}

		sortRecvFuncs(oldX, newX)
		return nil
	}))
}

func calcMissCode(oldX *astParam, newX *astParam) string {
	ptx2miss := printgo.NewPTX()
	for structName, newStructParam := range newX.mapStructParams {
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(newX.GetNode(newStructParam.structType))
		zaplog.SUG.Debugln("-")

		if ot, ok := oldX.mapStructParams[structName]; !ok {
			ptx2miss.Println(newX.GetStructNode(structName))
			recvFuncs := newX.mapStructParams[structName].recvFuncs
			for _, astFunc := range recvFuncs {
				ptx2miss.Println(newX.GetNode(astFunc))
			}
		} else {
			astFuncs := newX.mapStructParams[structName].recvFuncs
			for _, astFunc := range astFuncs {
				if oldFunc, ok := ot.recvFuncsMap[astFunc.Name.Name]; !ok {
					zaplog.SUG.Debugln("miss", astFunc.Name.Name)
					ptx2miss.Println(newX.GetNode(astFunc))
				} else {
					zaplog.SUG.Debugln("meet", oldFunc.Name.Name)
				}
			}
		}
	}
	missCode := strings.TrimSpace(ptx2miss.String())
	return missCode
}

func cleanMoreCode(oldX *astParam, newX *astParam) ([]byte, bool) {
	var moreFuncs []*ast.FuncDecl
	for structName, newStructParam := range oldX.mapStructParams {
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(oldX.GetNode(newStructParam.structType))
		zaplog.SUG.Debugln("-")
		recvFuncs := oldX.mapStructParams[structName].recvFuncs

		if nt, ok := newX.mapStructParams[structName]; !ok {
			moreFuncs = append(moreFuncs, recvFuncs...)
		} else {
			for _, astFunc := range recvFuncs {
				if oldFunc, ok := nt.recvFuncsMap[astFunc.Name.Name]; !ok {
					zaplog.SUG.Debugln("more", astFunc.Name.Name)
					moreFuncs = append(moreFuncs, astFunc)
				} else {
					zaplog.SUG.Debugln("meet", oldFunc.Name.Name)
				}
			}
		}
	}

	if len(moreFuncs) == 0 {
		return oldX.content, false
	} else {
		var source = utils.Clone(oldX.content)
		for _, missFunc := range moreFuncs {
			name := missFunc.Name.Name
			zaplog.SUG.Debugln("more", name)
			if utils.C0IsUpper(name) {
				newName := []byte(utils.CvtC0Lower(name))
				oldName := syntaxgo_astnode.GetCode(source, missFunc.Name)
				must.Same(len(newName), len(oldName))
				copy(oldName, newName)
			}
		}
		mustslice.Different(source, oldX.content)
		return source, true
	}
}

// 思路基本是这样的：首先拿到所有的函数，其次过滤出有用的函数
// 把有用的函数以及它下面的无用部分看做整块代码，直到下个有用的函数为止的范围都是代码块
// 这个代码块，有1个函数名称，按照这个名称排序即可
func sortRecvFuncs(oldX *astParam, newX *astParam) {
	for structName, newStructParam := range newX.mapStructParams {
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(newX.GetNode(newStructParam.structType))
		zaplog.SUG.Debugln("-")
		oldStructParam, ok := oldX.mapStructParams[structName]
		must.True(ok)

		var astFuncs []*ast.FuncDecl
		for _, astFunc := range oldStructParam.recvFuncs {
			_, ok := newStructParam.recvFuncsMap[astFunc.Name.Name]
			if ok {
				astFuncs = append(astFuncs, astFunc)
			}
		}
		if len(astFuncs) == 0 {
			return
		}

		ptx := printgo.NewPTX()

		{
			node := syntaxgo_astnode.NewNode(1, astFuncs[0].Pos())
			if astFuncs[0].Doc != nil {
				checkDocPos(astFuncs[0])
				node.SetEnd(astFuncs[0].Doc.Pos())
			}
			startString := syntaxgo_astnode.GetText(oldX.content, node)
			ptx.Println(startString)
		}

		var elems []*tuple.T2[*syntaxgo_astnode.Node, string]
		for sdx, edx, a := 0, 1, astFuncs; edx < len(a); sdx, edx = sdx+1, edx+1 {
			head := a[sdx]
			next := a[edx]
			node := syntaxgo_astnode.NewNode(head.Pos(), next.Pos())
			if head.Doc != nil {
				checkDocPos(head)
				node.SetPos(head.Doc.Pos())
			}
			if next.Doc != nil {
				checkDocPos(head)
				node.SetEnd(next.Doc.Pos())
			}

			zaplog.SUG.Debugln(eroticgo.BLUE.Sprint(syntaxgo_astnode.GetText(oldX.content, node)))

			elems = append(elems, &tuple.T2[*syntaxgo_astnode.Node, string]{
				V1: node,
				V2: head.Name.Name,
			})
		}
		if len(astFuncs) > 0 {
			astLastFunc := utils.SoftLast(astFuncs)

			node := syntaxgo_astnode.NewNode(astLastFunc.Pos(), token.Pos(1+len(oldX.content)))
			if astLastFunc.Doc != nil {
				checkDocPos(astLastFunc)
				node.SetPos(astLastFunc.Doc.Pos())
			}

			zaplog.SUG.Debugln(eroticgo.BLUE.Sprint(syntaxgo_astnode.GetText(oldX.content, node)))

			elems = append(elems, &tuple.T2[*syntaxgo_astnode.Node, string]{
				V1: node,
				V2: astLastFunc.Name.Name,
			})
		}

		must.Same(len(astFuncs), len(elems))

		sortslice.SortIStable(elems, func(i, j int) bool {
			return newStructParam.idxAstFuncs[elems[i].V2] < newStructParam.idxAstFuncs[elems[j].V2]
		})
		for _, elem := range elems {
			nodeString := syntaxgo_astnode.GetText(oldX.content, elem.V1)
			ptx.Println(nodeString)
		}

		newCode, _ := formatgo.FormatBytes(ptx.Bytes())
		must.Done(os.WriteFile(oldX.path, newCode, 0644))
	}
}

func checkDocPos(astFunc *ast.FuncDecl) {
	if astFunc.Doc != nil {
		must.True(astFunc.Doc.Pos() < astFunc.Pos())
		must.True(astFunc.Doc.End() < astFunc.Pos())
	}
}

func collectAstParam(path string) *astParam {
	content := rese.V1(os.ReadFile(path))
	astBundle := rese.P1(syntaxgo_ast.NewAstBundleV1(content))
	astFile, _ := astBundle.GetBundle()

	structTypes := syntaxgo_search.MapStructTypesByName(astFile)
	zaplog.SUG.Debugln(len(structTypes))

	var mapStructParams = make(map[string]*astStructParam, len(structTypes))
	for structName, structType := range structTypes {
		zaplog.SUG.Debugln(structName)
		zaplog.SUG.Debugln("-")
		zaplog.SUG.Debugln(syntaxgo_astnode.GetText(content, structType))
		zaplog.SUG.Debugln("-")

		astFuncs := syntaxgo_search.FindFunctionsByReceiverName(astFile, structName, true)
		zaplog.SUG.Debugln("-")
		for _, astFunc := range astFuncs {
			zaplog.SUG.Debugln(astFunc.Name.Name)
		}
		zaplog.SUG.Debugln("-")

		var mapAstFuncs = make(map[string]*ast.FuncDecl, len(astFuncs))
		var idxAstFuncs = make(map[string]int, len(astFuncs))
		for idx, astFunc := range astFuncs {
			mapAstFuncs[astFunc.Name.Name] = astFunc
			idxAstFuncs[astFunc.Name.Name] = idx
		}

		mapStructParams[structName] = &astStructParam{
			structType:   structType,
			recvFuncs:    astFuncs,
			recvFuncsMap: mapAstFuncs,
			idxAstFuncs:  idxAstFuncs,
		}
	}
	return &astParam{
		path:            path,
		content:         content,
		mapStructParams: mapStructParams,
	}
}

type astParam struct {
	path            string
	content         []byte
	mapStructParams map[string]*astStructParam
}

func (x *astParam) GetNode(astNode ast.Node) string {
	return syntaxgo_astnode.GetText(x.content, astNode)
}

func (x *astParam) GetStructNode(structName string) string {
	return syntaxgo_astnode.GetText(x.content, x.mapStructParams[structName].structType)
}

type astStructParam struct {
	structType   *ast.StructType
	recvFuncs    []*ast.FuncDecl
	recvFuncsMap map[string]*ast.FuncDecl
	idxAstFuncs  map[string]int
}
