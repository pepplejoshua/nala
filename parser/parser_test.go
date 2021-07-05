package parser

import (
	"fmt"
	"nala/ast"
	"nala/lexer"
	"testing"
)

type IdentifierTest struct {
	ID string
}

type LetStatementTest struct {
	input         string
	expectedID    string
	expectedValue interface{}
}

type PrefixTest struct {
	input    string
	operator string
	value    interface{}
}

type InfixTest struct {
	input      string
	leftValue  interface{}
	operator   string
	rightValue interface{}
}

type OperatorPrecTest struct {
	input    string
	expected string
}

type BooleanTest struct {
	input    string
	expected bool
}

type IfExpressionTest struct {
	input       string
	Condition   InfixTest
	Conseq      string
	Alternative string
}

type FuncParameterTest struct {
	input          string
	expectedParams []string
}

type GenericTest struct {
	input    string
	expected interface{}
}

func TestLetStatements(t *testing.T) {
	tests := []LetStatementTest{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParseErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain any statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedID, tt.expectedValue) {
			return
		}

	}
}

// handles testing for the Name (*Identifier) portion of a LetStatement
func testLetStatement(t *testing.T, s ast.Statement, name string, value interface{}) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral() not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	// type assertion
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		// %T for type info
		// %v for internal representation
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() is not %s. got=%s", name, letStmt.TokenLiteral())
		return false
	}
	if !testLiteralExpression(t, letStmt.Value, value) {
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	input := `
	return 5;
	return 10;
	return 20000000;
	`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParseErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral() not 'return. got=%q", returnStmt.TokenLiteral())
		}
	}
}

func checkParseErrors(t *testing.T, p *Parser) {
	errs := p.Errors()

	if len(errs) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errs))

	for _, err := range errs {
		t.Errorf("parser error: %q", err)
	}
	t.FailNow()
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	expected := "foobar"

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParseErrors(t, p)

	if len(prog.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(prog.Statements))
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", prog.Statements[0])
	}

	if !testIdentifier(t, stmt.Expression, expected) {
		return
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral() not %s. got=%s", value, ident.TokenLiteral())
		return false
	}
	return true
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "24;"
	var expectedVal int64 = 24

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParseErrors(t, p)

	if len(prog.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(prog.Statements))
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", prog.Statements[0])
	}

	if !testIntegerLiteral(t, stmt.Expression, expectedVal) {
		return
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []PrefixTest{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!a", "!", "a"},
		{"!false", "!", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		prog := p.ParseProgram()
		checkParseErrors(t, p)

		if len(prog.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d", len(prog.Statements))
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", prog.Statements[0])
		}

		if !testPrefixExpression(t, stmt.Expression, tt.operator, tt.value) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		// %q for integer
		t.Errorf("integ.Value not %q. got=%q", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral() not %s. got=%s", fmt.Sprintf("%d", 24), integ.TokenLiteral())
		return false
	}
	return true
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []InfixTest{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 % 5", 5, "%", 5},
		{"5 < 5", 5, "<", 5},
		{"5 > 5", 5, ">", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"a < b", "a", "<", "b"},
		{"b > c", "b", ">", "c"},
		{"c == d", "c", "==", "d"},
		{"d != e", "d", "!=", "e"},
		{"true != false", true, "!=", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		prog := p.ParseProgram()
		checkParseErrors(t, p)

		if len(prog.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d", len(prog.Statements))
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", prog.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []OperatorPrecTest{
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)"},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))"},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))"},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 4 == true",
			"((3 < 4) == true)",
		},
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"-1 * 2 + 3",
			"(((-1) * 2) + 3)",
		},
		{
			"let Joshua = deadAF;",
			"let Joshua = deadAF;",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
	}

	for _, tt := range tests {

		l := lexer.New(tt.input)
		p := New(l)
		prog := p.ParseProgram()
		checkParseErrors(t, p)

		actual := prog.String()
		t.Log(actual)
		if actual != tt.expected {
			t.Errorf("expected=%s, got=%s", tt.expected, actual)
		}
	}
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("exp is not *ast.InfixExpression. got=%T", exp)
	}
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		// %s is for strings
		t.Errorf("exp.Operator is not '%s'. got=%s", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

func testPrefixExpression(t *testing.T, exp ast.Expression, operator string, value interface{}) bool {
	pExp, ok := exp.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("exp not *ast.PrefixExpression. got=%T", exp)
		return false
	}
	if pExp.Operator != operator {
		// %s is for strings
		t.Errorf("exp.Operator is not '%s'. got=%s", operator, pExp.Operator)
		return false
	}
	if !testLiteralExpression(t, pExp.Right, value) {
		return false
	}
	return true
}

