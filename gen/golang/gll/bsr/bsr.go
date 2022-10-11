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

// Package bsr generates a Go BSR package
package bsr

import (
	"bytes"
	"text/template"

	"github.com/goccmack/goutil/ioutil"
)

func Gen(bsrFile string, pkg string) {
	tmpl, err := template.New("bsr").Parse(bsrTmpl)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, pkg); err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(bsrFile, buf.Bytes()); err != nil {
		panic(err)
	}
}

const bsrTmpl = `// Package bsr is generated by gogll. Do not edit.

/*
Package bsr implements a Binary Subtree Representation set as defined in

    Scott et al
    Derivation representation using binary subtree sets,
    Science of Computer Programming 175 (2019)
*/
package bsr

import (
    "bytes"
    "fmt"
    "sort"
    "strings"

    "{{.}}/lexer"
    "{{.}}/parser/slot"
    "{{.}}/parser/symbols"
    "{{.}}/sppf"
    "{{.}}/token"
)

type bsr interface {
    LeftExtent() int
    RightExtent() int
    Pivot() int
}

/*
Set contains the set of Binary Subtree Representations (BSR).
*/
type Set struct {
    slotEntries   map[BSR]bool
    ntSlotEntries map[ntSlot][]BSR
    stringEntries map[stringKey]*stringBSR
    rightExtent   int
    lex           *lexer.Lexer

    startSym symbols.NT
}

type ntSlot struct {
    nt          symbols.NT
    leftExtent  int
    rightExtent int
}

// BSR is the binary subtree representation of a parsed nonterminal
type BSR struct {
    Label       slot.Label
    leftExtent  int
    pivot       int
    rightExtent int
    set         *Set
}

type BSRs []BSR

type stringBSR struct {
    Symbols     symbols.Symbols
    leftExtent  int
    pivot       int
    rightExtent int
    set         *Set
}

type stringBSRs []*stringBSR

type stringKey string

// New returns a new initialised BSR Set
func New(startSymbol symbols.NT, l *lexer.Lexer) *Set {
    return &Set{
        slotEntries:   make(map[BSR]bool),
        ntSlotEntries: make(map[ntSlot][]BSR),
        stringEntries: make(map[stringKey]*stringBSR),
        rightExtent:   0,
        lex:           l,
        startSym:      startSymbol,
    }
}

/*
Add a bsr to the set. (i,j) is the extent. k is the pivot.
*/
func (s *Set) Add(l slot.Label, i, k, j int) {
    // fmt.Printf("bsr.Add(%s,%d,%d,%d l.Pos %d)\n", l, i, k, j, l.Pos())
    if l.EoR() {
        s.insert(BSR{l, i, k, j, s})
    } else {
        if l.Pos() > 1 {
            s.insert(&stringBSR{l.Symbols()[:l.Pos()], i, k, j, s})
        }
    }
}

// AddEmpty adds a grammar slot: X : ϵ•
func (s *Set) AddEmpty(l slot.Label, i int) {
    s.insert(BSR{l, i, i, i, s})
}

/*
Contain returns true iff the BSR Set contains the NT symbol with left and
right extent.
*/
func (s *Set) Contain(nt symbols.NT, left, right int) bool {
    // fmt.Printf("bsr.Contain(%s,%d,%d)\n",nt,left,right)
    for e := range s.slotEntries {
        // fmt.Printf("  (%s,%d,%d)\n",e.Label.Head(),e.leftExtent,e.rightExtent)
        if e.Label.Head() == nt && e.leftExtent == left && e.rightExtent == right {
            // fmt.Println("  true")
            return true
        }
    }
    // fmt.Println("  false")
    return false
}

// Dump prints all the NT and string elements of the BSR set
func (s *Set) Dump() {
    fmt.Println("Roots:")
    for _, rt := range s.GetRoots() {
        fmt.Println(rt)
    }
    fmt.Println()

    fmt.Println("NT BSRs:")
    for _, bsr := range s.getNTBSRs() {
        fmt.Println(bsr)
    }
    fmt.Println()

    fmt.Println("string BSRs:")
    for _, bsr := range s.getStringBSRs() {
        fmt.Println(bsr)
    }
    fmt.Println()
}

// GetAll returns all BSR grammar slot entries
func (s *Set) GetAll() (bsrs []BSR) {
    for b := range s.slotEntries {
        bsrs = append(bsrs, b)
    }
    return
}

// GetRightExtent returns the right extent of the BSR set
func (s *Set) GetRightExtent() int {
    return s.rightExtent
}

// GetRoot returns the root of the parse tree of an unambiguous parse.
// GetRoot fails if the parse was ambiguous. Use GetRoots() for ambiguous parses.
func (s *Set) GetRoot() BSR {
    rts := s.GetRoots()
    if len(rts) != 1 {
        failf("%d parse trees exist for start symbol %s", len(rts), s.startSym)
    }
    return rts[0]
}

// GetRoots returns all the roots of parse trees of the start symbol of the grammar.
func (s *Set) GetRoots() (roots []BSR) {
    for b := range s.slotEntries {
        if b.Label.Head() == s.startSym && b.leftExtent == 0 && s.rightExtent == b.rightExtent {
            roots = append(roots, b)
        }
    }
    return
}

// GetAllStrings returns all string elements with symbols = str,
// left extent = lext and right extent = rext
func (s *Set) GetAllStrings(str symbols.Symbols, lext, rext int) (strs []*stringBSR) {
    for _, s := range s.stringEntries {
        if s.Symbols.Equal(str) && s.leftExtent == lext && s.rightExtent == rext {
            strs = append(strs, s)
        }
    }
    return
}

func (s *Set) getNTBSRs() BSRs {
    bsrs := make(BSRs, 0, len(s.ntSlotEntries))
    for _, bsrl := range s.ntSlotEntries {
        for _, bsr := range bsrl {
            bsrs = append(bsrs, bsr)
        }
    }
    sort.Sort(bsrs)
    return bsrs
}

func (s *Set) getStringBSRs() stringBSRs {
    bsrs := make(stringBSRs, 0, len(s.stringEntries))
    for _, bsr := range s.stringEntries {
        bsrs = append(bsrs, bsr)
    }
    sort.Sort(bsrs)
    return bsrs
}

func (s *Set) getString(symbols symbols.Symbols, leftExtent, rightExtent int) *stringBSR {
    // fmt.Printf("Set.getString(%s,%d,%d)\n", symbols, leftExtent, rightExtent)

    strBsr, exist := s.stringEntries[getStringKey(symbols, leftExtent, rightExtent)]
    if exist {
        return strBsr
    }

    panic(fmt.Sprintf("Error: no string %s left extent=%d right extent=%d\n",
        symbols, leftExtent, rightExtent))
}

func (s *Set) insert(bsr bsr) {
    if bsr.RightExtent() > s.rightExtent {
        s.rightExtent = bsr.RightExtent()
    }
    switch b := bsr.(type) {
    case BSR:
        s.slotEntries[b] = true
        nt := ntSlot{b.Label.Head(), b.leftExtent, b.rightExtent}
        s.ntSlotEntries[nt] = append(s.ntSlotEntries[nt], b)
    case *stringBSR:
        s.stringEntries[b.key()] = b
    default:
        panic(fmt.Sprintf("Invalid type %T", bsr))
    }
}

func (s *stringBSR) key() stringKey {
    return getStringKey(s.Symbols, s.leftExtent, s.rightExtent)
}

func getStringKey(symbols symbols.Symbols, lext, rext int) stringKey {
    return stringKey(fmt.Sprintf("%s,%d,%d", symbols, lext, rext))
}

// Alternate returns the index of the grammar rule alternate.
func (b BSR) Alternate() int {
    return b.Label.Alternate()
}

// GetAllNTChildren returns all the NT Children of b. If an NT child of b has
// ambiguous parses then all parses of that child are returned.
func (b BSR) GetAllNTChildren() [][]BSR {
    children := [][]BSR{}
    for i, s := range b.Label.Symbols() {
        if s.IsNonTerminal() {
            sChildren := b.GetNTChildrenI(i)
            children = append(children, sChildren)
        }
    }
    return children
}

// GetNTChild returns the BSR of occurrence i of nt in s.
// GetNTChild fails if s has ambiguous subtrees of occurrence i of nt.
func (b BSR) GetNTChild(nt symbols.NT, i int) BSR {
    bsrs := b.GetNTChildren(nt, i)
    if len(bsrs) != 1 {
        ambiguousSlots := []string{}
        for _, c := range bsrs {
            ambiguousSlots = append(ambiguousSlots, c.String())
        }
        b.set.fail(b, "%s is ambiguous in %s\n  %s", nt, b, strings.Join(ambiguousSlots, "\n  "))
    }
    return bsrs[0]
}

// GetNTChildI returns the BSR of NT symbol[i] in the BSR set.
// GetNTChildI fails if the BSR set has ambiguous subtrees of NT i.
func (b BSR) GetNTChildI(i int) BSR {
    bsrs := b.GetNTChildrenI(i)
    if len(bsrs) != 1 {
        b.set.fail(b, "NT %d is ambiguous in %s", i, b)
    }
    return bsrs[0]
}

// GetNTChildren returns all the BSRs of occurrence i of nt in s
func (b BSR) GetNTChildren(nt symbols.NT, i int) []BSR {
    // fmt.Printf("GetNTChild(%s,%d) %s\n", nt, i, b)
    positions := []int{}
    for j, s := range b.Label.Symbols() {
        if s == nt {
            positions = append(positions, j)
        }
    }
    if len(positions) == 0 {
        b.set.fail(b, "Error: %s has no NT %s", b, nt)
    }
    return b.GetNTChildrenI(positions[i])
}

// GetNTChildrenI returns all the BSRs of NT symbol[i] in s
func (b BSR) GetNTChildrenI(i int) []BSR {
    // fmt.Printf("bsr.GetNTChildI(%d) %s Pos %d\n", i, b, b.Label.Pos())

    if i >= len(b.Label.Symbols()) {
        b.set.fail(b, "Error: cannot get NT child %d of %s", i, b)
    }
    if len(b.Label.Symbols()) == 1 {
        return b.set.getNTSlot(b.Label.Symbols()[i], b.pivot, b.rightExtent)
    }
    if len(b.Label.Symbols()) == 2 {
        if i == 0 {
            return b.set.getNTSlot(b.Label.Symbols()[i], b.leftExtent, b.pivot)
        }
        return b.set.getNTSlot(b.Label.Symbols()[i], b.pivot, b.rightExtent)
    }
    if b.Label.Pos() == i+1 {
        return b.set.getNTSlot(b.Label.Symbols()[i], b.pivot, b.rightExtent)
    }

    // Walk to pos i from the right
    symbols := b.Label.Symbols()[:b.Label.Pos()-1]
    str := b.set.getString(symbols, b.leftExtent, b.pivot)
    for len(symbols) > i+1 && len(symbols) > 2 {
        symbols = symbols[:len(symbols)-1]
        str = b.set.getString(symbols, str.leftExtent, str.pivot)
    }

    bsrs := []BSR{}
    if i == 0 {
        bsrs = b.set.getNTSlot(b.Label.Symbols()[i], str.leftExtent, str.pivot)
    } else {
        bsrs = b.set.getNTSlot(b.Label.Symbols()[i], str.pivot, str.rightExtent)
    }

    // fmt.Println(bsrs)

    return bsrs
}

// GetTChildI returns the terminal symbol at position i in b.
// GetTChildI panics if symbol i is not a valid terminal
func (b BSR) GetTChildI(i int) *token.Token {
    symbols := b.Label.Symbols()

    if i >= len(symbols) {
        panic(fmt.Sprintf("%s has no T child %d", b, i))
    }
    if symbols[i].IsNonTerminal() {
        panic(fmt.Sprintf("symbol %d in %s is an NT", i, b))
    }

    lext := b.leftExtent
    for j := 0; j < i; j++ {
        if symbols[j].IsNonTerminal() {
            nt := b.GetNTChildI(j)
            lext += nt.rightExtent - nt.leftExtent
        } else {
            lext++
        }
    }
    return b.set.lex.Tokens[lext]
}

// LeftExtent returns the left extent of the BSR
func (b BSR) LeftExtent() int {
    return b.leftExtent
}

// RightExtent returns the right extent of the BSR
func (b BSR) RightExtent() int {
    return b.rightExtent
}

// Pivot returns the pivot of the BSR
func (b BSR) Pivot() int {
    return b.pivot
}

func (b BSR) String() string {
    srcStr := "ℇ"
    if b.leftExtent < b.rightExtent {
        srcStr = b.set.lex.GetString(b.LeftExtent(), b.RightExtent()-1)
    }
    return fmt.Sprintf("%s,%d,%d,%d - %s",
        b.Label, b.leftExtent, b.pivot, b.rightExtent, srcStr)
}

// BSRs Sort interface
func (bs BSRs) Len() int {
    return len(bs)
}

func (bs BSRs) Less(i, j int) bool {
    if bs[i].Label < bs[j].Label {
        return true
    }
    if bs[i].Label > bs[j].Label {
        return false
    }
    if bs[i].leftExtent < bs[j].leftExtent {
        return true
    }
    if bs[i].leftExtent > bs[j].leftExtent {
        return false
    }
    return bs[i].rightExtent < bs[j].rightExtent
}

func (bs BSRs) Swap(i, j int) {
    bs[i], bs[j] = bs[j], bs[i]
}

// stringBSRs Sort interface
func (sbs stringBSRs) Len() int {
    return len(sbs)
}

func (sbs stringBSRs) Less(i, j int) bool {
    if sbs[i].Symbols.String() < sbs[j].Symbols.String() {
        return true
    }
    if sbs[i].Symbols.String() > sbs[j].Symbols.String() {
        return false
    }
    if sbs[i].leftExtent < sbs[j].leftExtent {
        return true
    }
    if sbs[i].leftExtent > sbs[j].leftExtent {
        return false
    }
    return sbs[i].rightExtent < sbs[j].rightExtent
}

func (sbs stringBSRs) Swap(i, j int) {
    sbs[i], sbs[j] = sbs[j], sbs[i]
}

func (s stringBSR) LeftExtent() int {
    return s.leftExtent
}

func (s stringBSR) RightExtent() int {
    return s.rightExtent
}

func (s stringBSR) Pivot() int {
    return s.pivot
}

func (s stringBSR) Empty() bool {
    return s.leftExtent == s.pivot && s.pivot == s.rightExtent
}

// String returns a string representation of s
func (s stringBSR) String() string {
    return fmt.Sprintf("%s,%d,%d,%d - %s", &s.Symbols, s.leftExtent, s.pivot,
        s.rightExtent, s.set.lex.GetString(s.LeftExtent(), s.RightExtent()))
}

func (s *Set) getNTSlot(sym symbols.Symbol, leftExtent, rightExtent int) (bsrs []BSR) {
    nt, ok := sym.(symbols.NT)
    if !ok {
        line, col := s.getLineColumn(leftExtent)
        failf("%s is not an NT at line %d col %d", sym, line, col)
    }
    return s.ntSlotEntries[ntSlot{nt, leftExtent, rightExtent}]
}

func (s *Set) fail(b BSR, format string, a ...interface{}) {
    msg := fmt.Sprintf(format, a...)
    line, col := s.getLineColumn(b.LeftExtent())
    panic(fmt.Sprintf("Error in BSR: %s at line %d col %d\n", msg, line, col))
}

func failf(format string, args ...interface{}) {
    panic(fmt.Sprintf("Error in BSR: %s\n", fmt.Sprintf(format, args...)))
}

func (s *Set) getLineColumn(cI int) (line, col int) {
    return s.lex.GetLineColumnOfToken(cI)
}

// ReportAmbiguous lists the ambiguous subtrees of the parse forest
func (s *Set) ReportAmbiguous() {
    fmt.Println("Ambiguous BSR Subtrees:")
    rts := s.GetRoots()
    if len(rts) != 1 {
        fmt.Printf("BSR has %d ambigous roots\n", len(rts))
    }
    for i, b := range s.GetRoots() {
        fmt.Println("In root", i)
        if !s.report(b) {
            fmt.Println("No ambiguous BSRs")
        }
    }
}

// report return true iff at least one ambigous BSR was found
func (s *Set) report(b BSR) bool {
    ambiguous := false
    for i, sym := range b.Label.Symbols() {
        ln, col := s.getLineColumn(b.LeftExtent())
        if sym.IsNonTerminal() {
            if len(b.GetNTChildrenI(i)) != 1 {
                ambiguous = true
                fmt.Printf("  Ambigous: in %s: NT %s (%d) at line %d col %d \n",
                    b, sym, i, ln, col)
                fmt.Println("   Children:")
                for _, c := range b.GetNTChildrenI(i) {
                    fmt.Printf("     %s\n", c)
                }
            }
            for _, b1 := range b.GetNTChildrenI(i) {
                s.report(b1)
            }
        }
    }
    return ambiguous
}

// IsAmbiguous returns true if the BSR set does not have exactly one root, or
// if any BSR in the set has an NT symbol, which does not have exactly one
// sub-tree.
func (s *Set) IsAmbiguous() bool {
    if len(s.GetRoots()) != 1 {
        return true
    }
    return isAmbiguous(s.GetRoot())
}

// isAmbiguous returns true if b or any of its NT children is ambiguous.
// A BSR is ambiguous if any of its NT symbols does not have exactly one
// subtrees (children).
func isAmbiguous(b BSR) bool {
    for i, s := range b.Label.Symbols() {
        if s.IsNonTerminal() {
            if len(b.GetNTChildrenI(i)) != 1 {
                return true
            }
            for _, b1 := range b.GetNTChildrenI(i) {
                if isAmbiguous(b1) {
                    return true
                }
            }
        }
    }
    return false
}

//---- SPPF ------------

type bldSPPF struct {
    root         *sppf.SymbolNode
    extLeafNodes []sppf.Node
    pNodes       map[string]*sppf.PackedNode
    sNodes       map[string]*sppf.SymbolNode // Index is Node.Label()
}

func (pf *Set) ToSPPF() *sppf.SymbolNode {
    bld := &bldSPPF{
        pNodes: map[string]*sppf.PackedNode{},
        sNodes: map[string]*sppf.SymbolNode{},
    }
    rt := pf.GetRoots()[0]
    bld.root = bld.mkSN(rt.Label.Head().String(), rt.leftExtent, rt.rightExtent)

    for len(bld.extLeafNodes) > 0 {
        // let w = (μ, i, j) be an extendable leaf node of G
        w := bld.extLeafNodes[len(bld.extLeafNodes)-1]
        bld.extLeafNodes = bld.extLeafNodes[:len(bld.extLeafNodes)-1]

        // μ is a nonterminal X in Γ
        if nt, ok := w.(*sppf.SymbolNode); ok && symbols.IsNT(nt.Symbol) {
            bsts := pf.getNTSlot(symbols.ToNT(nt.Symbol), nt.Lext, nt.Rext)
            // for each (X ::=γ,i,k, j)∈Υ { mkPN(X ::=γ·,i,k, j,G) } }
            for _, bst := range bsts {
                slt := bst.Label.Slot()
                nt.Children = append(nt.Children,
                    bld.mkPN(slt.NT, slt.Symbols, slt.Pos,
                        bst.leftExtent, bst.pivot, bst.rightExtent))
            }
        } else { // w is an intermediate node
            // suppose μ is X ::=α·δ
            in := w.(*sppf.IntermediateNode)
            if in.Pos == 1 {
                in.Children = append(in.Children, bld.mkPN(in.NT, in.Body, in.Pos,
                    in.Lext, in.Lext, in.Rext))
            } else {
                // for each (α,i,k, j)∈Υ { mkPN(X ::=α·δ,i,k, j,G) } } } }
                alpha, delta := in.Body[:in.Pos], in.Body[in.Pos:]
                for _, str := range pf.GetAllStrings(alpha, in.Lext, in.Rext) {
                    body := append(str.Symbols, delta...)
                    in.Children = append(in.Children,
                        bld.mkPN(in.NT, body, in.Pos, str.leftExtent, str.pivot, str.rightExtent))
                }
            }
        }
    }
    return bld.root
}

func (bld *bldSPPF) mkIN(nt symbols.NT, body symbols.Symbols, pos int,
    lext, rext int) *sppf.IntermediateNode {

    in := &sppf.IntermediateNode{
        NT:   nt,
        Body: body,
        Pos:  pos,
        Lext: lext,
        Rext: rext,
    }
    bld.extLeafNodes = append(bld.extLeafNodes, in)
    return in
}

func (bld *bldSPPF) mkPN(nt symbols.NT, body symbols.Symbols, pos int,
	lext, pivot, rext int) *sppf.PackedNode {
	// fmt.Printf("mkPN %s,%d,%d,%d\n", slotString(nt, body, pos), lext, pivot, rext)

	// X ::= ⍺ • β, k
	pn := &sppf.PackedNode{
		NT:         nt,
		Body:       body,
		Pos:        pos,
		Lext:       lext,
		Rext:       rext,
		Pivot:      pivot,
		LeftChild:  nil,
		RightChild: nil,
	}
	if pn1, exist := bld.pNodes[pn.Label()]; exist {
		return pn1
	}
	bld.pNodes[pn.Label()] = pn

	if len(body) == 0 { // ⍺ = ϵ
		pn.RightChild = bld.mkSN("ϵ", lext, lext)
	} else { // if ( α=βx, where |x|=1) {
		// mkN(x,k, j, y,G)
		pn.RightChild = bld.mkSN(pn.Body[pn.Pos-1].String(), pivot, rext)

		// if (|β|=1) mkN(β,i,k,y,G)
		if pos == 2 {
			pn.LeftChild = bld.mkSN(pn.Body[pn.Pos-2].String(), lext, pivot)
		}
		// if (|β|>1) mkN(X ::=β·xδ,i,k,y,G)
		if pos > 2 {
			pn.LeftChild = bld.mkIN(pn.NT, pn.Body, pn.Pos-1, lext, pivot)
		}
	}

	return pn
}

func (bld *bldSPPF) mkSN(symbol string, lext, rext int) *sppf.SymbolNode {
	sn := &sppf.SymbolNode{
		Symbol: symbol,
		Lext:   lext,
		Rext:   rext,
	}
	if sn1, exist := bld.sNodes[sn.Label()]; exist {
		return sn1
	}
	bld.sNodes[sn.Label()] = sn
	if symbols.IsNT(symbol) {
		bld.extLeafNodes = append(bld.extLeafNodes, sn)
	}
	return sn
}

func slotString(nt symbols.NT, body symbols.Symbols, pos int) string {
    w := new(bytes.Buffer)
    fmt.Fprintf(w, "%s:", nt)
    for i, sym := range body {
        fmt.Fprint(w, " ")
        if i == pos {
            fmt.Fprint(w, "•")
        }
        fmt.Fprint(w, sym)
    }
    if len(body) == pos {
        fmt.Fprint(w, "•")
    }
    return w.String()
}

`
