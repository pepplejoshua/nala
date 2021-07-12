package main

import (
	"flag"
	"fmt"
	"nala/compiler"
	"nala/evaluator"
	"nala/lexer"
	"nala/object"
	"nala/parser"
	"nala/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

// replace with scan or file input
var input = `
let fibo = fn(x) {
	if (x < 2) {
		return x
	}
	fibo(x-1) + fibo(x-2);
};
fibo(35);
`

func main() {
	flag.Parse()

	var duration time.Duration
	var res object.Object

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(prog)
		if err != nil {
			fmt.Println("compiler error: %s\n", err)
			return
		}

		machine := vm.New(comp.ByteCode())
		// capture time
		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s\n", err)
			return
		}

		duration = time.Since(start)
		res = machine.LastPoppedElement()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		res = evaluator.Eval(prog, env)
		duration = time.Since(start)
	}

	fmt.Printf("engine=%s, result=%s, duration=%s\n",
		*engine, res.Inspect(), duration)
}
