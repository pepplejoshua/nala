package repl

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"nala/ast"
	"nala/compiler"
	"nala/evaluator"
	"nala/lexer"
	"nala/object"
	"nala/parser"
	"nala/vm"
	"os"
	"time"
)

const PROMPT = "=> "

const FUNCSPATH = "functions"

var engine = flag.Bool("vm", true, "Use Compiler and Virtual Machine")
var file = flag.String("f", "", "Run Nala source file")

// const FUNCSPATH = "./nl/test.nl"

// TODO: TO BE CLEANED UP
func Start(in io.Reader, out io.Writer) {
	env := object.NewEnvironment()
	symbolTable := compiler.NewSymbolTable()
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	// IMPLEMENT COMPILATION DOWN THERE USING 3 state vars above
	nalaFuncsProg := parseNalaFunctions()

	flag.Parse()

	var userProg *ast.Program
	if *file != "" {
		// try to parse user file
		userProg = readAndParseSourceFile(*file)

		if userProg != nil {
			// go on to execute it
			// load in nalaFuncsProg first
			if *engine {
				globals, constants = compileAndRunNalaProg(nalaFuncsProg, symbolTable, constants, globals, false)
				compileAndRunNalaProg(userProg, symbolTable, constants, globals, true)
			} else {
				evaluateNalaProg(nalaFuncsProg, env, false)
				evaluateNalaProg(userProg, env, true)
			}
		}
		return // end execution
	}

	// run the predefined functions through first
	if *engine {
		fmt.Print("using VM...\n\n")
		globals, constants = compileAndRunNalaProg(nalaFuncsProg, symbolTable, constants, globals, false)
	} else {
		fmt.Print("using TreeWalker...\n\n")
		evaluateNalaProg(nalaFuncsProg, env, false)
	}

	for {
		fmt.Print(PROMPT)
		scanner := bufio.NewScanner(in)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		if line == ".q" {
			io.WriteString(out, "Arigatōgozaimashita!\n")
			break
		}

		if line == ".m" {
			for n, o := range env.GetStore() {
				fmt.Printf("%q: (%+v)", n, o)
			}
			fmt.Println()
			continue
		}
		if line == ".sb" {
			fmt.Println(".builtins.")
			fmt.Println(".========.")
			for _, fn := range object.Builtins {
				fmt.Println(fn.Name, ": ", fn.BuiltIn.Desc)
			}
			fmt.Println()
			continue
		}

		prog, ok := parseSource(line)
		if !ok {
			continue
		}

		if *engine {
			compileAndRunNalaProg(prog, symbolTable, constants, globals, true)
		} else {
			evaluateNalaProg(prog, env, true)
		}
	}
}

const CAT_FACE = ` A_A
(-.-)
 |-|
/   \
|     |   __
|  || |  |  \__
\_||_/_/
`

func compileAndRunNalaProg(prog *ast.Program, st *compiler.SymbolTable, cons,
	globals []object.Object, show bool) ([]object.Object, []object.Object) {
	comp := compiler.NewWithState(st, cons)
	err := comp.Compile(prog)
	if err != nil {
		if show {
			fmt.Println(fmt.Errorf("compiler error: %s", err))
		}
		return globals, cons
	}

	now := time.Now()
	top, globs, err := runVM(comp.ByteCode(), globals)
	duration := time.Since(now)

	if err != nil {
		if show {
			fmt.Println(fmt.Errorf("vm error: %s", err))
		}
		fmt.Println(comp.ByteCode().Instructions.String())
		return globals, cons
	}

	if show {
		fmt.Println(top.Inspect())
	}

	fmt.Printf("[duration: %s]\n", duration)
	cons = comp.ByteCode().Constants
	if show && top.Type() != object.COMPILED_FUNCTION_OBJ &&
		// top.Type() != object.ERROR_OBJ &&
		// top.Type() != object.CLOSURE_OBJ &&
		// top.Type() != object.ARRAY_OBJ &&
		top.Type() != object.HASHMAP_OBJ &&
		top.Type() != object.BUILTIN_OBJ {
		fmt.Println("\n*DISASSEMBLED BYTECODE*")
		fmt.Println("************************")
		ins := comp.ByteCode().Instructions
		comp.Decompile(ins, cons, globs, "", 0)
		println()
	}
	println()
	return globs, cons
}

func runVM(bc *compiler.ByteCode, globals []object.Object) (object.Object, []object.Object, error) {
	machine := vm.NewWithGlobalsStore(bc, globals)
	err := machine.Run()
	top := machine.LastPoppedElement()

	return top, machine.Globals(), err
}

func evaluateNalaProg(prog *ast.Program, env *object.Environment, show bool) {
	res := evaluator.Eval(prog, env)

	if show {
		if res != nil {
			io.WriteString(os.Stdout, res.Inspect()+"\n")
		} else {
			io.WriteString(os.Stdout, "NIL\n")
		}
	}
}

func parseNalaFunctions() *ast.Program {
	prog := readAndParseSourceFile(FUNCSPATH)
	if prog != nil {
		fmt.Println("read nala functions")
		return prog
	}

	return nil
}

func readAndParseSourceFile(path string) *ast.Program {
	src := getFileContents(path)

	prog, ok := parseSource(src)
	if ok {
		return prog
	}
	return nil
}

func parseSource(src string) (*ast.Program, bool) {
	l := lexer.New(src)
	p := parser.New(l)

	prog := p.ParseProgram()
	if hasErrors(p) {
		fmt.Println("couldn't parse source")
		printParseErrors(os.Stdout, p.Errors())
		return nil, false
	}
	return prog, true
}

func getFileContents(location string) string {
	data, err := ioutil.ReadFile("./nl/" + location + ".nl")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func hasErrors(p *parser.Parser) bool {
	return len(p.Errors()) != 0
}

func printParseErrors(out io.Writer, errs []string) {
	io.WriteString(out, CAT_FACE)
	io.WriteString(out, "Whoops! What an antagonized cat!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errs {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
