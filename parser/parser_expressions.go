package parser

import (
	"github.com/avinassh/monkey/ast"
	"github.com/avinassh/monkey/token"
)

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		Token: p.curToken,
	}
	stmt.Expression = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
