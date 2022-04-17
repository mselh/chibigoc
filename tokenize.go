package main

import (
	"fmt"
	"log"
	"os"
	"strings"
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

// Read a punctuator token from p and returns its length
//
// unicode.isPunct behaves differently then `ctype.h => int ispunct(char)`
//
// All punctuations in C:
// ! " # $ % & ' ( ) * + , - . / : ;
// < = > ? @ [ \ ] ^ _ ` { | } ~
func readPunct(p []rune) int {

	var s string
	// no need to stringify the []rune,
	// more than max. of possible punctuation length
	if len(p) > 2 {
		s = string(p[:2])
	} else {
		s = string(p[:len(p)])
	}

	if strings.HasPrefix(s, "==") ||
		strings.HasPrefix(s, "!=") ||
		strings.HasPrefix(s, "<=") ||
		strings.HasPrefix(s, ">=") {
		return 2
	}

	if unicode.IsPunct(p[0]) ||
		strings.HasPrefix(s, "+") ||
		strings.HasPrefix(s, "-") ||
		strings.HasPrefix(s, "<") ||
		strings.HasPrefix(s, ">") ||
		strings.HasPrefix(s, "=") ||
		strings.HasPrefix(s, ";") {
		return 1
	}
	return 0
}

//
// Tokenizer
//

// token data type
// it is implemented as a linked list

type TokenKind int

const (
	PUNCT TokenKind = iota // punctuators
	IDENT                  // identifiers
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

		// identifier
		if p[0] >= 'a' && p[0] <= 'z' {
			cur.Next = NewToken(IDENT, p[0:1])
			cur = cur.Next
			p = p[1:]
			continue
		}

		// punctuators
		if punctLen := readPunct(p); punctLen > 0 {
			cur.Next = NewToken(PUNCT, p[0:punctLen])
			cur = cur.Next
			p = p[punctLen:]
			continue
		}

		errorAt(p, "invalid token")
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
		s += fmt.Sprintln("[rune loc: ", (n.loc), "str: ", string(n.loc), " kind: ", n.Kind, " val:", n.val, "],")
	}
	return s
}

func errorAt(loc []rune, format string, v ...interface{}) {
	pos := len(currentInput) - cap(loc)
	fmt.Fprintln(os.Stderr, string(currentInput))
	fmt.Fprintf(os.Stderr, "%*s", pos, "") // print pos spaces
	fmt.Fprintln(os.Stderr, "^ ")
	fmt.Fprintln(os.Stderr, "token: ", string(loc))

	log.Fatalf(format, v...)
}

func errorTok(tok *Token, fmt string, v ...interface{}) {
	errorAt(tok.loc, fmt, v...)
}
