package yail

import "testing"

func TestParenthesis(t *testing.T) {
	runLexTest(t, "()", []lex{{lexLeftPar, "("}, {lexRightPar, ")"}})
}

func TestBraces(t *testing.T) {
	runLexTest(t, "}{", []lex{{lexRightBrace, "}"}, {lexLeftBrace, "{"}})
}

func TestArithmetic(t *testing.T) {
	runLexTest(t, "-+*/%", []lex{{lexMinus, "-"}, {lexPlus, "+"}, {lexMul, "*"}, {lexDiv, "/"}, {lexMod, "%"}})
}

func TestPunctuation(t *testing.T) {
	runLexTest(t, ",.", []lex{{lexComma, ","}, {lexDot, "."}})
}

func TestEos(t *testing.T) {
	runLexTest(t, ";\n", []lex{{lexEos, ";"}, {lexEos, "\n"}})
}

func TestLogic(t *testing.T) {
	runLexTest(t, "!||&&", []lex{{lexNot, "!"}, {lexOr, "||"}, {lexAnd, "&&"}})
}

func TestCompareOp(t *testing.T) {
	runLexTest(t, "< <= >= > != ==", []lex{{lexLess, "<"}, {lexLeq, "<="}, {lexGeq, ">="}, {lexGreater, ">"}, {lexNeq, "!="}, {lexEqEq, "=="}})
}

func TestName(t *testing.T) {
	runLexTest(t, "xe123b[58]", []lex{{lexName, "xe123b"}, {lexLeftBracket, "["}, {lexInt, "58"}, {lexRightBracket, "]"}})
}

func TestFloatAssign(t *testing.T) {
	runLexTest(t, "f = 0.543 -0.234 16.", []lex{{lexName, "f"}, {lexEq, "="}, {lexFloat, "0.543"}, {lexMinus, "-"}, {lexFloat, "0.234"}, {lexFloat, "16."}})
}

func TestKeyword(t *testing.T) {
	runLexTest(t, "if ifa while whileb for for3 returni return", []lex{{lexIf, "if"}, {lexName, "ifa"}, {lexWhile, "while"}, {lexName, "whileb"}, {lexFor, "for"}, {lexName, "for3"}, {lexName, "returni"}, {lexReturn, "return"}})
}

func TestBool(t *testing.T) {
	runLexTest(t, "true1 true false falseb", []lex{{lexName, "true1"}, {lexBool, "true"}, {lexBool, "false"}, {lexName, "falseb"}})
}

func TestString(t *testing.T) {
	runLexTest(t, `"lorem \\ ipsum šlept\\\n \\\" \"foo\"" bar ""`, []lex{{lexString, `"lorem \\ ipsum šlept\\\n \\\" \"foo\""`}, {lexName, "bar"}, {lexString, `""`}})
}

func runLexTest(t *testing.T, input string, expect []lex) {
	lexer := newLexer(input)
	for _, l := range expect {
		if got := lexer.get(); l != *got {
			t.Errorf("Got %v, expected %v.", got, l)
		}
	}
	if got := lexer.get(); *got != (lex{lexEof, ""}) {
		t.Errorf("Expected EOF, got %v.", got)
	}
}
