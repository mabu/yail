// YAIL - Yet Another Interpreted Language
package yail

// This file contains a lexical analyser.

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var runeLexeme = map[rune]lexType{
	'(':  lexLeftPar,
	')':  lexRightPar,
	'{':  lexLeftBrace,
	'}':  lexRightBrace,
	'-':  lexMinus,
	'+':  lexPlus,
	'*':  lexMul,
	'%':  lexMod,
	',':  lexComma,
	'.':  lexDot,
	'[':  lexLeftBracket,
	']':  lexRightBracket,
	';':  lexEos,
	'\n': lexEos,
}

// conflicts: Eq-EqEq, Div-comment, Less-Leq, Greater-Geq, Not-Neq
var twoRunesLexeme = map[string]lexType{
	"||": lexOr,
	"&&": lexAnd,
	"==": lexEqEq,
	"!=": lexNeq,
	">=": lexGeq,
	"<=": lexLeq,
}

var stringLexeme = map[string]lexType{
	"true":     lexBool,
	"false":    lexBool,
	"if":       lexIf,
	"else":     lexElse,
	"for":      lexFor,
	"while":    lexWhile,
	"return":   lexReturn,
	"@int":     lexReadInt,
	"@float":   lexReadFloat,
	"@line":    lexReadLine,
	"@char":    lexReadChar,
	"@print":   lexPrint,
	"@println": lexPrintLn,
	"@rnd":     lexRnd,
}

var conflictingRuneLexeme = map[rune]lexType{
	'=': lexEq,
	'/': lexDiv,
	'<': lexLess,
	'>': lexGreater,
	'!': lexNot,
}

type lexer struct {
	input   string
	pos     int // start position of the current lexeme
	lexemes chan *lex
}

func newLexer(input string) *lexer {
	l := lexer{input: input, lexemes: make(chan *lex, lexemeBuffer)}
	go l.run()
	return &l
}

func (l *lexer) run() {
	defer close(l.lexemes)
Start:
	for r, s := l.next(); r != eof; r, s = l.next() {
		if lex, ok := runeLexeme[r]; ok {
			l.emit(lex, s)
			continue
		}
		if unicode.IsSpace(r) {
			l.pos += s
			continue
		}
		if r == '/' {
			l.pos += s
			switch r2, s2 := l.next(); r2 {
			case '/': // single line comment
				l.pos += s2
				l.skipUntil("\n")
				l.emit(lexEos, 0)
				continue Start
			case '*': // comment
				l.pos += s2
				l.skipUntil("*/")
				continue Start
			default:
				l.pos -= s
			}
		}
		for str, lex := range twoRunesLexeme {
			if strings.HasPrefix(l.input[l.pos:], str) {
				l.emit(lex, len(str))
				continue Start
			}
		}
		for str, lex := range stringLexeme {
			if strings.HasPrefix(l.input[l.pos:], str) {
				r, _ := utf8.DecodeRuneInString(l.input[l.pos+len(str):])
				if !isAlphaNum(r) {
					l.emit(lex, len(str))
					continue Start
				}
			}
		}
		if lex, ok := conflictingRuneLexeme[r]; ok {
			l.emit(lex, s)
			continue
		}
		if unicode.IsNumber(r) {
			i := strings.IndexFunc(l.input[l.pos+s:], notNumber)
			if i == -1 {
				l.emit(lexInt, len(l.input)-l.pos)
			} else if l.input[l.pos+s+i] == '.' {
				i++
				j := strings.IndexFunc(l.input[l.pos+s+i:], notNumber)
				if j == -1 {
					l.emit(lexFloat, len(l.input)-l.pos)
				} else {
					l.emit(lexFloat, s+i+j)
				}
			} else {
				l.emit(lexInt, s+i)
			}
		} else if r == '"' { // string
			i := strings.Index(l.input[l.pos+s:], "\"")
			for i != -1 {
				if (i-strings.LastIndexFunc(l.input[l.pos+s-1:l.pos+s+i], func(r rune) bool { return r != '\\' }))%2 == 0 {
					break
				}
				s += i + 1
				i = strings.Index(l.input[l.pos+s:], "\"")
			}
			if i == -1 {
				l.emit(lexError, len(l.input)-l.pos)
			} else {
				l.emit(lexString, s+i+1)
			}
		} else if unicode.IsLetter(r) {
			i := strings.IndexFunc(l.input[l.pos+s:], func(r rune) bool { return !isAlphaNum(r) })
			if i == -1 {
				s = len(l.input) - l.pos
			} else {
				s += i
			}
			l.emit(lexName, s)
		} else {
			l.emit(lexError, s)
			return
		}
	}
	l.emit(lexEof, 0)
}

func notNumber(r rune) bool { return !unicode.IsNumber(r) }

func (l *lexer) get() *lex {
	return <-l.lexemes
}

func (l *lexer) emit(t lexType, size int) {
	l.lexemes <- &lex{typ: t, val: l.input[l.pos : l.pos+size]}
	l.pos += size
}

func isAlphaNum(r rune) bool {
	return unicode.IsOneOf([]*unicode.RangeTable{unicode.Letter, unicode.Number}, r)
}

func (l *lexer) next() (r rune, size int) {
	if l.pos >= len(l.input) {
		return eof, 0
	}
	r, size = utf8.DecodeRuneInString(l.input[l.pos:])
	if r == utf8.RuneError {
		panic("utf8.RuneError")
	}
	return
}

func (l *lexer) skipUntil(s string) {
	if skip := strings.Index(l.input[l.pos:], s); skip == -1 {
		l.pos = len(l.input)
	} else {
		l.pos += skip + len(s)
	}
}
