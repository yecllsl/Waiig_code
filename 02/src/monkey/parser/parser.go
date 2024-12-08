package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type (
	prefixParseFn func() ast.Expression               //这种类型的函数通常用于解析前缀表达式。前缀表达式是指操作符位于操作数之前的表达式形式，例如-5或!true。
	infixParseFn  func(ast.Expression) ast.Expression //这种类型的函数用于解析中缀表达式。中缀表达式是指操作符位于两个操作数之间的表达式形式，例如2 + 3或var1 == var2。
)

// Parser 结构体定义了一个解析器，包含一个词法分析器、错误列表、当前词法单元和下一个词法单元，以及前缀和后缀解析函数的映射
type Parser struct {
	l      *lexer.Lexer // 词法分析器
	errors []string     // 错误列表

	curToken  token.Token // 当前词法单元
	peekToken token.Token // 下一个词法单元

	prefixParseFns map[token.TokenType]prefixParseFn // 前缀解析函数的映射
	infixParseFns  map[token.TokenType]infixParseFn  // 后缀解析函数的映射
}

// New函数用于创建一个新的Parser实例
func New(l *lexer.Lexer) *Parser {
	// 创建一个新的Parser实例
	p := &Parser{
		l:      l,          // 将传入的lexer实例赋值给Parser的l字段
		errors: []string{}, // 初始化错误列表
	}

	// 初始化前缀解析函数映射
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// 注册前缀解析函数
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	// 初始化中缀解析函数映射
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	// 注册中缀解析函数
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken 移动解析器的当前令牌和窥视令牌。
// 该方法将当前令牌设置为之前的窥视令牌，并从词法分析器获取下一个令牌作为新的窥视令牌。
// 这有助于解析器在处理令牌流时向前移动。
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs 检查当前解析器的令牌是否与给定的令牌类型匹配。
// 这个函数用于在解析过程中进行语法检查，确保解析的正确性。
// 参数:
//
//	t token.TokenType: 需要比较的令牌类型。
//
// 返回值:
//
//	bool: 如果当前令牌与给定的令牌类型匹配，则返回true，否则返回false。
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs 检查当前解析器的下一个记号是否为指定的类型
// 参数:
//
//	t: 待比较的记号类型
//
// 返回值:
//
//	bool: 如果下一个记号的类型与指定的类型相同，则返回true，否则返回false
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek 检查当前的 peekToken 是否与预期的 token 类型匹配。
// 如果匹配，则移动到下一个 token 并返回 true。
// 如果不匹配，则记录错误并返回 false。
// 此函数用于在解析过程中验证预期的 token，以确保解析的正确性。
// 其实就是语法分析器都有的-断言函数，用来检查语法分析器是否按照预期进行。
func (p *Parser) expectPeek(t token.TokenType) bool {
	// 检查 peekToken 是否与预期的 token 类型匹配
	if p.peekTokenIs(t) {
		// 如果匹配，则移动到下一个 token 并返回 true
		p.nextToken()
		return true
	} else {
		// 如果不匹配，则记录错误并返回 false
		p.peekError(t)
		return false
	}
}

/*
	Parser中现在有一个errors字段，这是一个字符串切片。该字段会在New中初始化，

当peekToken的类型与预期不符时，它会使用辅助函数peekError向errors中添加错误信息。
有了Errors方法就可以检查语法分析器是否遇到了错误。
*/
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError 添加一个错误信息到解析器的错误列表中，当解析器期望下一个令牌是特定类型时，但实际类型与期望不符。
// 这个函数主要用于解析过程中错误处理，它通过比较期望的令牌类型和实际的令牌类型来生成一个错误信息。
// 参数:
//
//	t (token.TokenType): 期望的令牌类型。
func (p *Parser) peekError(t token.TokenType) {
	// 格式化错误信息，说明期望的令牌类型和实际得到的令牌类型。
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	// 将生成的错误信息添加到解析器的错误列表中。
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// ParseProgram 解析整个程序，返回一个抽象语法树（AST）表示的程序结构。
// 该函数初始化一个程序节点，并逐步解析每个语句，直到遇到文件末尾（EOF）。
func (p *Parser) ParseProgram() *ast.Program {
	// 初始化一个空的程序节点,构造AST的根节点。
	program := &ast.Program{}
	// 初始化程序的语句列表为空。
	program.Statements = []ast.Statement{}

	// 循环解析语句，直到达到文件末尾。
	for !p.curTokenIs(token.EOF) {
		// 解析当前的语句。
		stmt := p.parseStatement()
		// 如果解析出的语句不为空，则添加到程序的语句列表中。
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// 移动到下一个令牌。
		p.nextToken()
	}

	// 返回解析完成的程序节点。
	return program
}

// parseStatement 根据当前词法分析器的状态解析并返回一个语句节点。
// 该函数通过检查当前词法符号的类型来决定解析哪种类型的语句。
// 参数: 无
// 返回值: ast.Statement 接口类型的语句节点。
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		// 当前词法符号类型为 LET 时，解析 let 语句。
		return p.parseLetStatement()
	case token.RETURN:
		// 当前词法符号类型为 RETURN 时，解析 return 语句。
		return p.parseReturnStatement()
	default:
		// 当前词法符号类型不属于上述情况时，解析表达式语句。
		return p.parseExpressionStatement()
	}
}

// parseLetStatement 解析 let 语句，如 "let x = 5;"
// 该函数主要负责处理变量声明的语法解析，包括变量名和赋值表达式的解析。
func (p *Parser) parseLetStatement() *ast.LetStatement {
	// 初始化 LetStatement 结构体，并将当前 token 赋值给其 Token 字段。
	stmt := &ast.LetStatement{Token: p.curToken}

	// 检查下一张牌是否为标识符，如果不是，则返回 nil。
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// 解析变量名，将其存储为 Identifier 类型。
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 检查下一张牌是否为赋值符号，如果不是，则返回 nil。
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// 移动到下一个 token，开始解析赋值表达式。
	p.nextToken()

	// 解析表达式，赋值给 LetStatement 的 Value 字段。
	stmt.Value = p.parseExpression(LOWEST)

	// 如果下一张牌是分号，则移动到下一个 token，表示语句结束。
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// 返回解析完成的 LetStatement 结构体。
	return stmt
}

// parseReturnStatement 解析返回语句，返回一个ReturnStatement节点
// 该函数不接受任何参数
// 返回值是一个指向ast.ReturnStatement的指针，表示解析后的返回语句
// 该函数首先创建一个ReturnStatement实例，然后解析返回值表达式
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// 初始化ReturnStatement实例
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// 移动到下一个token，准备解析返回值
	p.nextToken()

	// 解析返回值表达式
	stmt.ReturnValue = p.parseExpression(LOWEST)

	// 如果下一个token是分号，说明语句结束，移动到下一个token
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// 返回解析后的ReturnStatement节点
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

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
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
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
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
