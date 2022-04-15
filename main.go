package main

import (
	"fmt"
	"log"
	"os"
	"unicode"
)

// helpers

// base10
func strtol(s *[]rune) int64 {

	var val int64 = 0

	for i, r := range *s {
		if !unicode.IsDigit(r) {
			*s = (*s)[i:]
			return val
		}

		digit := int64(r) - 48
		val = 10*val + digit
	}

	*s = []rune("")
	return val
}

func assert(ok bool) {
	if !ok {
		panic("FAIL")
	}
}

//
// Tokenizer
//

// token data type
// it is implemented as a linked list

type TokenKind int

const (
	PUNCT TokenKind = iota // punctuators
	NUM                    // numeric literals
	EOF                    // end-of-file markers
)

type Token struct {
	Kind TokenKind // token kind
	Next *Token    // next token
	val  int       // if kind is NUM, its value
	loc  []rune    // the rune slice, underlying the the token val
	// not needed, use len(loc)
	// len  int
}

// Consumes the current token, if it matches
func (t *Token) equal(s string) bool {
	return string(t.loc) == s
}

func NewToken(kind TokenKind, text []rune) *Token {
	t := new(Token)
	t.Kind = kind
	t.loc = text
	return t
}

// ensure token->Kind == NUM
func (t *Token) number() int {
	if t.Kind != NUM {
		errorTok(t, "expected a number")
	}
	return t.val
}

// Ensure that the current token is `s`
func skip(t *Token, s string) *Token {
	if !t.equal(s) {
		errorTok(t, "expected '%s'", s)
	}
	return t.Next
}

// Tokenize `p` and returns new tokens.
func tokenize() *Token {
	p := currentInput
	// start node of the linked list
	head := new(Token)
	cur := head

	for len(p) > 0 {

		// skip whitespace characters
		if unicode.IsSpace(p[0]) {
			p = p[1:]
			continue
		}

		// numeric literal
		if unicode.IsDigit(p[0]) {
			q := p
			num := strtol(&p)
			l := len(q) - len(p)

			cur.Next = NewToken(NUM, q[0:l])
			cur = cur.Next
			cur.val = int(num)

			continue
		}

		if unicode.IsPunct(p[0]) ||
			p[0] == '+' || p[0] == '-' {
			cur.Next = NewToken(PUNCT, p[0:1])
			cur = cur.Next
			p = p[1:]
			continue
		}

		errorAt(p, "invalid token")
	}

	cur, cur.Next = cur.Next, NewToken(EOF, p)
	return head.Next
}

//
// Parser
//

type NodeKind int

const (
	ND_ADD NodeKind = iota // +
	ND_SUB                 // -
	ND_MUL                 // *
	ND_DIV                 // /
	ND_NUM                 // Integer
)

// AST node type
type Node struct {
	kind NodeKind // node kind
	lhs  *Node    // left hand side
	rhs  *Node    // right hand side
	val  int      // Used if kind == ND_NUM
}

func NewNode(kind NodeKind) *Node {
	node := new(Node)
	node.kind = kind

	return node
}

func NewBinary(kind NodeKind, lhs *Node, rhs *Node) *Node {
	node := NewNode(kind)
	node.lhs = lhs
	node.rhs = rhs

	return node
}

func NewNum(val int) *Node {
	node := NewNode(ND_NUM)
	node.val = val

	return node
}

// expr = mul ("+" mul | "-" mul)*
func expr(rest **Token, tok *Token) *Node {
	node := mul(&tok, tok)

	for {
		if tok.equal("+") {
			node = NewBinary(ND_ADD, node, mul(&tok, tok.Next))
			continue
		}

		if tok.equal("-") {
			node = NewBinary(ND_SUB, node, mul(&tok, tok.Next))
			continue
		}

		*rest = tok
		return node
	}
}

// mul = primary ("*" primary | "/" primary)*
func mul(rest **Token, tok *Token) *Node {
	node := primary(&tok, tok) // left node for the new binary node

	for {
		if tok.equal("*") {
			// rhs is primary(&tok,.)
			node = NewBinary(ND_MUL, node, primary(&tok, tok.Next))
			continue
		}

		if tok.equal("/") {
			node = NewBinary(ND_DIV, node, primary(&tok, tok.Next))
			continue
		}

		*rest = tok
		return node
	}
}

// primary = "(" expr ")" | num
func primary(rest **Token, tok *Token) *Node {
	if tok.equal("(") {
		node := expr(&tok, tok.Next)
		*rest = skip(tok, ")")
		return node
	}

	if tok.Kind == NUM {
		node := NewNum(tok.val)
		*rest = tok.Next
		return node
	}

	errorTok(tok, "expected an expression")
	return nil
}

//
// Code generator
//

var depth int = 0

func push() {
	fmt.Println(" push %rax")
	depth++
}

func pop(arg string) {
	fmt.Printf(" pop %s\n", arg)
	depth--
}

func genExpr(node *Node) {
	if node.kind == ND_NUM {
		fmt.Printf(" mov $%d, %%rax\n", node.val)
		return
	}

	genExpr(node.rhs)
	push()
	genExpr(node.lhs)
	pop("%rdi")

	switch node.kind {
	case ND_ADD:
		fmt.Println(" add %rdi, %rax")
		return
	case ND_SUB:
		fmt.Println(" sub %rdi, %rax")
		return
	case ND_MUL:
		fmt.Println(" imul %rdi, %rax")
		return
	case ND_DIV:
		fmt.Println(" cqo")
		fmt.Println(" idiv %rdi, %rax")
		return
	}

	log.Fatalln("invalid expression")
}

// walks and prints the linked list
func (t *Token) String() string {
	s := ""
	n := new(Token)
	*n = *t
	for ; n != nil; n = n.Next {
		s += fmt.Sprintln("[rune loc: ", (n.loc), " kind: ", n.Kind, " val:", n.val, "],")
	}
	return s
}

func errorAt(loc []rune, format string, v ...interface{}) {
	pos := len(currentInput) - cap(loc)
	fmt.Fprintln(os.Stderr, string(currentInput))
	fmt.Fprintf(os.Stderr, "%*s", pos, "") // print pos spaces
	fmt.Fprintln(os.Stderr, "^ ")

	log.Fatalf(format, v...)
}

func errorTok(tok *Token, fmt string, v ...interface{}) {
	errorAt(tok.loc, fmt, v...)
}

var currentInput []rune

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

	fmt.Println(" .globl main")
	fmt.Println("main:")

	// Traverse the AST to emit assembly.
	genExpr(node)
	fmt.Println(" ret")

	assert(depth == 0)
}
