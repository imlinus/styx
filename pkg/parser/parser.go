package parser

import (
	"fmt"
	"strconv"
	"styx/pkg/ast"
	"styx/pkg/lexer"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	EQUALS      // == or ~=
	LESSGREATER // > or <
	SUM         // + or ..
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X) or obj[X] or obj.X
	AND         // and
	OR          // or
)

var precedences = map[lexer.TokenType]int{
	lexer.ASSIGN:   ASSIGN,
	lexer.OR:       OR,
	lexer.AND:      AND,
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.GT:       LESSGREATER,
	lexer.GE:       LESSGREATER,
	lexer.LT:       LESSGREATER,
	lexer.LE:       LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.ASTERISK: PRODUCT,
	lexer.SLASH:    PRODUCT,
	lexer.CONCAT:   SUM,
	lexer.LPAREN:   CALL,
	lexer.LBRACKET: CALL,
	lexer.DOT:      CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.TRUE, p.parseBoolean)
	p.registerPrefix(lexer.FALSE, p.parseBoolean)
	p.registerPrefix(lexer.NULL, p.parseNull)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TEMPLATE, p.parseTemplateLiteral)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.IF, p.parseIfExpression)
	p.registerPrefix(lexer.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseHashLiteral)
	p.registerPrefix(lexer.LBRACKET, p.parseArrayLiteral)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.GE, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.CONCAT, p.parseInfixExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseIndexExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.LET:
		return p.parseLetStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.LOOP:
		return p.parseLoopStatement()
	case lexer.BREAK:
		return &ast.BreakStatement{Token: p.curToken}
	case lexer.CONTINUE:
		return &ast.ContinueStatement{Token: p.curToken}
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLoopStatement() *ast.LoopStatement {
	stmt := &ast.LoopStatement{Token: p.curToken}
	p.nextToken()

	// Peek to see if it's an iteration loop or condition loop
	if p.curTokenIs(lexer.IDENT) && (p.peekTokenIs(lexer.IN) || p.peekTokenIs(lexer.COMMA)) {
		// Looping with identifiers: loop i, v in hash
		stmt.Key = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			stmt.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}

		if !p.expectPeek(lexer.IN) {
			return nil
		}

		p.nextToken()
		stmt.Iter = p.parseExpression(LOWEST)
	} else if !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.END) {
		// Could be a condition: loop x < 10
		// We peek to make sure we're not starting a block immediately
		stmt.Condition = p.parseExpression(LOWEST)
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseImportStatement() ast.Statement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	// Could be 'import name' or 'import name from "path"'
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(lexer.FROM) {
		p.nextToken() // skip to 'from'
		if !p.expectPeek(lexer.STRING) {
			return nil
		}
		stmt.Name = name
		stmt.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	} else {
		// Just 'import name' (legacy behavior for stdlib)
		stmt.Name = name
		// For stdlib imports, we don't have a path yet
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.ELSE) && !p.curTokenIs(lexer.ELSEIF) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// Use peekPrecedence() which defaults to LOWEST if not in the map
	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
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

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parseNull() ast.Expression {
	return &ast.Null{Token: p.curToken}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseTemplateLiteral() ast.Expression {
	lit := &ast.TemplateLiteral{Token: p.curToken}
	raw := p.curToken.Literal

	for len(raw) > 0 {
		idx := indexOf(raw, "${")
		if idx == -1 {
			lit.Parts = append(lit.Parts, &ast.StringLiteral{
				Token: p.curToken,
				Value: raw,
			})
			break
		}

		// Text before ${...
		if idx > 0 {
			lit.Parts = append(lit.Parts, &ast.StringLiteral{
				Token: p.curToken,
				Value: raw[:idx],
			})
		}

		// Find matching }
		end := indexOf(raw[idx+2:], "}")
		if end == -1 {
			p.errors = append(p.errors, "unterminated template expression")
			return nil
		}

		exprStr := raw[idx+2 : idx+2+end]

		// Parse the expression
		exprLexer := lexer.New(exprStr)
		exprParser := New(exprLexer)
		expr := exprParser.parseExpression(LOWEST)

		if len(exprParser.Errors()) > 0 {
			p.errors = append(p.errors, exprParser.Errors()...)
			return nil
		}

		lit.Parts = append(lit.Parts, expr)
		raw = raw[idx+2+end+1:]
	}

	return lit
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.THEN) {
		p.nextToken()
	}

	expression.Consequence = p.parseBlockStatement()

	curr := expression
	for p.curTokenIs(lexer.ELSEIF) {
		p.nextToken() // consume 'elseif'

		elseifExp := &ast.IfExpression{Token: p.curToken}
		elseifExp.Condition = p.parseExpression(LOWEST)

		if p.peekTokenIs(lexer.THEN) {
			p.nextToken()
		}

		elseifExp.Consequence = p.parseBlockStatement()
		curr.Alternative = &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ExpressionStatement{Expression: elseifExp},
			},
		}
		curr = elseifExp
	}

	if p.curTokenIs(lexer.ELSE) {
		curr.Alternative = p.parseBlockStatement()
	}

	if p.curTokenIs(lexer.END) {
	}

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		lit.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	expression := &ast.AssignmentExpression{
		Token: p.curToken,
		Left:  left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Value = p.parseExpression(precedence - 1)

	return expression
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(lexer.RBRACE) && !p.expectPeek(lexer.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(lexer.RBRACKET)
	return array
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	if p.curTokenIs(lexer.LBRACKET) {
		p.nextToken()
		exp.Index = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	} else if p.curTokenIs(lexer.DOT) {
		p.nextToken()
		// For obj.key, the index is simply the key identifier treated as a string or identifier
		// We'll parse it as an identifier (expression)
		exp.Index = p.parseIdentifier()
	}

	return exp
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
