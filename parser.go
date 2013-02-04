package yail

// This file contains YAIL parser. It uses a lexer and produces a bytecode.

import (
	"fmt"
	"os"
	"strconv"
)

type parser struct {
	l   *lexer
	nxt []*lex
}

func parse(source string) function {
	p := parser{newLexer(source), make([]*lex, 0)}
	return p.fun(make(function, 0))
}

// f contains function loading function arguments
func (p *parser) fun(f function) function {
	endCriteria := lexEof
	if p.next(0).typ == lexLeftBrace {
		endCriteria = lexRightBrace
		p.skip(1)
	}
	for {
		switch l := p.get(); l.typ {
		case lexEos:
			continue
		case lexIf:
			f = p.pIf(f)
		case lexFor:
			f = p.pFor(f)
		case lexWhile:
			f = p.pWhile(f)
		case lexPrint:
			var n int
			f, n = p.callArgs(f)
			f = append(f, op{opPrint, n})
		case lexPrintLn:
			var n int
			f, n = p.callArgs(f)
			f = append(f, op{opPrintLn, n})
		case lexDot:
			fallthrough
		case lexName:
			f = p.name(f, l)
			switch n := p.get(); n.typ {
			case lexEq:
				f = p.assign(f)
			case lexLeftPar:
				f = p.call(f)
			default:
				parseErr("= or (", n)
			}
		case lexReturn:
			if nt := p.next(0).typ; nt != lexEos && nt != endCriteria {
				f = p.expr(f)
				f = append(f, op{opReturn, 1})
			} else {
				f = append(f, op{opReturn, 0})
			}
		case endCriteria:
			return f
		default:
			parseErr("if, for, while or name", l)
		}
		if l := p.get(); l.typ != lexEos && l.typ != endCriteria {
			parseErr("; or newline", l)
		} else if l.typ == endCriteria {
			return f
		}
	}
	return f
}

// already parsed name and =; name is on top of the stack
func (p *parser) assign(f function) function {
	if p.next(0).typ == lexLeftPar {
		switch p.next(1).typ {
		case lexRightPar:
			p.skip(2)
			if p.next(0).typ != lexLeftBrace {
				parseErr("{", p.next(0))
			}
			f = append(f, op{opFunction, p.fun(make(function, 0))})
		case lexName:
			if p.next(2).typ == lexComma || (p.next(2).typ == lexRightPar && p.next(3).typ == lexLeftBrace) {
				f = append(f, op{opFunction, p.fun(p.funArgs())})
			} else {
				f = p.expr(f)
			}
		default:
			f = p.expr(f)
		}
	} else {
		f = p.expr(f)
	}
	f = append(f, op{opStoreStr, nil})
	return f
}

// parses (arg, arg, arg...) and retuns opStore for each arg
func (p *parser) funArgs() function {
	p.skip(1) // already know that we have opLeftPar here
	ret := make(function, 0)
	for arg := p.get(); arg.typ != lexRightPar; arg = p.get() {
		if arg.typ != lexName {
			parseErr("name", arg)
		}
		ret = append(ret, op{opStore, arg.val})
		if p.next(0).typ == lexComma {
			p.skip(1)
		}
	}
	return ret
}

// parses (expr, expr, expr...), returns number of arguments
func (p *parser) callArgs(f function) (function, int) {
	var args int
	if l := p.get(); l.typ != lexLeftPar {
		parseErr("(", l)
	}
	for p.next(0).typ != lexRightPar {
		args++
		f = p.expr(f)
		if p.next(0).typ == lexComma {
			p.skip(1)
		} else if p.next(0).typ != lexRightPar {
			parseErr(", or )", p.next(0))
		}
	}
	p.skip(1)
	return f, args
}

var levelLex = [...][]lexType{{lexOr}, {lexAnd}, {lexEqEq, lexNeq},
	{lexLess, lexGreater, lexLeq, lexGeq}, {lexPlus, lexMinus},
	{lexMul, lexDiv, lexMod}}
var levelOp = [...][]opType{{opOr}, {opAnd}, {opEq, opNeq},
	{opLess, opGreater, opLeq, opGeq}, {opSum, opSub},
	{opMul, opDiv, opMod}}

