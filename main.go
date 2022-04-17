package main

import (
	"log"
	"os"
)

func assert(ok bool) {
	if !ok {
		panic("FAIL")
	}
}

// at this stage chibicc has the following globals
//
// want them to be as visible as possible,
// hence, before the main()
var currentInput []rune

// codegen's depth state
var depth int = 0

// All local variable instances created during parsing are
// accumulated to this list
var locals *Obj

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("%s: invalid number of args\n", os.Args[0])
	}

	// a unicode lexer
	currentInput = []rune(os.Args[1])

	// Tokenize  and parse.
	var tok *Token = tokenize()
	//log.Println(tok.String())
	var prog *Function = parse(tok)
	//log.Println(prog)

	// Traverse the AST to emit assembly.
	codegen(prog)
}
