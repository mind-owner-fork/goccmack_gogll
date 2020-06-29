//! Module ast is generated by GoGLL. Do not edit.

use crate::token;

use std::rc::Rc;

#[allow(dead_code)]
pub enum Node { 
    Lines(Vec<Box<Line>>),
    Line(Box<Line>),
    T(Rc<token::Token>),
    None,
}

pub struct Line {
    sap: String,
    ip: String,
    name1: String,
    name2: String,
    timestamp: String,
    string1: String,
    number1: String,
    number2: String,
    string2: String,
    string3: String,
}

#[allow(dead_code)]
pub enum NT { 
	Lines(), 
	Line, 
}

/// G0 : Lines ;
pub fn g_0_0(mut params: Vec<Node>) -> Result<Node, String> {
    Ok(params.remove(0))
}

/// Lines : Line ;
pub fn lines_0(mut params: Vec<Node>) -> Result<Node, String> {
    if let Node::Line(ln) = params.remove(0) {
        Ok(Node::Lines(vec![ln]))
    } else {
        panic!()
    }
}

/// Lines : Lines Line ;
pub fn lines_1(mut params: Vec<Node>) -> Result<Node, String> {
    let mut lns: Vec<Box<Line>> = if let Node::Lines(lns) = params.remove(0) {
        lns
    } else {
        panic!()
    };
    if let Node::Line(ln) = params.remove(0) {
        lns.push(ln)
    } else {
        panic!()
    };
    Ok(Node::Lines(lns))
}

/// Line : sap ip name name timestamp string number1 number1 string string ;
pub fn line_0(mut params: Vec<Node>) -> Result<Node, String> {
    let ln: Box<Line> = Box::new(Line{
        sap: get_literal_string(&params[0]),
        ip: get_literal_string(&params[1]),
        name1: get_literal_string(&params[2]),
        name2: get_literal_string(&params[3]),
        timestamp: get_literal_string(&params[4]),
        string1: get_literal_string(&params[5]),
        number1: get_literal_string(&params[6]),
        number2: get_literal_string(&params[7]),
        string2: get_literal_string(&params[8]),
        string3: get_literal_string(&params[9]),
    });
    Ok(Node::Line(ln))
}

// p is Node::T(Rc<token::Token>)
fn get_literal_string(p: &Node) -> String {
    if let Node::T(tok) = p {
        tok.literal_string()
    } else {
        panic!()
    }
}
