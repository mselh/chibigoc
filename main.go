package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "%s: invalid number of args\n", os.Args[0])
	}

	i, _ := strconv.Atoi(os.Args[1])

	fmt.Println(" .globl main")
	fmt.Println("main:")
	fmt.Printf(" mov $%d, %%rax\n", i)
	fmt.Println(" ret")
	os.Exit(0)
}
