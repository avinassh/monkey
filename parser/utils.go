package parser

import (
	"fmt"
	"github.com/avinassh/monkey/token"
)

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
//
// from the book:
// The expectPeek method is one of the “assertion functions” nearly
// all parsers share. Their primary purpose is to enforce the correctness
// of the order of tokens by checking the type of the next token. Our
// expectPeek here checks the type of the peekToken and only if the
// type is correct does it advance the tokens by calling nextToken.
// As you’ll see, this is something a parser does a lot.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}
