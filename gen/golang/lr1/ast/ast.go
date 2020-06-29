//  Copyright 2020 Marius Ackerman
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package ast generates an Go AST for an LR(1) parser
package ast

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	"github.com/goccmack/gogll/ast"
	"github.com/goccmack/gogll/cfg"
	"github.com/goccmack/gogll/lr1/basicprod"
	"github.com/goccmack/goutil/ioutil"
)

type Data struct {
	Package    string
	Types      []string
	BasicProds []*BasicProd
}

type BasicProd struct {
	Comment    string
	ID         string
	Params     []*Param
	ReturnType string
}

type Param struct {
	ID   string
	Type string
}

func Gen(pkg string, bprods []*basicprod.Production) {
	fname := filepath.Join(cfg.BaseDir, "ast", "ast.go")
	if ioutil.Exist(fname) && !*cfg.All {
		// Do not regenerate
		return
	}

	tmpl, err := template.New("AST").Parse(src)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, getData(pkg, bprods)); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(fname, buf.Bytes()); err != nil {
		panic(err)
	}
}

func getData(pkg string, bprods []*basicprod.Production) *Data {
	return &Data{
		Package:    pkg,
		Types:      getTypes(bprods),
		BasicProds: getBasicProds(bprods),
	}
}

func getBasicProds(bprods []*basicprod.Production) (prods []*BasicProd) {
	for _, prod := range bprods {
		prods = append(prods, getBasicProd(prod))
	}
	return
}

func getBasicProd(prod *basicprod.Production) *BasicProd {
	return &BasicProd{
		Comment: fmt.Sprintf("%s : %s ;",
			prod.Head,
			strings.Join(prod.Body.GetSymbols(), " ")),

		ID: fmt.Sprintf("%s%d", prod.Head, prod.Alternate),

		Params: getParams(prod.Body),

		ReturnType: prod.Head,
	}
}

func getParams(body *ast.SyntaxAlternate) (params []*Param) {
	for i, sym := range body.Symbols {
		var param *Param
		if _, ok := sym.(*ast.NT); ok {
			param = &Param{
				ID:   strcase.ToLowerCamel(sym.String()),
				Type: sym.String(),
			}
		} else {
			param = &Param{
				ID:   fmt.Sprintf("symbol_%d", i),
				Type: "*token.Token",
			}

		}
		params = append(params, param)
	}
	return
}

func getTypes(bprods []*basicprod.Production) (types []string) {
	idMap := make(map[string]bool)
	for _, prod := range bprods {
		if _, exist := idMap[prod.Head]; !exist {
			types = append(types, prod.Head)
			idMap[prod.Head] = true
		}
	}
	return
}

const src = `// Generated by GoGLL.
package ast

import(
    "fmt"
)

{{range $bprod := .BasicProds}}{{$bp := $bprod}}
// {{$bp.Comment}}
func {{$bp.ID}}({{range $i, $p := $bp.Params}}{{if ne $i 0}}, {{end}}p{{$i}}{{end}} interface{})(interface{}, error){
    fmt.Println("ast.{{$bprod.ID}} is unimplemented")
    return nil, nil
}
{{end}}
`
