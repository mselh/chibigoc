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
		log.Fatal("expected a number, instead: ", *t)
	}
	return t.val
}

// Ensure that the current token is `s`
func skip(t *Token, s string) *Token {
	if !t.equal(s) {
		log.Fatalf("expected '%s'", s)
	}
	return t.Next
}

// Tokenize `p` and returns new tokens.
func tokenize(p []rune) *Token {
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

		if p[0] == '+' || p[0] == '-' {
			cur.Next = NewToken(PUNCT, p[0:1])
			cur = cur.Next
			p = p[1:]
			continue
		}

		log.Fatal("invalid token")
	}

	cur, cur.Next = cur.Next, NewToken(EOF, p)
	return head.Next
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

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("%s: invalid number of args\n", os.Args[0])
	}

	// a unicode lexer
	p := []rune(os.Args[1])

	var tok *Token = tokenize(p)
	// fmt.Fprintln(os.Stderr, "input: ", p)
	// fmt.Fprintln(os.Stderr, tok.String())

	fmt.Println(" .globl main")
	fmt.Println("main:")

	// the first token must be a number
	fmt.Printf("  mov $%d, %%rax\n", tok.number())
	tok = tok.Next

	// ... followed by either `+ <number>` or `- <number>`.
	for tok.Kind != EOF {

		if tok.equal("+") {
			fmt.Printf(" add $%d, %%rax\n", tok.Next.number())
			tok = tok.Next.Next
			continue
		}

		tok = skip(tok, "-")
		fmt.Printf(" sub $%d, %%rax\n", tok.number())
		tok = tok.Next
		continue

	}

	fmt.Println(" ret")
}
