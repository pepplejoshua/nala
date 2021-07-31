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
	lispparser "nala/lisp_parser"
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
var lang = flag.Bool("nl", true, "Interpret Nala language. Set to false for Ellisp")

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
		userProg = readAndParseSourceFile(*file, *lang)

		if userProg != nil {
			// go on to execute it
			// load in nalaFuncsProg first
			if *engine {
				globals, constants = compileAndRunProg(nalaFuncsProg, symbolTable, constants, globals, false, false)
				compileAndRunProg(userProg, symbolTable, constants, globals, false, true)
			} else {
				evaluateProg(nalaFuncsProg, env, false)
				evaluateProg(userProg, env, true)
			}
		}
		return // end execution
	}

	// run the predefined functions through first
	if *engine {
		fmt.Print("using VM...\n")
		globals, constants = compileAndRunProg(nalaFuncsProg, symbolTable, constants, globals, false, false)
	} else {
		fmt.Print("using TreeWalker...\n")
		evaluateProg(nalaFuncsProg, env, false)
	}

	if *lang {
		fmt.Print("Parsing Nala source code...\n\n")
	} else {
		fmt.Print("Parsing Ellisp source code...\n\n")
	}

	showBC := false
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
		if line == ".bc" {
			showBC = !showBC
			msg := "showing bytecode..."
			if !showBC {
				msg = "not " + msg
			}
			fmt.Println(msg)
			continue
		}

		prog, ok := parseSource(line, *lang)
		if !ok {
			continue
		}

		if *engine {
			compileAndRunProg(prog, symbolTable, constants, globals, showBC, true)
		} else {
			evaluateProg(prog, env, true)
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

func compileAndRunProg(prog *ast.Program, st *compiler.SymbolTable, cons,
	globals []object.Object, show bool, showRes bool) ([]object.Object, []object.Object) {
	comp := compiler.NewWithState(st, cons)
	err := comp.Compile(prog)
	if err != nil {
		if showRes {
			fmt.Println(fmt.Errorf("compiler error: %s", err))
		}
		return globals, cons
	}

	now := time.Now()
	top, globs, err := runVM(comp.ByteCode(), globals)
	duration := time.Since(now)

	if err != nil {
		if showRes {
			fmt.Println(fmt.Errorf("vm error: %s", err))
		}
		fmt.Println(comp.ByteCode().Instructions.String())
		return globals, cons
	}

	if showRes {
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

func evaluateProg(prog *ast.Program, env *object.Environment, show bool) {
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
	prog := readAndParseSourceFile(FUNCSPATH, true)
	if prog != nil {
		fmt.Println("read nala functions")
		return prog
	}

	return nil
}

func readAndParseSourceFile(path string, nalaSrc bool) *ast.Program {
	var src string
	if nalaSrc {
		src = getFileContents(path + ".nl")
	} else {
		src = getFileContents(path + ".el")
	}

	prog, ok := parseSource(src, nalaSrc)
	if ok {
		return prog
	}
	return nil
}

func parseSource(src string, nalaSrc bool) (*ast.Program, bool) {
	l := lexer.New(src)

	var prog *ast.Program
	if nalaSrc {
		p := parser.New(l)
		prog = p.ParseProgram()
		if hasErrors(p) {
			fmt.Println("couldn't parse source")
			printParseErrors(os.Stdout, p.Errors())
			return nil, false
		}
	} else {
		p := lispparser.New(l)
		prog = p.ParseProgram()
		if hasErrorsL(p) {
			fmt.Println("couldn't parse source")
			printParseErrors(os.Stdout, p.Errors())
			return nil, false
		}
	}
	return prog, true
}

func getFileContents(location string) string {
	var data []byte
	var err error
	data, err = ioutil.ReadFile("./code/" + location)

	if err != nil {
		panic(err)
	}
	return string(data)
}

func hasErrors(p *parser.Parser) bool {
	return len(p.Errors()) != 0
}

func hasErrorsL(p *lispparser.Parser) bool {
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
