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

func genExpr(node *Node) {

	switch node.kind {
	case ND_NUM:
		fmt.Printf(" mov $%d, %%rax\n", node.val)
		return
	case ND_NEG:
		genExpr(node.lhs)
		fmt.Println(" neg %rax")
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

func codegen(node *Node) {
	fmt.Println(" .globl main")
	fmt.Println("main:")

	genExpr(node)
	fmt.Println(" ret")

	assert(depth == 0)
}
