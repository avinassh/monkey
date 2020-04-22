package parser

import (
	"github.com/avinassh/monkey/ast"
	"github.com/avinassh/monkey/token"
)

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{
		Token: p.curToken,
	}
	stmt.Expression = p.parseExpression(LOWEST)
	// from book:
	// we check for an optional semicolon. Yes, it’s optional. If the
	// peekToken is a token.SEMICOLON, we advance so it’s the curToken.
	// If it’s not there, that’s okay too, we don’t add an error to the
	// parser if it’s not there. That’s because we want expression
	// statements to have optional semicolons (which makes it easier to
	// type something like 5 + 5 into the REPL later on).
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expr.Right = p.parseExpression(PREFIX)

	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	ifExp := &ast.IfExpression{Token: p.curToken}

	// current token is `if`, so will move next which would be `( condition )`
	// so we will verify if the peek is `(` and move from `if` to `(`
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// move next from `(`
	p.nextToken()

	// parse the `condition`
	ifExp.Condition = p.parseExpression(LOWEST)

	// we have parsed the condition and at the last token of condition
	// so peek token should be `)`. If so, we will move to it
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	// we have parsed the condition, currently we are at `(`
	// after the condition, next token should be `{`, so we will peek and move
	// if not we will return
	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}

	// we will parse the statement block. Current token is now at `{`
	ifExp.Consequence = p.parseBlockStatement()

	// currently we are at `}`. So, do we have an ELSE later?
	// if the peek token is else, then we have an alternate block which we need to
	// execute
	if p.peekTokenIs(token.ELSE) {

		// we are at `}`, so we move to `ELSE`
		p.nextToken()

		// we are at `else` currently
		// after else, we should have an `{`
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// we will parse the statement block. Current token is now at `{`
		ifExp.Alternative = p.parseBlockStatement()
		// now the token will be at `}`
	}

	return ifExp
}

// from book:
// the tokens get advanced just enough so that parseBlockStatement sits on the { with p.curToken
// being of type token.LBRACE.
//
// so we start the block statements with `{` at start. So we need to call p.nextToken() to
// to start with the statements
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	bs := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if stmt := p.parseStatement(); stmt != nil {
			bs.Statements = append(bs.Statements, stmt)
		}
		p.nextToken()
	}
	return bs
}

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
