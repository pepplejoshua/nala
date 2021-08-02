package main

import (
	"nala/repl"
	"os"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
