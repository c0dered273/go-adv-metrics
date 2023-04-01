package main

import (
	"github.com/MakeNowJust/enumcase"
	"github.com/c0dered273/go-adv-metrics/internal/staticlint"
	"github.com/gostaticanalysis/sqlrows/passes/sqlrows"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

// Выполняет проверку исходного кода комплектом статических анализаторов
// Использование:
//
//	staticlint [-flag] [package]
//	Запустить staticlint help для вывода помощи
//
// В комплект входят следующие анализаторы:
// * все анализаторы из пакета golang.org/x/tools/go/analysis
// * staticcheck - содержит более 150 различных проверок
// * enumcase - проверяет каждый switch на использование всех const значений типа
// * sqlrows - ищет ошибки при использовании sql.Rows
// * osExit - ищет вызов os.Exit() в функции main()
func main() {
	checks := staticlint.GetChecks(
		staticlint.GetStdChecks(),
		staticlint.GetStaticChecks(),
		staticlint.GetOsExitCheck(),
		[]*analysis.Analyzer{enumcase.Analyzer},
		[]*analysis.Analyzer{sqlrows.Analyzer},
	)

	multichecker.Main(
		checks...,
	)
}
