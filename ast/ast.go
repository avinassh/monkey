package ast

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	nodeStatement()
}

type Expression interface {
	Node
	nodeExpression()
}

// this is root AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}
