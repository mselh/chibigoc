package main

import (
	"fmt"
	"log"
)

// helpers

// Round up `n` to the nearest multiple of `align`.
// For instance,
//
// align_to(5, 8) returns 8
// align_to(11, 8) returns 16.
func alignTo(n, align int) int {
	return (n + align - 1) / align * align
}

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
		fmt.Printf(" lea %d(%%rbp), %%rax\n", node.variable.offset)
		// `lea %offset(%src), %t` => t = %rsp + %d
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
	switch node.kind {
	case ND_IF:
		count++
		genExpr(node.cond)
		fmt.Println(" cmp $0, %rax")
		fmt.Printf(" je .L.else.%d\n", count)
		genStmt(node.then)
		fmt.Printf(" jmp .L.end.%d\n", count)
		fmt.Printf(".L.else.%d:\n", count)
		if node.els != nil {
			genStmt(node.els)
		}
		fmt.Printf(".L.end.%d:\n", count)
		return
	case ND_FOR:
		count++
		genStmt(node.init)
		fmt.Printf(".L.begin.%d:\n", count)
		if node.cond != nil {
			genExpr(node.cond)
			fmt.Println(" cmp $0, %rax")
			fmt.Printf(" je .L.end.%d\n", count)
		}
		genStmt(node.then)
		if node.inc != nil {
			genExpr(node.inc)
		}
		fmt.Printf("  jmp .L.begin.%d\n", count)
		fmt.Printf(".L.end.%d:\n", count)
		return
	case ND_BLOCK:
		for n := node.body; n != nil; n = n.next {
			genStmt(n)
		}
		return
	case ND_RETURN:
		genExpr(node.lhs)
		fmt.Println(" jmp .L.return")
		// jump to .L.return label
		return
	case ND_EXPR_STMT:
		genExpr(node.lhs)
		return
	}

	log.Fatalln("invalid expression")
}

// Assign offsets to local variables
func assignLVarOffsets(prog *Function) {
	offset := 0
	for v := prog.locals; v != nil; v = v.next {
		offset += 8
		v.offset = -offset
	}
	prog.stackSize = alignTo(offset, 16)
}

func codegen(prog *Function) {
	assignLVarOffsets(prog)

	fmt.Println(" .globl main")
	fmt.Println("main:")

	// Prologue
	fmt.Println(" push %rbp")
	fmt.Println(" mov %rsp, %rbp")
	fmt.Printf(" sub $%d, %%rsp\n", prog.stackSize)

	// assumes body is a ND_BLOCK kind
	genStmt(prog.body)
	assert(depth == 0)

	fmt.Println(".L.return:")
	fmt.Println(" mov %rbp, %rsp")
	fmt.Println(" pop %rbp")
	fmt.Println(" ret")
}
