package yail

import "testing"

func TestEmpty(t *testing.T) {
	runParseTest(t, "", function{})
}

func TestAssign(t *testing.T) {
	runParseTest(t, "a = 5", function{op{opString, "a"}, op{opInt, int64(5)}, op{opStoreStr, nil}})
}

func TestCall(t *testing.T) {
	runParseTest(t, "blac()", function{op{opString, "blac"}, op{opLoadStr, nil}, op{opCall, int(0)}, op{opPop, nil}})
}

func TestParseName(t *testing.T) {
	runParseTest(t, "..name[5] = b[a]", function{
		op{opString, "..name"},
		op{opString, "["},
		op{opSum, nil},
		op{opInt, int64(5)},
		op{opSum, nil},
		op{opString, "]"},
		op{opSum, nil},
		op{opString, "b"},
		op{opString, "["},
		op{opSum, nil},
		op{opString, "a"},
		op{opLoadStr, nil},
		op{opSum, nil},
		op{opString, "]"},
		op{opSum, nil},
		op{opLoadStr, nil},
		op{opStoreStr, nil}})
}

func TestNumExpr(t *testing.T) {
	runParseTest(t, "a = 5 + 4 * (3 - 7.2 / 2)", function{
		op{opString, "a"},
		op{opInt, int64(5)},
		op{opInt, int64(4)},
		op{opInt, int64(3)},
		op{opFloat, float64(7.2)},
		op{opInt, int64(2)},
		op{opDiv, nil},
		op{opSub, nil},
		op{opMul, nil},
		op{opSum, nil},
		op{opStoreStr, nil}})
}

func TestBoolExpr(t *testing.T) {
	f := function{op{opString, "a"},
		op{opString, "b"},
		op{opLoadStr, nil},
		op{opString, "x"},
		op{opLoadStr, nil},
		op{opInt, int64(4)},
		op{opFloat, float64(3.5)},
		op{opGreater, nil},
		op{opAnd, nil},
		op{opOr, nil},
		op{opStoreStr, nil}}
	runParseTest(t, "a = b || x && 4 > 3.5", f)
	runParseTest(t, "a = (b || (x && (4 > 3.5)))", f)
	runParseTest(t, "a = (b || x) && (4 > 3.5)", function{op{opString, "a"},
		op{opString, "b"},
		op{opLoadStr, nil},
		op{opString, "x"},
		op{opLoadStr, nil},
		op{opOr, nil},
		op{opInt, int64(4)},
		op{opFloat, float64(3.5)},
		op{opGreater, nil},
		op{opAnd, nil},
		op{opStoreStr, nil}})
}

func TestIf(t *testing.T) {
	runParseTest(t, "if a <= 5 { b = 4 } else { x = c }", function{
		op{opString, "a"},
		op{opLoadStr, nil},
		op{opInt, int64(5)},
		op{opLeq, nil},
		op{opJmpFalse, int(5)},
		op{opString, "b"},
		op{opInt, int64(4)},
		op{opStoreStr, nil},
		op{opJmp, int(5)},
		op{opString, "x"},
		op{opString, "c"},
		op{opLoadStr, nil},
		op{opStoreStr, nil}})
}

func TestForPrint(t *testing.T) {
	f := function{
		op{opString, "i"},
		op{opInt, int64(0)},
		op{opStoreStr, nil},
		op{opString, "i"},
		op{opLoadStr, nil},
		op{opInt, int64(10)},
		op{opLess, nil},
		op{opJmpFalse, 15},
		op{opString, "i"},
		op{opLoadStr, nil},
		op{opString, "i"},
		op{opLoadStr, nil},
		op{opInt, int64(1)},
		op{opSum, nil},
		op{opPrintLn, 2},
		op{opString, "i"},
		op{opString, "i"},
		op{opLoadStr, nil},
		op{opInt, int64(1)},
		op{opSum, nil},
		op{opStoreStr, nil},
		op{opJmp, -18},
	}
	runParseTest(t, "for i = 0; i < 10; i = i + 1 { @println(i, i + 1) }", f)
	runParseTest(t, `for i = 0; i < 10; i = i + 1 {
		@println(i, i + 1)
	}`, f)
}

func TestWhileRead(t *testing.T) {
	runParseTest(t, `while @int > 0	{
	@println("still positive")
}
	@println("end")`, function{
		op{opReadInt, nil},
		op{opInt, int64(0)},
		op{opGreater, nil},
		op{opJmpFalse, 4},
		op{opString, "still positive"},
		op{opPrintLn, 1},
		op{opJmp, -6},
		op{opString, "end"},
		op{opPrintLn, 1},
	})
}

func TestFunctions(t *testing.T) {
	runParseTest(t, `fun = (x) {
		if x == 0 {
			return 1
		} else {
			return x * .fun(x - 1)
		}
	}
	@print(fun(5))`, function{
		op{opString, "fun"},
		op{opFunction, function{
			op{opStore, "x"},
			op{opString, "x"},
			op{opLoadStr, nil},
			op{opInt, int64(0)},
			op{opEq, nil},
			op{opJmpFalse, 4},
			op{opInt, int64(1)},
			op{opReturn, 1},
			op{opJmp, 12},
			op{opString, "x"},
			op{opLoadStr, nil},
			op{opString, ".fun"},
			op{opLoadStr, nil},
			op{opString, "x"},
			op{opLoadStr, nil},
			op{opInt, int64(1)},
			op{opSub, nil},
			op{opCall, 1},
			op{opMul, nil},
			op{opReturn, 1},
		}},
		op{opStoreStr, nil},
		op{opString, "fun"},
		op{opLoadStr, nil},
		op{opInt, int64(5)},
		op{opCall, 1},
		op{opPrint, 1},
	})
}

func runParseTest(t *testing.T, source string, expect function) {
	testFun(t, parse(source), expect)
}

func testFun(t *testing.T, f, expect function) {
	if len(f) < len(expect) {
		t.Fatalf("Result too short: %d < %d.\nGot: %#v\nExpected: %#v\n",
			len(f), len(expect), f, expect)
	}
	for i, op := range expect {
		if op.typ == opFunction {
			if f[i].typ != opFunction {
				t.Errorf("On position %d got %#v, expected %#v.\n", i, f[i], op)
			}
			testFun(t, f[i].param.(function), op.param.(function))
			continue
		}
		if op != f[i] {
			t.Errorf("On position %d got %#v, expected %#v.", i, f[i], op)
		}
	}
	if len(f) > len(expect) {
		t.Errorf("Result too long: %d < %d.\nGot: %#v\nExpected: %#v\n",
			len(f), len(expect), f, expect)
	}
}
