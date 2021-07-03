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

func TestLetStatements(t *testing.T) {
	input := `
	let x = 5;
	let y = 200;
	let someName25 = 25;
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

	tests := []IdentifierTest{
		{"x"},
		{"y"},
		{"someName25"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]

		if !testLetStatement(t, stmt, tt.ID) {
			return
		}
	}
}

// handles testing for the Name (*Identifier) portion of a LetStatement
func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
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
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
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
		t.Log(tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		prog := p.ParseProgram()
		checkParseErrors(t, p)

		actual := prog.String()

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
