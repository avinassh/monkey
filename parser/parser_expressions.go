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
	ifExp := &ast.IfExpression{
		Token: token.Token{
			Type:    p.curToken.Type,
			Literal: p.curToken.Literal,
		},
	}

	// current token is `if`, so will move next which would be `( expression )`
	p.nextToken()

	// parse the `expression`
	ifExp.Condition = p.parseExpression(LOWEST)

	// we have parsed the condition, currently we are at `(`
	// after the condition, next token should be `{`, so we will peek and move
	// if not we will return
	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}

	// we are at `{`, so we will move next which would start the block
	p.nextToken()

	consequence := &ast.BlockStatement{}
	for !p.curTokenIs(token.RBRACE) {
		if stmt := p.parseStatement(); stmt != nil {
			consequence.Statements = append(consequence.Statements, stmt)
		}
		p.nextToken()
	}

	ifExp.Consequence = consequence

	// currently we are at `}`
	// so we will move next and complete the if block
	p.nextToken()

	// if the next token is else, then we have an alternate block which we need to
	// execute
	if !p.curTokenIs(token.ELSE) {
		return ifExp
	}

	// we are at `else` currently
	// after else, we should have an `{`
	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}

	// we are at `{`
	p.nextToken()

	alternative := &ast.BlockStatement{}
	for !p.curTokenIs(token.RBRACE) {
		if stmt := p.parseStatement(); stmt != nil {
			alternative.Statements = append(alternative.Statements, stmt)
		}
		p.nextToken()
	}

	ifExp.Alternative = alternative
	return ifExp
}
