package parser

import "github.com/avinassh/monkey/token"

// checks if the current token is `t`
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// checks if the peek token is `t`
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// is the next token `t`? if yes, then move the parser to next
// else throw error
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		return false
	}
}
