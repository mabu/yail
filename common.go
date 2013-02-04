package yail

// Lexeme.
type lex struct {
	typ lexType
	val string
}

type lexType int

const (
	lexemeBuffer            = 10
	eof                     = 0
	lexError        lexType = iota
	lexLeftPar              // (
	lexRightPar             // )
	lexLeftBrace            // {
	lexRightBrace           // }
	lexMinus                // -
	lexPlus                 // +
	lexMul                  // *
	lexDiv                  // /
	lexMod                  // %
	lexComma                // ,
	lexDot                  // .
	lexLeftBracket          // [
	lexRightBracket         // ]
	lexEos                  // ; or \n
	lexOr                   // ||
	lexAnd                  // &&
	lexNot                  // !
	lexEq                   // =
	lexEqEq                 // ==
	lexNeq                  // !=
	lexGreater              // >
	lexLess                 // <
	lexGeq                  // >=
	lexLeq                  // <=
	lexIf
	lexElse
	lexFor
	lexWhile
	lexReturn
	lexBool
	lexInt
	lexFloat
	lexString
	lexName
	lexReadInt    // @int
	lexReadFloat  // @float
	lexReadString // @string
	lexReadLine   // @line
	lexReadChar   // @char
	lexPrint      // @print
	lexPrintLn    // @println
	lexRnd        // @rnd
	lexEof
)

type opType int

const (
	opCall     opType = iota // calls a function; param: number of arguments int
	opJmp                    // execution jump; param: diff int
	opJmpFalse               // jump if false; param: diff int
	opLoad                   // load constant from variable; param: variable name string
	opStore                  // store constant to variable; param: variable name string
	opLoadStr                // load constant from variable; variable name is a string on the stack
	opStoreStr               // store constant to variable; constant is on the top of the stack, variable name is the next string
	opSum                    // sums (or concatenates) two constants
	opSub                    // subtracts two numbers
	opMul                    // multiplies two numbers
	opDiv                    // divides two numbers
	opMod                    // calculates modulo of two integers
	opNeg                    // negates a number
	opOr                     // logical or
	opAnd                    // logical and
	opNot                    // logical not
	opEq                     // ==
	opNeq                    // !=
	opLess                   // <
	opGreater                // >
	opLeq                    // <=
	opGeq                    // >=
	opInt                    // constant; param: int64
	opFloat                  // constant; param: float64
	opBool                   // constant; param: bool
	opString                 // constant; param: string
	opFunction               // constant; param: function
	opReadInt
	opReadFloat
	opReadString
	opReadLine
	opReadChar
	opPrint   // prints space separated values; param: number of values int
	opPrintLn // prints space separated values, appends newline; param: number of values int
	opRnd     // puts random int to the stack
	opPop     // discard the top element of the stack
	opReturn  // returns from the function
)

type function []op

type variable struct {
}

type op struct {
	typ   opType
	param interface{}
}
