
// Module bsr is generated by gogll. Do not edit.

/*
Module bsr implements a Binary Subtree Representation set as defined in

    Scott et al
    Derivation representation using binary subtree sets,
    Science of Computer Programming 175 (2019)

ToDo:

* StrBSR is specific to a grammar slot rather than a symbol string. In an
ambiguous grammar this leads to more string BSRs than necessary.
*/

use crate::lexer;
use crate::parser::{slot, symbols};
use crate::parser::symbols::{NT, Symbol};
use crate::token::{Token};

use std::cmp::Ordering;
use std::cmp::Ordering::{Less, Greater};
use std::collections::HashMap;
use std::rc::Rc;
use std::fmt;

// The kind of BSR added.
enum Kind {
    NT(Rc<BSR>),
    Str(Rc<BSR>),
}

/**
Set contains the set of Binary Subtree Representations (BSR).
*/
#[allow(dead_code)]
pub struct Set {
    slot_entries: HashMap<Rc<BSR>, bool>,
    nt_slot_entries: HashMap<NTSlot, Vec<Rc<BSR>>>,
    string_entries: HashMap<Rc<BSR>, bool>,
    pub rext: usize,
    lex: Rc<lexer::Lexer>,

    start_sym: NT,
}

#[derive(Hash, Eq, PartialEq)]
struct NTSlot {
    nt: NT,
    lext: usize,
    rext: usize,
}

/// BSR is the binary subtree representation of a parsed nonterminal
#[derive(Hash, Eq, PartialEq)]
pub struct BSR {
    pub label: slot::Label,
    pub lext: usize,
    pub pivot: usize,
    pub rext: usize,
}

impl BSR {
    fn cmp(&self, other: &Self) -> Ordering {
        if self.lext < other.lext {
            return Less;
        }
        if self.lext > other.lext {
            return Greater;
        }
        // self.lext == other.lext
        if self.rext > other.rext {
            return Less;
        }
        if self.rext < other.rext {
            return Greater;
        }
        // self.rext == other.rext
        if self.pivot < other.pivot {
            return Less;
        }
        Greater
    }
}

impl Set {
    /// New returns a new initialised BSR Set
    #[allow(dead_code)]
    pub fn new(start_symbol: NT, l: Rc<lexer::Lexer>) -> Box<Set> {
        Box::new(Set {
            slot_entries: HashMap::with_capacity(1024),
            nt_slot_entries: HashMap::with_capacity(1024),
            string_entries: HashMap::with_capacity(1024),
            rext: 0,
            lex: l.clone(),
            start_sym: start_symbol,
        })
    }

    /// Add a BSR to the set. (i,j) is the extent. k is the pivot.
    #[allow(dead_code)]
    pub fn add(&mut self, l: slot::Label, i: usize, k: usize, j: usize) {
        let b = Rc::new(BSR {
            label: l,
            lext: i,
            pivot: k,
            rext: j,
        });
        if l.eor() {
            self.insert(Kind::NT(b))
        } else {
            if l.pos() > 1 {
                self.insert(Kind::Str(b))
            }
        }
    }

    /// Returns the index of the grammar rule alternate.
    #[allow(dead_code)]
    pub fn alternate(&self, b: Rc<BSR>) -> usize {
    	return b.label.alternate()
    }

    fn insert(&mut self, bsr: Kind) {
        if bsr.rext() > self.rext {
            self.rext = bsr.rext()
        }
        match bsr {
            Kind::NT(b) => {
                self.slot_entries.insert(b.clone(), true);
                let nt_slot = NTSlot::new(b.label.head(), b.lext, b.rext);
                match self.nt_slot_entries.get_mut(&nt_slot) {
                    None => {
                        self.nt_slot_entries.insert(nt_slot, vec![b.clone()]);
                    },
                    Some(bsrs) => bsrs.push(b.clone())
                }
            }
            Kind::Str(b) => {
                self.string_entries.insert(b.clone(), true);
            }
        };
    }

    /// AddEmpty adds a grammar slot: X : ϵ•
    #[allow(dead_code)]
    pub fn add_empty(&mut self, l: slot::Label, i: usize) {
        self.insert(Kind::NT(Rc::new(BSR {
            label: l,
            lext: i,
            pivot: i,
            rext: i,
        })))
    }

    /**
    contain returns true iff the BSR Set contains the NT symbol with left and
    right extent.
    */
    #[allow(dead_code)]
    pub fn contain(&self, nt: &NT, left: usize, right: usize) -> bool {
        for e in self.slot_entries.keys() {
            if e.label.head() == nt && e.lext == left && e.rext == right {
                return true;
            }
        }
        return false;
    }

    /// Returns all the NT BSR entries. Used for debugging.
    #[allow(dead_code)]
    pub fn get_all(&self) -> Vec<Rc<BSR>> {
        let mut bsrs: Vec<Rc<BSR>> = Vec::with_capacity(128);
        for b in self.slot_entries.keys() {
            bsrs.push(b.clone());
        }
        bsrs.sort_by(|a, b| a.cmp(b));
        bsrs
    }

    // get_root returns the root of the parse tree of an unambiguous parse.
    // get_root fails if the parse was ambiguous. Use get_roots() for ambiguous parses.
    #[allow(dead_code)]
    pub fn get_root(&self) -> Rc<BSR> {
        let rts = self.get_roots();
        if rts.len() != 1 {
            fail(format!("{} parse trees exist for start symbol {}", 
            rts.len(), self.start_sym))
        }
        return rts[0].clone()
    }

    // get_roots returns all the roots of parse trees of the start symbol of the grammar.
    #[allow(dead_code)]
    pub fn get_roots(&self) -> Vec<Rc<BSR>> {
        let mut roots: Vec<Rc<BSR>> = Vec::with_capacity(128);
        for b in self.slot_entries.keys() {
            if b.label.head() == &self.start_sym && b.lext == 0 && b.rext == self.rext {
                roots.push(b.clone())
            }
        }
        roots
    }

