package ast

type ModifierFunc func(Node) Node

// to allow rewriting operators, change them strings to StringLiterals
// then add an option for them here
func Modify(node Node, modifier ModifierFunc) Node {
	switch node := node.(type) {
	case *Program:
		for i, statement := range node.Statements {
			t, ok := Modify(statement, modifier).(Statement)
			if ok {
				node.Statements[i] = t
			}
		}
	case *ExpressionStatement:
		node.Expression, _ = Modify(node.Expression, modifier).(Expression)
	case *InfixExpression:
		node.Left, _ = Modify(node.Left, modifier).(Expression)
		node.Right, _ = Modify(node.Right, modifier).(Expression)
	case *PrefixExpression:
		node.Right, _ = Modify(node.Right, modifier).(Expression)
	case *IndexExpression:
		node.Left, _ = Modify(node.Left, modifier).(Expression)
		node.Index, _ = Modify(node.Index, modifier).(Expression)
	case *IfExpression:
		node.Condition, _ = Modify(node.Condition, modifier).(Expression)
		node.Consequence, _ = Modify(node.Consequence, modifier).(*BlockStatement)
		if node.Alternative != nil {
			node.Alternative, _ = Modify(node.Alternative, modifier).(*BlockStatement)
		}
	case *BlockStatement:
		for i, s := range node.Statements {
			node.Statements[i], _ = Modify(s, modifier).(Statement)
		}
	case *ReturnStatement:
		node.ReturnValue, _ = Modify(node.ReturnValue, modifier).(Expression)
	case *LetStatement:
		node.Value, _ = Modify(node.Value, modifier).(Expression)
	case *FunctionLiteral:
		for i, p := range node.Parameters {
			node.Parameters[i], _ = Modify(p, modifier).(*Identifier)
		}
		node.Body, _ = Modify(node.Body, modifier).(*BlockStatement)
	case *ArrayLiteral:
		for i, e := range node.Elements {
			node.Elements[i], _ = Modify(e, modifier).(Expression)
		}
	case *HashLiteral:
		nPairs := make(ExpressionPairs)
		for k, v := range node.Pairs {
			nk, _ := Modify(k, modifier).(Expression)
			nv, _ := Modify(v, modifier).(Expression)
			nPairs[nk] = nv
		}
		node.Pairs = nPairs
	}

	return modifier(node)
}
