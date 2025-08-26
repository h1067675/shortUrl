// Package mycheck представляет собой пакет multichecker
package main

import (
	"encoding/json"
	"mycheck/osexitcheck"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow" // импортируем дополнительный анализатор
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

// Config определяет файл с настройками пакетов staticcheck.io
// по умолчанию в настройки включены все проверки SA, а также проверки
// S1011 проверяет на использование одного append для объединения двух срезов и
// QF1003 проверяет на отсутствие цепочек if/else-if
//
// в файле настроек возможно использовать маску для определения пакетов используя *
const Config = `config.json`

// ConfigData определяет структуру для парсинга файла настрок
type ConfigData struct {
	Staticcheck []string
}

func main() {
	// получаем директорию текущего файла
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	// читаем файл
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}
	// помещаем настройки из файла в структуру
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	// оаределяем используемые анализаторы
	mychecks := []*analysis.Analyzer{
		osexitcheck.Analyzer,

		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		ifaceassert.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		httpresponse.Analyzer,

		bodyclose.Analyzer,
		errcheck.Analyzer,
	}
	// собираем настройки из staticcheck.io исходя из масок настройки
	checks := make(map[string]bool)
	var checksReg []string
	for _, v := range cfg.Staticcheck {
		if strings.Contains(v, "*") {
			checksReg = append(checksReg, v)
		} else {
			checks[v] = true
		}
	}

	// добавляем в массив нужные проверки
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		} else {
			for _, e := range checksReg {
				match, err := path.Match(e, v.Analyzer.Name)
				if err != nil {
					panic(err)
				}
				if match {
					mychecks = append(mychecks, v.Analyzer)
				}
			}
		}
	}

	// запускаяем multichecker
	multichecker.Main(mychecks...)
}
