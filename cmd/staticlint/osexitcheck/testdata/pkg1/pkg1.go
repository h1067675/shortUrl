package main

import (
	"fmt"
	"os"
)

func outmain() {
	os.Exit(0) // want
}

func main() {
	fmt.Print(1) // want
	os.Exit(0)   // want "used os.Exit"
}
