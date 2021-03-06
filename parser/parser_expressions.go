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

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken, Parameters: []*ast.Identifier{}}

	// currently we are at token `fn`. Next token should be `(`, so, we will peek and
	// if not we will return
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fn.Parameters = p.parseFunctionParameters()

	// we are now at `)`, so we will peek if the next token is `{`
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// currently at `{`, so lets move and start parsing fn block
	fn.Body = p.parseBlockStatement()
	// now the token will be at `}`

	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	// currently we are at `(`. If the next token is `)`, then this
	// function has no parameters. So we will move next and return
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// we are at `(`, so we will move next and next token should be an
	// identifier
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// our algorithm changes if there are one parameter or more than one.
	// if its one, we will consume it first
	// then we will check for next ones in a for loop, till the next token is
	// comma
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// lets say we have three params: a, b, c
	// `a` would have been consumed by previous statements. Next, will check if the
	// peek is comma, if that is so, then we know that there should be one more
	// identifier. So, we will move twice in the for loop. One move, will move us from
	// `a` to `,`, the second move is `,` to `b`. Now the current token is `b` which
	// is an identifier and we will consume it. And we will repeat the loop!
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// once all the identifiers have been consumed, the last token will be
	// `)`
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	callExp := &ast.CallExpression{Token: p.curToken, Function: fn,
		Arguments: p.parseExpressionList(token.RPAREN)}
	return callExp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	callExp := &ast.IndexExpression{Token: p.curToken, Left: left}

	// currently we are at `[`, we will move one step
	// and the current position will be at the start of index
	//
	// e.g. myArray[1 + 1] , after the following call, current token will be at `1`
	p.nextToken()

	callExp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return callExp
}

func (p *Parser) parseExpressionList(endToken token.TokenType) []ast.Expression {
	var args []ast.Expression

	// currently we are at `(`, if the next immediate token is `)`,
	// there are no args to parse
	if p.peekTokenIs(endToken) {
		p.nextToken()
		return args
	}

	// so we will move from current token `(`
	p.nextToken()
	//  and parse the expression
	args = append(args, p.parseExpression(LOWEST))

	// then we will parse the rest of the expressions like how we
	// did in `parseFunctionParameters`
	for p.peekTokenIs(token.COMMA) {
		// to move to `,`
		p.nextToken()
		// to move to the token of expression from the `,`
		p.nextToken()

		//  and then we will parse the expression
		args = append(args, p.parseExpression(LOWEST))
	}

	// at the end, we should have an `)`
	if !p.expectPeek(endToken) {
		return nil
	}

	return args
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	return &ast.ArrayLiteral{
		Token:    p.curToken,
		Elements: p.parseExpressionList(token.RBRACKET),
	}
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hl := &ast.HashLiteral{Token: p.curToken, Pairs: map[ast.Expression]ast.Expression{}}

	// currently we are at `{`, if the next immediate token is `}`,
	// then this is an empty hash
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hl
	}

	// so we will move from current token `{`
	p.nextToken()

	// next we will parse everything in a for loop
	for {
		//  and parse the expression, which will be our key
		key := p.parseExpression(LOWEST)

		// format is, <expression> : <expression>
		// so next token should be :
		if !p.expectPeek(token.COLON) {
			return nil
		}

		// so we will move from current token `:`
		p.nextToken()

		// now will parse the value side of expression
		value := p.parseExpression(LOWEST)

		hl.Pairs[key] = value

		// we have parsed the value. Now what to do next depends on whats the next token.
		// if the next token is `}`, then it has only one kv pair. So we will parse
		// that and return
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return hl
		}

		// so the next token has to is `,`, then it has multiple kv pairs
		// so will move from `,` and next will be the key of next kv pair
		if !p.expectPeek(token.COMMA) {
			return nil
		}
		// currently, we are comma, so we will move to next and we will be at
		// the start of our next key
		p.nextToken()
	}
}
