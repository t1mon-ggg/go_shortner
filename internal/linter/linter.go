package linter

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer - linter for os.Exit in main function
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osExit",
	Doc:  "check for os.Exit in main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		packName := f.Name
		funcName := ""
		if packName.Name == "main" {
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					res := fmt.Sprintf("%s", x.Fun)
					if res == "&{os Exit}" && funcName == "main" {
						pass.Reportf(x.Pos(), "os.Exit usage")
					}

				case *ast.FuncDecl:
					funcName = x.Name.Name
				}
				return true
			})
		}
	}
	return nil, nil
}