    /// Return the (line, column) of the left extent of token i.
    fn get_line_column(&self, i: usize) -> (usize, usize) {
    	return self.lex.get_line_column_of_token(i)
    }

    // get_nt_child_i returns the BSR of NT symbol[i] in the BSR set.
    // get_nt_child_i fails if the BSR set has ambiguous subtrees of NT i.
    #[allow(dead_code)]
    pub fn get_nt_child_i(&self, b: Rc<BSR>, i: usize) -> Rc<BSR> {
        let bsrs = self.get_nt_children_i(b.clone(), i);
        if bsrs.len() != 1 {
            panic!("NT {} is ambiguous in {}", i, b.clone());
        }
        return bsrs[0].clone()
    }

    // get_nt_children_i returns all the BSRs of NT symbol[i] in s
    #[allow(dead_code)]
    pub fn get_nt_children_i(&self, b: Rc<BSR>, i: usize) -> &Vec<Rc<BSR>> {
        if i >= b.label.symbols().len() {
            fail(format!("Error: cannot get NT child {} of {}", i, b))
        }
        if b.label.symbols().len() == 1 {
            return self.get_nt_slot(&b.label.symbols()[i], b.pivot, b.rext)
        }
        if b.label.symbols().len() == 2 {
            if i == 0 {
                return self.get_nt_slot(&b.label.symbols()[i], b.lext, b.pivot)
            }
            return self.get_nt_slot(&b.label.symbols()[i], b.pivot, b.rext)
        }
        let mut idx = b.label.index();
        let mut str_bsr = Rc::new(BSR{label: b.label, lext: b.lext, pivot: b.pivot, rext: b.rext});
        while idx.pos > i+1 && idx.pos > 2 {
            idx.pos -= 1;
            str_bsr = self.get_string(slot::get_label(&idx.nt, idx.alt, idx.pos), 
                str_bsr.lext, str_bsr.pivot);
        }
        if i == 0 {
            return self.get_nt_slot(&b.label.symbols()[i], str_bsr.lext, str_bsr.pivot)
        }
        return self.get_nt_slot(&b.label.symbols()[i], str_bsr.pivot, str_bsr.rext)
    }

    fn get_nt_slot(&self, sym: &Symbol, lext: usize, rext: usize) -> &Vec<Rc<BSR>> {
        if let Symbol::NT(nt) = sym {
            if let Some(bsrs) = self.nt_slot_entries.get(&NTSlot::new(&nt, lext, rext)) {
                return bsrs
            }
            panic!("{} ({},{}) has no slot entry", nt, lext, rext)
        }
        let (line, col) = self.get_line_column(lext);
        panic!("{} is not an NT at line {} col {}", sym, line, col);
    }
    
    fn get_string(&self, l: slot::Label, lext: usize, rext: usize) -> Rc<BSR> {
        for st in self.string_entries.keys() {
            if st.label == l && st.lext == lext && st.rext == rext {
                return st.clone()
            }
        }
        panic!("Error: no string BSR {} left extent={} right extent={} pos={}",
            symbols::to_string(l.symbols()), lext, rext, l.pos())
    }

    /**
    GetTChildI returns the terminal symbol at position i in b.   
    GetTChildI panics if symbol i is not a valid terminal
    */
    #[allow(dead_code)]
    pub fn get_t_child_i(&self, b: Rc<BSR>, i: usize) -> Rc<Token> {
		let symbols = b.label.symbols();

        if i >= symbols.len() {
            panic!("{} has no T child {}", b, i);
        }
        if symbols[i].is_nt() {
            panic!("symbol {} in {} is an NT", i, b);
		}
		
		let mut lext: usize = 0;
		for j in 0..i {
			if symbols[j].is_nt() {
				let nt = self.get_nt_child_i(b.clone(), j);
				lext += nt.rext - nt.lext;
			} else {
				lext += 1;
			}
		}

        self.lex.tokens[lext].clone()
    }

    /// Returns true if the BSR set does not have exactly one root, or
    /// if any BSR in the set has an NT symbol, which does not have exactly one
    /// sub-tree.
    #[allow(dead_code)]
    pub fn is_ambiguous(&self) -> bool {
        if self.get_roots().len() != 1 {
            return true
        }
        self.is_ambiguous_bsr(self.get_root())
    }

    /// Returns true if b or any of its NT children is ambiguous.
    /// A BSR is ambigous if any of its NT symbols does not have exactly one
    /// subtrees (children).
    fn is_ambiguous_bsr(&self, b: Rc<BSR>) -> bool {
        for (i, s) in b.label.symbols().iter().enumerate() {
            if s.is_nt() {
                if self.get_nt_children_i(b.clone(), i).len() != 1 {
                    return true
                }
                for b1 in self.get_nt_children_i(b.clone(), i).iter() {
                    if self.is_ambiguous_bsr(b1.clone()) {
                        return true
                    }
                }
            }
        }
        return false
    }

} // impl Set

impl Kind {
    fn rext(&self) -> usize {
        match self {
            Kind::NT(b) => b.rext,
            Kind::Str(b) => b.rext,
        }
    }
}

impl NTSlot {
    fn new(nt: &NT, lext: usize, rext: usize) -> NTSlot {
        NTSlot{
            nt: nt.clone(), 
            lext: lext,
            rext: rext,
        }
    }
}

impl fmt::Display for BSR {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "({} ({},{},{})", self.label, self.lext, self.pivot, self.rext)
    }
}

fn fail(msg: String) {
	panic!("Error in BSR: {}", msg)
}