func TestBooleanExpressions(t *testing.T) {
	tests := []BooleanTest{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		prog := p.ParseProgram()
		checkParseErrors(t, p)

		if len(prog.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d", len(prog.Statements))
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", prog.Statements[0])
		}

		if !testBooleanLiteral(t, stmt.Expression, tt.expected) {
			return
		}
	}
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	booln, ok := exp.(*ast.Boolean)
	if !ok {
		t.Fatalf("exp is not *ast.Boolean. got=%T", exp)
		return false
	}
	if booln.Value != value {
		// %t for boolean
		t.Errorf("booln.Value not %t. got=%t", value, booln.Value)
		return false
	}
	if booln.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("booln.TokenLiteral() not %s. got=%s", fmt.Sprintf("%d", 24), booln.TokenLiteral())
		return false
	}
	return true
}

func TestIfExpression(t *testing.T) {
	tests := []IfExpressionTest{
		{
			input: `if (x < y) { x }`,
			Condition: InfixTest{
				leftValue:  "x",
				operator:   "<",
				rightValue: "y",
			},
			Conseq: "x",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParseErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.IfExpression)

		if !ok {
			t.Fatalf("exp is not *ast.IfExpression. got=%T", stmt.Expression)
		}

		if !testInfixExpression(t, exp.Condition, tt.Condition.leftValue, tt.Condition.operator, tt.Condition.rightValue) {
			return
		}

		if len(exp.Consequence.Statements) != 1 {
			t.Fatalf("consequence is not 1 statements. got=%d", len(program.Statements))
		}

		conseq, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
		}

		if !testIdentifier(t, conseq.Expression, tt.Conseq) {
			return
		}

		if exp.Alternative != nil {
			// +v addes the field names for Go structures
			t.Errorf("alternative was not nil. got=%+v", exp.Alternative)
		}
	}

}

func TestIfElseExpression(t *testing.T) {
	tests := []IfExpressionTest{
		{
			input: `if (x > y) { x } else { y }`,
			Condition: InfixTest{
				leftValue:  "x",
				operator:   ">",
				rightValue: "y",
			},
			Conseq:      "x",
			Alternative: "y",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.IfExpression)

		if !ok {
			t.Fatalf("exp is not *ast.IfExpression. got=%T", stmt.Expression)
		}

		if !testInfixExpression(t, exp.Condition, tt.Condition.leftValue, tt.Condition.operator, tt.Condition.rightValue) {
			return
		}

		if len(exp.Consequence.Statements) != 1 {
			t.Fatalf("consequence is not 1 statements. got=%d", len(program.Statements))
		}

		conseq, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("conseq.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
		}

		if !testIdentifier(t, conseq.Expression, tt.Conseq) {
			return
		}

		if len(exp.Consequence.Statements) != 1 {
			t.Fatalf("consequence is not 1 statements. got=%d", len(program.Statements))
		}

		alt, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("alt.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Alternative.Statements[0])
		}

		if !testIdentifier(t, alt.Expression, tt.Alternative) {
			return
		}
	}

}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if len(fn.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2. got=%d", len(fn.Parameters))
	}

	testLiteralExpression(t, fn.Parameters[0], "x")
	testLiteralExpression(t, fn.Parameters[1], "y")

	if len(fn.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got%d\n", len(fn.Body.Statements))
	}

	bodyStmt, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body statement is not ast.ExpressionStatement. got=%t", fn.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []FuncParameterTest{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		fn := stmt.Expression.(*ast.FunctionLiteral)

		if len(fn.Parameters) != len(tt.expectedParams) {
			t.Errorf("length of parameters is wrong. want=%d. got=%d\n",
				len(tt.expectedParams), len(fn.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, fn.Parameters[i], ident)
		}
	}

}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestStringLiteralExpressions(t *testing.T) {
	tests := []GenericTest{
		{
			`"hello world";`,
			"hello world",
		},
		{
			`"joshua pepple";`,
			"joshua pepple",
		},
		{
			`"stringggg";`,
			"stringggg",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		prog := p.ParseProgram()
		checkParseErrors(t, p)

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("prog.Statements[0] is not *ast.ExpressionStatement. got=%T", prog.Statements[0])
		}

		if !testStringLiteral(t, stmt.Expression, tt.expected) {
			return
		}
	}
}

func testStringLiteral(t *testing.T, stmt ast.Expression, expected interface{}) bool {
	str, ok := stmt.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt)
		return false
	}
	if str.Value != expected {
		t.Errorf("literal.Value not %q. got=%q", expected, str.Value)
		return false
	}
	return true
}
