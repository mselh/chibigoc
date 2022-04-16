package main

type NodeKind int

const (
	ND_ADD NodeKind = iota // +
	ND_SUB                 // -
	ND_MUL                 // *
	ND_DIV                 // /
	ND_NEG                 // unary -
	ND_EQ                  // ==
	ND_NE                  // !=
	ND_LT                  // <
	ND_LE                  // <=
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

func NewUnary(kind NodeKind, expr *Node) *Node {
	node := NewNode(kind)
	node.lhs = expr
	return node
}

func NewNum(val int) *Node {
	node := NewNode(ND_NUM)
	node.val = val

	return node
}

// expr = equality
func expr(rest **Token, tok *Token) *Node {
	return equality(rest, tok)
}

// equality = relational ("==" relational | "!=" relational)*
func equality(rest **Token, tok *Token) *Node {
	node := relational(&tok, tok)

	for {

		if tok.equal("==") {
			node = NewBinary(ND_EQ, node, relational(&tok, tok.Next))
			continue
		}

		if tok.equal("!=") {
			node = NewBinary(ND_NE, node, relational(&tok, tok.Next))
			continue
		}

		*rest = tok
		return node
	}
}

// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
func relational(rest **Token, tok *Token) *Node {
	node := add(&tok, tok)

	for {

		if tok.equal("<") {
			node = NewBinary(ND_LT, node, add(&tok, tok.Next))
			continue
		}

		if tok.equal("<=") {
			node = NewBinary(ND_LE, node, add(&tok, tok.Next))
			continue
		}

		if tok.equal(">") {
			node = NewBinary(ND_LT, add(&tok, tok.Next), node)
			continue
		}

		if tok.equal(">=") {
			node = NewBinary(ND_LE, add(&tok, tok.Next), node)
			continue
		}

		*rest = tok
		return node
	}

}

// add = mul ("+" mul | "-" mul)*
func add(rest **Token, tok *Token) *Node {
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

// mul = unary ("*" unary | "/" unary)*
func mul(rest **Token, tok *Token) *Node {
	node := unary(&tok, tok) // left node for the new binary node

	for {
		if tok.equal("*") {
			// rhs is primary(&tok,.)
			node = NewBinary(ND_MUL, node, unary(&tok, tok.Next))
			continue
		}

		if tok.equal("/") {
			node = NewBinary(ND_DIV, node, unary(&tok, tok.Next))
			continue
		}

		*rest = tok
		return node
	}
}

// unary = ("+" | "-") unary
//       | primary
func unary(rest **Token, tok *Token) *Node {

	// doesn't affect the sign
	if tok.equal("+") {
		return unary(rest, tok.Next)
	}

	if tok.equal("-") {
		return NewUnary(ND_NEG, unary(rest, tok.Next))
	}

	return primary(rest, tok)
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