func (p *parser) expr0(f function, level int) function {
	if level < len(levelLex) {
		f := p.expr0(f, level+1)
		lex := levelLex[level]
		opt := levelOp[level]
		found := true
		for found {
			found = false
			nt := p.next(0).typ
			for i, l := range lex {
				if l == nt {
					p.skip(1)
					f = p.expr0(f, level+1)
					f = append(f, op{opt[i], nil})
					found = true
					break
				}
			}
		}
		return f
	}
	switch l := p.get(); l.typ {
	case lexMinus:
		f = p.expr0(f, level)
		f = append(f, op{opNeg, nil})
	case lexNot:
		f = p.expr0(f, level)
		f = append(f, op{opNot, nil})
	case lexLeftPar:
		f = p.expr(f)
		if n := p.get(); n.typ != lexRightPar {
			parseErr(")", n)
		}
	case lexInt:
		i, err := strconv.ParseInt(l.val, 10, 64)
		if err != nil {
			parseErr("int", l)
		}
		f = append(f, op{opInt, i})
	case lexFloat:
		fl, err := strconv.ParseFloat(l.val, 64)
		if err != nil {
			parseErr("float", l)
		}
		f = append(f, op{opFloat, fl})
	case lexBool:
		b, err := strconv.ParseBool(l.val)
		if err != nil {
			parseErr("bool", l)
		}
		f = append(f, op{opBool, b})
	case lexString:
		s, err := strconv.Unquote(l.val)
		if err != nil {
			parseErr("string", l)
		}
		f = append(f, op{opString, s})
	case lexReadInt:
		f = append(f, op{opReadInt, nil})
	case lexReadFloat:
		f = append(f, op{opReadFloat, nil})
	case lexReadLine:
		f = append(f, op{opReadLine, nil})
	case lexReadChar:
		f = append(f, op{opReadChar, nil})
	case lexRnd:
		f = append(f, op{opRnd, nil})
	case lexDot:
		fallthrough
	case lexName:
		f = p.name(f, l)
		f = append(f, op{opLoadStr, nil})
		if p.next(0).typ == lexLeftPar { // function call
			var num int
			f, num = p.callArgs(f)
			f = append(f, op{opCall, num})
		}
	default:
		parseErr("int, float, string, bool, name or call", l)
	}
	return f
}

func (p *parser) name(f function, l *lex) function {
	n := l.val
	for l.typ == lexDot {
		l = p.get()
		n += l.val
	}
	if l.typ != lexName {
		parseErr("name or .", l)
	}
	f = append(f, op{opString, n})
	for p.next(0).typ == lexLeftBracket {
		p.skip(1)
		f = append(f, op{opString, "["})
		f = append(f, op{opSum, nil})
		f = p.expr(f)
		f = append(f, op{opSum, nil})
		if n := p.get(); n.typ != lexRightBracket {
			parseErr("[", n)
		}
		f = append(f, op{opString, "]"})
		f = append(f, op{opSum, nil})
	}
	return f
}

func (p *parser) expr(f function) function {
	return p.expr0(f, 0)
}

// already parsed name and (; name is a string in the stack
func (p *parser) call(f function) function {
	f = append(f, op{opLoadStr, nil})
	var args int
	for args = 0; p.next(0).typ != lexRightPar; args++ {
		f = p.expr(f)
		if p.next(0).typ == lexComma {
			p.skip(1)
		}
	}
	p.skip(1) // )
	f = append(f, op{opCall, args})
	f = append(f, op{opPop, nil}) // ignore return value
	return f
}

func (p *parser) pIf(f function) function {
	f = p.expr(f)
	if p.next(0).typ != lexLeftBrace {
		parseErr("{", p.next(0))
	}
	body := p.fun(make(function, 0))
	elseBody := make(function, 0)
	if p.next(0).typ == lexElse {
		p.skip(1)
		if p.next(0).typ != lexLeftBrace {
			parseErr("{", p.next(0))
		}
		elseBody = p.fun(elseBody)
		body = append(body, op{opJmp, len(elseBody) + 1})
	}
	f = append(f, op{opJmpFalse, len(body) + 1})
	f = append(f, body...)
	f = append(f, elseBody...)
	return f
}

func (p *parser) assignOrNone(f function) function {
	if l := p.next(0); l.val != ";" {
		f = p.name(f, p.get())
		if n := p.get(); n.typ != lexEq {
			parseErr("=", n)
		}
		f = p.assign(f)
	}
	return f
}

func (p *parser) semicolon() {
	if l := p.get(); l.val != ";" {
		parseErr(";", l)
	}
}

func (p *parser) pFor(f function) function {
	f = p.assignOrNone(f)
	p.semicolon()
	start := len(f)
	f = p.expr(f)
	p.semicolon()
	after := p.assignOrNone(make(function, 0))
	if p.next(0).typ != lexLeftBrace {
		parseErr("{", p.next(0))
	}
	body := p.fun(make(function, 0))
	f = append(f, op{opJmpFalse, len(body) + len(after) + 2})
	f = append(f, body...)
	f = append(f, after...)
	f = append(f, op{opJmp, start - len(f)})
	return f
}

func (p *parser) pWhile(f function) function {
	start := len(f)
	f = p.expr(f)
	if p.next(0).typ != lexLeftBrace {
		parseErr("{", p.next(0))
	}
	body := p.fun(make(function, 0))
	f = append(f, op{opJmpFalse, len(body) + 2})
	f = append(f, body...)
	f = append(f, op{opJmp, start - len(f)})
	return f
}

func parseErr(expected string, got *lex) {
	fmt.Printf("Parse error: expected %s, got %#v.\n", expected, *got)
	os.Exit(1)
}

func (p *parser) get() *lex {
	if len(p.nxt) == 0 {
		return p.l.get()
	}
	ret := p.nxt[0]
	p.nxt = p.nxt[1:]
	return ret
}

func (p *parser) next(i int) *lex {
	for len(p.nxt) <= i {
		p.nxt = append(p.nxt, p.l.get())
	}
	return p.nxt[i]
}

func (p *parser) skip(x int) {
	p.nxt = p.nxt[x:]
}
