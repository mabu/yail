package yail

// This file contains a stack-based YAIL interpreter.

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type interpreter struct {
	stdin  io.Reader
	stdout io.Writer
	f      function
	vars   map[string]interface{}
	stack  []interface{}
	parent *interpreter
}

func Interpret(source string, input io.Reader, output io.Writer) {
	rand.Seed(time.Now().UTC().UnixNano())
	i := newInterpreter(parse(source), input, output)
	i.run()
}

func newInterpreter(f function, input io.Reader, output io.Writer) *interpreter {
	return &interpreter{input, output, f, make(map[string]interface{}), make([]interface{}, 0), nil}
}

func (i *interpreter) run() interface{} {
	for ic := 0; ic < len(i.f); {
		op := (i.f)[ic]
		switch op.typ {
		case opInt:
			fallthrough
		case opFloat:
			fallthrough
		case opBool:
			fallthrough
		case opString:
			fallthrough
		case opFunction:
			i.push(op.param)
		case opCall:
			args := getInt(op.param, "opCall failed: number of arguments is not int")
			child := newInterpreter(nil, i.stdin, i.stdout)
			child.parent = i
			for args > 0 { // order reversal is intended
				child.push(i.pop())
				args--
			}
			f, ok := i.pop().(function)
			if !ok {
				runtimeErr("opCall failed: function not found")
				return nil
			}
			child.f = f
			i.push(child.run())
		case opJmp:
			ic += getInt(op.param, "opJmp failed") - 1
		case opJmpFalse:
			if !i.popBool("opJmpFalse failed: non-bool") {
				ic += getInt(op.param, "opJmpFalse failed: non-int param") - 1
			}
		case opLoad:
			name0 := getString(op.param, "opLoad failed: non-string param")
			interpr, name := trimDots(i, name0)
			val, ok := interpr.vars[name]
			if !ok {
				runtimeErr("opLoad failed: variable " + name0 + " is undefined")
			}
			i.push(val)
		case opStore:
			name0 := getString(op.param, "opStore failed: non-string param")
			interpr, name := trimDots(i, name0)
			interpr.vars[name] = i.pop()
		case opLoadStr:
			name0 := getString(i.pop(), "opLoadStr failed: non-string name")
			interpr, name := trimDots(i, name0)
			val, ok := interpr.vars[name]
			if !ok {
				runtimeErr("opLoadStr failed: variable " + name0 + " is undefined")
			}
			i.push(val)
		case opStoreStr:
			val := i.pop()
			name0 := getString(i.pop(), "opStoreStr failed: non-string name")
			interpr, name := trimDots(i, name0)
			interpr.vars[name] = val
		case opOr:
			i.push(i.popBool("opOr failed") || i.popBool("opOr failed"))
		case opAnd:
			i.push(i.popBool("opAnd failed") && i.popBool("opAnd failed"))
		case opNot:
			i.push(!i.popBool("opNot failed"))
		case opEq:
			i.cmpOp("opEq", func(a, b int64) bool { return a == b },
				func(a, b float64) bool { return a == b },
				func(a, b bool) bool { return a == b },
				func(a, b string) bool { return a == b })
		case opNeq:
			i.cmpOp("opNeq", func(a, b int64) bool { return a != b },
				func(a, b float64) bool { return a != b },
				func(a, b bool) bool { return a != b },
				func(a, b string) bool { return a != b })
		case opLess:
			i.cmpOp("opLess", func(a, b int64) bool { return a < b },
				func(a, b float64) bool { return a < b },
				func(a, b bool) bool { runtimeErr("opLess not defined on bool"); return false },
				func(a, b string) bool { return a < b })
		case opGreater:
			i.cmpOp("opGreater", func(a, b int64) bool { return a > b },
				func(a, b float64) bool { return a > b },
				func(a, b bool) bool { runtimeErr("opGreater not defined on bool"); return false },
				func(a, b string) bool { return a > b })
		case opLeq:
			i.cmpOp("opLeq", func(a, b int64) bool { return a <= b },
				func(a, b float64) bool { return a <= b },
				func(a, b bool) bool { runtimeErr("opLeq not defined on bool"); return false },
				func(a, b string) bool { return a <= b })
		case opGeq:
			i.cmpOp("opGeq", func(a, b int64) bool { return a >= b },
				func(a, b float64) bool { return a >= b },
				func(a, b bool) bool { runtimeErr("opGeq not defined on bool"); return false },
				func(a, b string) bool { return a >= b })
		case opSum:
			switch s1 := i.pop().(type) {
			case int64:
				switch s2 := i.pop().(type) {
				case int64:
					i.push(s2 + s1)
				case float64:
					i.push(s2 + float64(s1))
				case string:
					i.push(s2 + itoa(s1))
				default:
					runtimeErr("opSum failed: could not add int and unknown")
				}
			case float64:
				switch s2 := i.pop().(type) {
				case int64:
					i.push(float64(s2) + s1)
				case float64:
					i.push(s2 + s1)
				case string:
					i.push(s2 + ftoa(s1))
				default:
					runtimeErr("opSum failed: could not add float and unknown")
				}
			case string:
				switch s2 := i.pop().(type) {
				case int64:
					i.push(itoa(s2) + s1)
				case float64:
					i.push(ftoa(s2) + s1)
				case string:
					i.push(s2 + s1)
				default:
					runtimeErr("opSum failed: could not add string and unknown")
				}
			default:
				runtimeErr("opSum failed: unknown type")
			}
		case opSub:
			i.numberOp("opSub", func(a, b int64) int64 { return a - b },
				func(a, b float64) float64 { return a - b })
		case opNeg:
			i.push(-1)
			fallthrough
		case opMul:
			i.numberOp("opMul", func(a, b int64) int64 { return a * b },
				func(a, b float64) float64 { return a * b })
		case opDiv:
			i.numberOp("opDiv", func(a, b int64) int64 { return a / b },
				func(a, b float64) float64 { return a / b })
		case opMod:
			i.numberOp("opMod", func(a, b int64) int64 { return a % b },
				func(a, b float64) float64 { runtimeErr("Operator % not defined on float"); return 0 })
		case opReadInt:
			var val int64
			i.read(&val, "%d")
			i.push(val)
		case opReadFloat:
			var val float64
			i.read(&val, "%f")
			i.push(val)
		case opReadString:
			var val string
			i.read(&val, "%s")
			i.push(val)
		case opReadLine:
			val, err := bufio.NewReader(i.stdin).ReadString('\n')
			if err != nil {
				runtimeErr("opReadLine failed: " + err.Error())
			}
			i.push(val)
		case opReadChar:
			var val rune
			i.read(&val, "%c")
			i.push(string(val))
		case opRnd:
			i.push(rand.Int63())
		case opPrint:
			n := getInt(op.param, "opPrint failed: param not int")
			if n > len(i.stack) {
				runtimeErr("opPrint failed: not enough parameters on the stack")
			}
			fmt.Fprint(i.stdout, i.stack[len(i.stack)-n:]...)
			i.stack = i.stack[:len(i.stack)-n]
		case opPrintLn:
			n := getInt(op.param, "opPrintLn failed: param not int")
			if n > len(i.stack) {
				runtimeErr("opPrintLn failed: not enough parameters on the stack")
			}
			fmt.Fprintln(i.stdout, i.stack[len(i.stack)-n:]...)
			i.stack = i.stack[:len(i.stack)-n]
		case opPop:
			i.pop()
		case opReturn:
			if getInt(op.param, "opReturn failed: param not int") == 0 {
				return nil
			}
			return i.pop()
		}
		ic++
	}
	return nil
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

type intF func(a, b int64) int64
type floatF func(a, b float64) float64
type intFbool func(a, b int64) bool
type floatFbool func(a, b float64) bool
type boolF func(a, b bool) bool
type stringFbool func(a, b string) bool

func (i *interpreter) read(val interface{}, format string) {
	if n, err := fmt.Fscanf(i.stdin, format, val); n == 0 || err != nil {
		runtimeErr("read failed: " + err.Error())
	}
}

func (i *interpreter) numberOp(name string, fi intF, ff floatF) {
	switch s1 := i.pop().(type) {
	case int64:
		switch s2 := i.pop().(type) {
		case int64:
			i.push(fi(s2, s1))
		case float64:
			i.push(ff(s2, float64(s1)))
		default:
			runtimeErr(name + " failed: wrong type of second argument")
		}
	case float64:
		switch s2 := i.pop().(type) {
		case int64:
			i.push(ff(float64(s2), s1))
		case float64:
			i.push(ff(s2, s1))
		default:
			runtimeErr(name + " failed: wrong type of second argument")
		}
	default:
		runtimeErr(name + " failed: wrong type of first argument")
	}
}

func (i *interpreter) cmpOp(name string, fi intFbool, ff floatFbool, fb boolF, fs stringFbool) {
	switch s1 := i.pop().(type) {
	case int64:
		switch s2 := i.pop().(type) {
		case int64:
			i.push(fi(s2, s1))
		case float64:
			i.push(ff(s2, float64(s1)))
		default:
			runtimeErr(name + " failed: wrong type of second argument")
		}
	case float64:
		switch s2 := i.pop().(type) {
		case int64:
			i.push(ff(float64(s2), s1))
		case float64:
			i.push(ff(s2, s1))
		default:
			runtimeErr(name + " failed: wrong type of second argument")
		}
	case bool:
		if s2, ok := i.pop().(bool); ok {
			i.push(fb(s2, s1))
		} else {
			runtimeErr(name + " failed: second argument not bool")
		}
	case string:
		if s2, ok := i.pop().(string); ok {
			i.push(fs(s2, s1))
		} else {
			runtimeErr(name + " failed: second argument not string")
		}
	default:
		runtimeErr(name + " failed: wrong type of first argument")
	}
}

func trimDots(i0 *interpreter, name0 string) (i *interpreter, name string) {
	i = i0
	for name = name0; len(name) > 0 && name[0] == '.'; name = name[1:] {
		i = i.parent
		if i == nil {
			runtimeErr("too many dots in variable name: " + name0)
		}
	}
	if len(name) == 0 {
		runtimeErr("variable name is empty")
	}
	return
}

func (i *interpreter) push(val interface{}) {
	i.stack = append(i.stack, val)
}

func (i *interpreter) pop() interface{} {
	l := len(i.stack) - 1
	if l < 0 {
		runtimeErr("Stack underflow")
	}
	ret := i.stack[l]
	i.stack = i.stack[:l]
	return ret
}

func (i *interpreter) popBool(err string) bool {
	ret, ok := i.pop().(bool)
	if !ok {
		runtimeErr(err)
	}
	return ret
}

func getInt(i interface{}, err string) int {
	ret, ok := i.(int)
	if !ok {
		runtimeErr(err)
	}
	return ret
}

func getString(i interface{}, err string) string {
	ret, ok := i.(string)
	if !ok {
		runtimeErr(err)
	}
	return ret
}

func runtimeErr(str string) {
	fmt.Printf("Runtime error: %s.\n", str)
	os.Exit(0)
}
