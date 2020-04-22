// contains alternate and simplified methods
package parser

import (
	"github.com/avinassh/monkey/ast"
	"github.com/avinassh/monkey/token"
)

// alternate to `parseFunctionLiteral`
func (p *Parser) parseFnExpression() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken, Parameters: []*ast.Identifier{}}

	// currently we are at token `fn`. Next token should be `(`, so, we will peek and
	// if not we will return
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// currently we are at `(`, so will move next
	p.nextToken()

	for !p.curTokenIs(token.RPAREN) {
		if !p.curTokenIs(token.IDENT) {
			return nil
		}
		fn.Parameters = append(fn.Parameters, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	// we are now at `)`, so we will peek if the next token is `{`
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// currently at `{`, so lets move and start parsing fn block
	fn.Body = p.parseBlockStatement()
	// now the token will be at `}`

	return fn
}
