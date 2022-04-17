package main

import (
	"fmt"
)

// local variable
type Obj struct {
	next   *Obj
	name   string // variable name
	offset int    // offset from rbp
}

// Find a local variable by name.
// walks the token list
func findVar(tok *Token) *Obj {
	for v := locals; v != nil; v = v.next {
		if v.name == string(tok.loc) {
			return v
		}
	}
	return nil
}

// function
type Function struct {
	body      *Node
	locals    *Obj
	stackSize int
}

func (f *Function) String() string {
	str := `{ "function body": [`
	for v := f.body; v != nil; v = v.next {
		str += v.String() + ","
	}
	return str + "]}"
}

// AST node

type NodeKind int

const (
	ND_ADD       NodeKind = iota // +
	ND_SUB                       // -
	ND_MUL                       // *
	ND_DIV                       // /
	ND_NEG                       // unary -
	ND_EQ                        // ==
	ND_NE                        // !=
	ND_LT                        // <
	ND_LE                        // <=
	ND_ASSIGN                    // =
	ND_EXPR_STMT                 // Expression statement
	ND_VAR                       // Variable
	ND_NUM                       // Integer
)

func (nd NodeKind) String() string {
	switch nd {
	case ND_ADD:
		return "+"
	case ND_SUB:
		return " - "
	case ND_MUL:
		return " * "
	case ND_DIV:
		return " / "
	case ND_NEG:
		return " unary - "
	case ND_EQ:
		return "  =="
	case ND_NE:
		return "  !="
	case ND_LT:
		return "  < "
	case ND_LE:
		return "  <= "
	case ND_ASSIGN:
		return " = "
	case ND_EXPR_STMT:
		return " Expression statement"
	case ND_VAR:
		return " Variable "
	case ND_NUM:
		return " Integer "
	default:
		return "unknown node kind"
	}
}

// AST node type
type Node struct {
	kind     NodeKind // node kind
	next     *Node    // next node, (nodes are stored in a linked list)
	lhs      *Node    // left hand side
	rhs      *Node    // right hand side
	variable *Obj     // used if kind == ND_VAR
	val      int      // used if kind == ND_NUM
}

func (n *Node) String() string {
	// print recursively
	if n == nil {
		return "nil"
	}

	lhs := n.lhs.String()
	rhs := n.rhs.String()

	str := fmt.Sprintln(`{ "kind" : "`, n.kind, `", "var": "`, n.variable, `", "val": "`, n.val, "\"")
	return str + ",\"left\": " + lhs + ", \"right\": " + rhs + "}"
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

func NewVarNode(v *Obj) *Node {
	node := NewNode(ND_VAR)
	node.variable = v
	return node
}

// adds to the start of the locals
// LVar = Local Variable
func newLVar(name string) *Obj {
	v := new(Obj)
	v.name = name
	v.next = locals
	locals = v
	return v
}

// stmt = expr-stmt
func stmt(rest **Token, tok *Token) *Node {
	return exprStmt(rest, tok)
}

// expr-stmt = expr ";"
func exprStmt(rest **Token, tok *Token) *Node {
	node := NewUnary(ND_EXPR_STMT, expr(&tok, tok))
	*rest = skip(tok, ";")
	return node
}

// expr = assign
func expr(rest **Token, tok *Token) *Node {
	return assign(rest, tok)
}

// assign = equality ("=" assign)?
func assign(rest **Token, tok *Token) *Node {
	node := equality(&tok, tok)
	if tok.equal("=") {
		node = NewBinary(ND_ASSIGN, node, assign(&tok, tok.Next))
	}
	*rest = tok
	return node
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

// primary = "(" expr ")" | ident | num
func primary(rest **Token, tok *Token) *Node {
	if tok.equal("(") {
		node := expr(&tok, tok.Next)
		*rest = skip(tok, ")")
		return node
	}

	if tok.Kind == IDENT {
		v := findVar(tok)
		if v == nil {
			v = newLVar(string(tok.loc))
		}
		*rest = tok.Next
		return NewVarNode(v)
	}

	if tok.Kind == NUM {
		node := NewNum(tok.val)
		*rest = tok.Next
		return node
	}

	errorTok(tok, "expected an expression")
	return nil
}

// program = stmt*
// the returned Node is also a linked list of Nodes
func parse(tok *Token) *Function {
	head := new(Node)
	cur := head
	for tok.Kind != EOF {
		cur.next = stmt(&tok, tok)
		cur = cur.next
	}

	prog := new(Function)
	prog.body = head.next
	prog.locals = locals
	return prog
}
