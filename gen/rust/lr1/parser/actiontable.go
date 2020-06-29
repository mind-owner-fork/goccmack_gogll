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

package parser

import (
	"bytes"
	"fmt"
	"path"
	"text/template"

	"github.com/goccmack/gogll/lr1/action"
	"github.com/goccmack/gogll/lr1/basicprod"
	"github.com/goccmack/gogll/lr1/states"
	"github.com/goccmack/gogll/symbols"
	"github.com/goccmack/goutil/ioutil"
)

func genActionTable(pkg, outDir string, prods []*basicprod.Production, states *states.States, actions action.Actions) {
	tmpl, err := template.New("parser action table").Parse(actionTableSrc)
	if err != nil {
		panic(err)
	}
	wr := new(bytes.Buffer)
	if err := tmpl.Execute(wr, getActionTableData(pkg, prods, states, actions)); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(
		path.Join(outDir, "src", "parser", "action_table", "mod.rs"),
		wr.Bytes()); err != nil {

		panic(err)
	}
}

type actionTableData struct {
	Package string
	Rows    []*row
	NumRows int
}

type row struct {
	CanRecover bool
	Actions    []*actionRec
	NumActions int
}

type actionRec struct {
	Token  string
	Action string
}

func getActionTableData(
	pkg string,
	prods []*basicprod.Production, states *states.States, actions action.Actions,
) (actTab *actionTableData) {
	actTab = &actionTableData{
		Package: pkg,
		Rows:    make([]*row, states.Size()),
		NumRows: states.Size(),
	}
	for i := range actTab.Rows {
		actTab.Rows[i] = getActionRow(prods, states.List[i], actions[i])
	}
	return
}

func getActionRow(prods []*basicprod.Production, state *states.State, actions map[string]action.Action) (data *row) {
	data = &row{
		CanRecover: state.CanRecover(),
		Actions:    []*actionRec{},
		NumActions: len(symbols.GetNonTerminals()),
	}
	for _, sym := range symbols.GetTerminals() {
		if actions[sym.Literal()] != nil {
			var actStr string
			switch act := actions[sym.Literal()].(type) {
			case action.Accept:
				actStr = fmt.Sprintf("Accept,\t\t/* %s */", sym)
			case action.Reduce:
				actStr = fmt.Sprintf("Reduce(%d),\t\t/* %s, Reduce: %s */", int(act), sym, prods[int(act)].Head)
			case action.Shift:
				actStr = fmt.Sprintf("Shift(%d),\t\t/* %s */", int(act), sym)
			default:
				panic(fmt.Sprintf("Unknown action type: %T", act))
			}
			data.Actions = append(data.Actions,
				&actionRec{
					Token:  sym.GoString(),
					Action: actStr,
				})
		}
	}
	return
}

const actionTableSrc = `//! Generated by GoGLL. Do not edit.

use crate::token;

use std::collections::HashMap;
use lazy_static::lazy_static;

use super::productions_table::PROD_TABLE;

#[derive(PartialEq, Eq)]
pub enum Action {
	Accept,
	Shift(usize),
	Reduce(usize),
}

impl std::fmt::Display for Action {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		match self {
			Action::Accept => write!(f, "Accept(0)"),
			Action::Reduce(s) => write!(f, "Reduce({}) {}", s, PROD_TABLE[*s].string),
			Action::Shift(s) => write!(f, "Shift({})", s),
		}
    }
}

#[derive(std::default::Default)]
pub struct ActionRow {
	pub can_recover: bool,
	pub actions: HashMap<token::Type,  Action>,
}

lazy_static! {
	pub static ref ACTION_TABLE: Vec<ActionRow> = {
		let mut v: Vec<ActionRow> = Vec::with_capacity({{.NumRows}});
        v.resize_with({{.NumRows}}, Default::default);
		{{range $i, $row := .Rows}}
		v[{{$i}}].can_recover = {{printf "%t" $row.CanRecover}};{{range $a := .Actions}}
		v[{{$i}}].actions.insert(token::Type::{{$a.Token}}, Action::{{$a.Action}}); {{end}}{{end}}

		v
	};
}
`
