package repl

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"nala/compiler"
	"nala/evaluator"
	"nala/lexer"
	"nala/object"
	"nala/parser"
	"nala/vm"
	"os"
)

const PROMPT = "=> "

const FUNCSPATH = "./nl/functions.nl"

// const FUNCSPATH = "./nl/test.nl"

func Start(in io.Reader, out io.Writer) {
	// this wraps the input with a Buffer that we can Scan?
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	// pre read some of handwritten functions from files
	readNalaFunctions(env, out)
	vmSelector := flag.Bool("vm", false, "use Virtual Machine")
	flag.Parse()

	if *vmSelector {
		fmt.Println("using VM...")
	}
	args := os.Args[1:]

	if len(args) == 1 && args[0] != "-vm=true" {
		fName := args[0]
		pth := "./nl/" + fName + ".nl"
		// fmt.Println(pth)
		src := getFileContents(pth)
		// fmt.Println(src)

		// pl := lexer.New(src)
		l := lexer.New(src)
		p := parser.New(l)

		prog := p.ParseProgram()
		if hasErrors(p, out) {
			io.WriteString(out, fmt.Sprintf("Couldn't read Nala Functions Source from %q\n", pth))
			printParseErrors(out, p.Errors())
		}

		// this expands all macros, i.e edits the source tree, replacing all macros with their exact definitions
		res := evaluator.Eval(prog, env)
		if res != nil {
			io.WriteString(out, res.Inspect()+"\n")
		} else {
			io.WriteString(out, "NIL\n")
		}
		return
	}

	for {
		fmt.Print(PROMPT)
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
		// pl := lexer.New(line)
		l := lexer.New(line)
		p := parser.New(l)

		// for tok := pl.NextToken(); tok.Type != token.EOF; tok = pl.NextToken() {
		// 	fmt.Printf("%+v\n", tok)
		// }

		prog := p.ParseProgram()

		if hasErrors(p, out) {
			printParseErrors(out, p.Errors())
		} else {
			var res object.Object
			if *vmSelector {

				compiler := compiler.New()
				err := compiler.Compile(prog)
				if err != nil {
					fmt.Println(fmt.Errorf("compiler error: %s", err))
				}
				vm := vm.New(compiler.ByteCode())
				err = vm.Run()
				if err != nil {
					fmt.Println(fmt.Errorf("vm error: %s", err))
					continue
				}

				stackElem := vm.LastPoppedElement()
				io.WriteString(out, stackElem.Inspect()+"\n")
				io.WriteString(out, compiler.ByteCode().Instructions.String()+"\n")
			} else {
				res = evaluator.Eval(prog, env)
				if res != nil {
					io.WriteString(out, res.Inspect()+"\n")
				} else {
					io.WriteString(out, "NIL\n")
				}
			}
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

func printParseErrors(out io.Writer, errs []string) {
	io.WriteString(out, CAT_FACE)
	io.WriteString(out, "Whoops! What an antagonized cat!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errs {
		io.WriteString(out, "\t"+msg+"\n")
	}

}

func readNalaFunctions(env *object.Environment, out io.Writer) {
	src := getFileContents(FUNCSPATH)
	// pl := lexer.New(src)
	l := lexer.New(src)
	p := parser.New(l)

	// for tok := pl.NextToken(); tok.Type != token.EOF; tok = pl.NextToken() {
	// fmt.Printf("%+v\n", tok)
	// }

	prog := p.ParseProgram()
	if hasErrors(p, out) {
		io.WriteString(out, fmt.Sprintf("Couldn't read Nala Functions Source from %q\n", FUNCSPATH))
		printParseErrors(out, p.Errors())
	}

	evaluator.Eval(prog, env)
}

func getFileContents(location string) string {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func hasErrors(p *parser.Parser, out io.Writer) bool {
	return len(p.Errors()) != 0
}
