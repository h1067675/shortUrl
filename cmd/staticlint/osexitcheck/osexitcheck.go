// Package osexitcheck проверяет не используется ли прямой вызов os.Exit в функции main пакета main
package osexitcheck

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer определяет анализатор
var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for don't use os.Exit in main package",
	Run:  run,
}

// Pass определяет структуру для анализатора
type Pass struct {
	Fset         *token.FileSet // информация о позиции токенов
	Files        []*ast.File    // AST для каждого файла
	OtherFiles   []string       // имена файлов не на Go в пакете
	IgnoredFiles []string       // имена игнорируемых исходных файлов в пакете
	Pkg          *types.Package // информация о типах пакета
	TypesInfo    *types.Info    // информация о типах в AST
}

// run запускает анализатор
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверяем имя пакета
		if pass.Pkg.Name() == "main" {
			// Определяем переменную для хранения имени текущей функции
			var currentFunc string
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				// Проверяем является функцией
				case *ast.FuncDecl:
					// Если является функцией присваиваем значение переменной
					currentFunc = x.Name.Name
					// Проверяем является ли вызовом функции
				case *ast.CallExpr:
					// Проверяем находится ли вызов функции в функции main
					if currentFunc == "main" {
						call, ok := node.(*ast.CallExpr)
						if !ok {
							return true
						}

						// Находим индентификатор вызова
						var ident *ast.Ident
						switch fun := call.Fun.(type) {
						case *ast.Ident:
							ident = fun
						case *ast.SelectorExpr:
							ident = fun.Sel
						}

						// Если пустой идем к следующему
						if ident == nil {
							return true
						}
						obj := pass.TypesInfo.Uses[ident]
						if obj == nil {
							return true
						}

						// Проверяем пакет и имя
						if obj.Pkg() != nil && obj.Pkg().Path() == "os" && obj.Name() == "Exit" {
							pass.Reportf(call.Pos(), "used os.Exit")
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
