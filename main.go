package main

import (
	"fmt"
	"nala/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	if len(os.Args) != 1 {
		repl.Start(os.Stdin, os.Stdout)
	} else {
		fmt.Printf("Hello %s! This is Nala programming language!\n", user.Username)
		fmt.Printf("Feel free to type commands\n")
		fmt.Printf("Enter '.q' to quit the REPL.\n")
		repl.Start(os.Stdin, os.Stdout)
	}

}
