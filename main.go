package main

import (
	"fmt"
	"os"
	"unicode"
)

// base10
func strtol(s *string) int64 {

	var val int64 = 0

	for i, r := range *s {
		if !unicode.IsDigit(r) {
			*s = string([]rune(*s)[i:])
			return val
		}

		digit := int64(r) - 48
		val = 10*val + digit
	}

	*s = ""
	return val
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "%s: invalid number of args\n", os.Args[0])
		os.Exit(1)
	}

	p := os.Args[1]

	fmt.Println(" .globl main")
	fmt.Println("main:")
	fmt.Printf("  mov $%d, %%rax\n", strtol(&p))

	for len(p) > 0 {

		if p[0] == '+' {

			p = p[1:]
			fmt.Printf(" add $%d, %%rax\n", strtol(&p))
			continue
		}

		if p[0] == '-' {

			p = p[1:]
			fmt.Printf(" sub $%d, %%rax\n", strtol(&p))
			continue
		}

		fmt.Fprintf(os.Stderr, "unexpected character: '%c'\n", p[0])
		os.Exit(1)
	}

	fmt.Println(" ret")
	os.Exit(0)
}
