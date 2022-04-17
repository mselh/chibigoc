package main

import (
	"fmt"
	"log"
)

//
// Code generator
//

func push() {
	fmt.Println(" push %rax")
	depth++
}

func pop(arg string) {
	fmt.Printf(" pop %s\n", arg)
	depth--
}

// Compute the absolute address of a given node.
// It's an error if a given node does not reside in memory.
func genAddr(node *Node) {
	if node.kind == ND_VAR {
		var offset int = (int(node.name) - 'a' + 1) * 8
		fmt.Printf(" lea %d(%%rbp), %%rax\n", -offset)
		// lea %offset(%src), %t
		// t = %rsp + %d
		return
	}

	log.Fatalln("not an lvalue")
}

func genExpr(node *Node) {

	switch node.kind {
	case ND_NUM:
		fmt.Printf(" mov $%d, %%rax\n", node.val)
		return
	case ND_NEG:
		genExpr(node.lhs)
		fmt.Println(" neg %rax")
		return
	case ND_VAR:
		genAddr(node)
		fmt.Println(" mov (%rax), %rax")
		// rax ta adres var
		// rax in isaret ettigi adresin icindeki degeri, rax a tasi
		// new_rax = *rax
		return
	case ND_ASSIGN:
		genAddr(node.lhs)
		push()
		genExpr(node.rhs)
		pop("%rdi")
		fmt.Println(" mov %rax, (%rdi)")
		// raxdaki degeri, rdi nin isaret ettigi konuma tasi
		// *rdi = rax
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
		fmt.Println(" idiv %rdi")
		return
	case ND_EQ, ND_NE, ND_LT, ND_LE:
		fmt.Println(" cmp %rdi, %rax")

		if node.kind == ND_EQ {
			fmt.Println(" sete %al")
		} else if node.kind == ND_NE {
			fmt.Println(" setne %al")
		} else if node.kind == ND_LT {
			fmt.Println(" setl %al")
		} else if node.kind == ND_LE {
			fmt.Println(" setle %al")
		}

		fmt.Println(" movzb %al, %rax")
		return

	}

	log.Fatalln("invalid expression")
}

func genStmt(node *Node) {
	if node.kind == ND_EXPR_STMT {
		genExpr(node.lhs)
		return
	}

	log.Fatalln("invalid expression")
}

func codegen(node *Node) {
	fmt.Println(" .globl main")
	fmt.Println("main:")

	// Prologue
	fmt.Println(" push %rbp")
	fmt.Println(" mov %rsp, %rbp")
	fmt.Println(" sub $208, %rsp")
	// a note from github.com/ksco
	// `208 == ('z' - 'a' + 1) * 8,
	// it's the stack size for all possible,
	// single-letter 64 bit integer variables.`
	//
	// right now, stack size is fixed to 208

	for n := node; n != nil; n = n.next {
		genStmt(n)
		assert(depth == 0)
	}

	fmt.Println(" mov %rbp, %rsp")
	fmt.Println(" pop %rbp")
	fmt.Println(" ret")
}
