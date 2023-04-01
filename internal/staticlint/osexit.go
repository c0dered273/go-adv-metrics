package staticlint

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func GetOsExitCheck() []*analysis.Analyzer {
	osExit := &analysis.Analyzer{
		Name: "osExit",
		Doc:  "check not allowed os.Exit in main function",
		Run:  checkOsExit,
	}

	return []*analysis.Analyzer{osExit}
}

func checkOsExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		checkMain(pass, file)
	}
	return nil, nil
}

func checkMain(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if decl, ok := node.(*ast.FuncDecl); ok {
			if decl.Name.Name == "main" {
				checkCallExpr(pass, node)
			}
		}
		return true
	})
}

func checkCallExpr(pass *analysis.Pass, node ast.Node) {
	ast.Inspect(node, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			checkOSExitCall(pass, call)
		}
		return true
	})
}

func checkOSExitCall(pass *analysis.Pass, call *ast.CallExpr) {
	if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
		if x, ok := fun.X.(*ast.Ident); ok {
			reportIfExists(pass, x, fun.Sel)
		}
	}
}

func reportIfExists(pass *analysis.Pass, x *ast.Ident, sel *ast.Ident) {
	funcName := fmt.Sprintf("%s.%s", x.Name, sel.Name)
	if funcName == "os.Exit" {
		pass.Reportf(x.Pos(), "call os.Exit from main()")
	}
}
