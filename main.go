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

// at this stage chibicc has globals
var currentInput []rune

// codegen's depth state
var depth int = 0

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("%s: invalid number of args\n", os.Args[0])
	}

	// a unicode lexer
	currentInput = []rune(os.Args[1])

	// Tokenize  and parse.
	var tok *Token = tokenize()
	node := expr(&tok, tok)

	if tok.Kind != EOF {
		errorTok(tok, "extra token")
	}

	// Traverse the AST to emit assembly.
	codegen(node)
}
