package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

// 运算符优先级常量定义，使用iota从LOWEST开始递增
// 优先级越高，运算符绑定越紧密
const (
	_           int = iota
	LOWEST          // 最低优先级，用于基础表达式
	EQUALS          // == 和 != 运算符
	LESSGREATER     // > 和 < 运算符
	SUM             // + 和 - 运算符
	PRODUCT         // * 和 / 运算符
	PREFIX          // -X 或 !X 前缀运算符
	CALL            // myFunction(X) 函数调用
	INDEX           // array[index] 数组索引
)

// precedences 映射表定义了各种运算符的优先级
// 键为token类型，值为对应的优先级常量
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,      // == 运算符
	token.NOT_EQ:   EQUALS,      // != 运算符
	token.LT:       LESSGREATER, // < 运算符
	token.GT:       LESSGREATER, // > 运算符
	token.PLUS:     SUM,         // + 运算符
	token.MINUS:    SUM,         // - 运算符
	token.SLASH:    PRODUCT,     // / 运算符
	token.ASTERISK: PRODUCT,     // * 运算符
	token.LPAREN:   CALL,        // ( 函数调用
	token.LBRACKET: INDEX,       // [ 数组索引
}

// 解析函数类型定义
type (
	// prefixParseFn 前缀解析函数，处理前缀表达式（如标识符、字面量、前缀运算符）
	prefixParseFn func() ast.Expression

	// infixParseFn 中缀解析函数，处理中缀表达式（如二元运算符）
	infixParseFn func(ast.Expression) ast.Expression
)

// Parser 结构体表示Monkey语言的语法分析器
// 负责将词法分析器生成的token序列转换为抽象语法树(AST)
type Parser struct {
	l      *lexer.Lexer // 词法分析器实例
	errors []string     // 解析过程中收集的错误信息

	curToken  token.Token // 当前处理的token
	peekToken token.Token // 下一个token（前瞻token）

	// 前缀解析函数映射表，根据token类型调用对应的解析函数
	prefixParseFns map[token.TokenType]prefixParseFn
	// 中缀解析函数映射表，根据token类型调用对应的解析函数
	infixParseFns map[token.TokenType]infixParseFn
}

// New 创建并初始化一个新的语法分析器
// 参数 l: 词法分析器实例
// 返回值: 初始化完成的Parser指针
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// 初始化前缀解析函数映射表
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)         // 标识符解析
	p.registerPrefix(token.INT, p.parseIntegerLiteral)       // 整数字面量解析
	p.registerPrefix(token.STRING, p.parseStringLiteral)     // 字符串字面量解析
	p.registerPrefix(token.BANG, p.parsePrefixExpression)    // ! 前缀运算符
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)   // - 前缀运算符
	p.registerPrefix(token.TRUE, p.parseBoolean)             // true布尔值
	p.registerPrefix(token.FALSE, p.parseBoolean)            // false布尔值
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression) // 分组表达式 (expr)
	p.registerPrefix(token.IF, p.parseIfExpression)          // if条件表达式
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral) // 函数字面量
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)    // 数组字面量
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)       // 哈希字面量

	// 初始化中缀解析函数映射表
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)     // + 中缀运算符
	p.registerInfix(token.MINUS, p.parseInfixExpression)    // - 中缀运算符
	p.registerInfix(token.SLASH, p.parseInfixExpression)    // / 中缀运算符
	p.registerInfix(token.ASTERISK, p.parseInfixExpression) // * 中缀运算符
	p.registerInfix(token.EQ, p.parseInfixExpression)       // == 中缀运算符
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)   // != 中缀运算符
	p.registerInfix(token.LT, p.parseInfixExpression)       // < 中缀运算符
	p.registerInfix(token.GT, p.parseInfixExpression)       // > 中缀运算符

	p.registerInfix(token.LPAREN, p.parseCallExpression)    // 函数调用
	p.registerInfix(token.LBRACKET, p.parseIndexExpression) // 数组索引

	// 读取前两个token，初始化curToken和peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken 前进到下一个token
// 将peekToken设置为当前token，然后读取新的peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs 检查当前token是否为指定类型
// 参数 t: 要检查的token类型
// 返回值: 如果当前token类型匹配则返回true
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs 检查下一个token是否为指定类型
// 参数 t: 要检查的token类型
// 返回值: 如果下一个token类型匹配则返回true
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek 期望下一个token为指定类型，如果是则前进
// 参数 t: 期望的token类型
// 返回值: 如果匹配则前进并返回true，否则记录错误并返回false
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// Errors 返回解析过程中收集的所有错误信息
// 返回值: 错误字符串切片
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError 记录下一个token类型不匹配的错误
// 参数 t: 期望的token类型
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError 记录没有找到前缀解析函数的错误
// 参数 t: 当前token类型
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// ParseProgram 解析整个程序，生成抽象语法树
// 返回值: 表示整个程序的Program节点
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// 循环解析所有语句，直到遇到EOF
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement 根据当前token类型解析对应的语句
// 返回值: 解析出的语句节点
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement() // let语句
	case token.RETURN:
		return p.parseReturnStatement() // return语句
	default:
		return p.parseExpressionStatement() // 表达式语句
	}
}

// parseLetStatement 解析let语句：let <identifier> = <expression>;
// 返回值: LetStatement节点，如果解析失败返回nil
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// 期望下一个token是标识符
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// 解析标识符名称
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// 期望下一个token是赋值运算符
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	// 解析赋值表达式
	stmt.Value = p.parseExpression(LOWEST)

	// 可选的分号
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement 解析return语句：return <expression>;
// 返回值: ReturnStatement节点
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// 解析返回值表达式
	stmt.ReturnValue = p.parseExpression(LOWEST)

	// 可选的分号
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement 解析表达式语句：<expression>;
// 返回值: ExpressionStatement节点
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// 解析表达式
	stmt.Expression = p.parseExpression(LOWEST)

	// 可选的分号
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpression 使用Pratt解析算法解析表达式
// 参数 precedence: 当前优先级，控制运算符绑定
// 返回值: 解析出的表达式节点
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 获取当前token对应的前缀解析函数
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// 循环处理中缀表达式，直到遇到分号或优先级不足
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

// peekPrecedence 获取下一个token的优先级
// 返回值: 下一个token的优先级，如果未定义则返回LOWEST
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence 获取当前token的优先级
// 返回值: 当前token的优先级，如果未定义则返回LOWEST
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

// parseIdentifier 解析标识符表达式
// 返回值: Identifier节点
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析整数字面量表达式
// 返回值: IntegerLiteral节点，如果解析失败返回nil
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// 将字符串转换为int64
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

// parseStringLiteral 解析字符串字面量表达式
// 返回值: StringLiteral节点
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parsePrefixExpression 解析前缀表达式（如!true, -5）
// 返回值: PrefixExpression节点
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	// 解析右侧表达式，使用PREFIX优先级
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression 解析中缀表达式（如a + b, x == y）
// 参数 left: 左侧表达式节点
// 返回值: InfixExpression节点
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// 获取当前运算符的优先级
	precedence := p.curPrecedence()
	p.nextToken()
	// 解析右侧表达式，使用当前优先级
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseBoolean 解析布尔值表达式
// 返回值: Boolean节点，值为当前token是否为TRUE
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseGroupedExpression 解析分组表达式（如(1 + 2) * 3）
// 返回值: 分组内的表达式节点
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	// 解析括号内的表达式
	exp := p.parseExpression(LOWEST)

	// 期望右括号
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression 解析if条件表达式
// 返回值: IfExpression节点
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	// 期望左括号
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	// 解析条件表达式
	expression.Condition = p.parseExpression(LOWEST)

	// 期望右括号
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 期望左花括号
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析条件为真时的语句块
	expression.Consequence = p.parseBlockStatement()

	// 处理可选的else分支
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		// 期望左花括号
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// 解析else分支的语句块
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseBlockStatement 解析语句块（由花括号包围的语句序列）
// 返回值: BlockStatement节点
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// 循环解析语句，直到遇到右花括号或EOF
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseFunctionLiteral 解析函数字面量表达式
// 返回值: FunctionLiteral节点
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// 期望左括号
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// 解析函数参数列表
	lit.Parameters = p.parseFunctionParameters()

	// 期望左花括号
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析函数体语句块
	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters 解析函数参数列表
// 返回值: 参数标识符切片
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// 处理空参数列表的情况
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	// 解析第一个参数
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// 循环解析逗号分隔的后续参数
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// 期望右括号
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// parseCallExpression 解析函数调用表达式
// 参数 function: 被调用的函数表达式
// 返回值: CallExpression节点
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	// 解析参数列表
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseExpressionList 解析表达式列表（如函数参数、数组元素）
// 参数 end: 列表结束的token类型
// 返回值: 表达式切片
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// 处理空列表的情况
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	// 解析第一个表达式
	list = append(list, p.parseExpression(LOWEST))

	// 循环解析逗号分隔的后续表达式
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	// 期望结束token
	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseArrayLiteral 解析数组字面量表达式
// 返回值: ArrayLiteral节点
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	// 解析数组元素列表
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseIndexExpression 解析数组索引表达式
// 参数 left: 数组表达式
// 返回值: IndexExpression节点
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	// 解析索引表达式
	exp.Index = p.parseExpression(LOWEST)

	// 期望右方括号
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseHashLiteral 解析哈希字面量表达式
// 返回值: HashLiteral节点
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	// 循环解析键值对，直到遇到右花括号
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		// 解析键表达式
		key := p.parseExpression(LOWEST)

		// 期望冒号分隔符
		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		// 解析值表达式
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		// 处理逗号分隔或结束
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	// 期望右花括号
	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

// registerPrefix 注册前缀解析函数
// 参数 tokenType: token类型
// 参数 fn: 对应的解析函数
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix 注册中缀解析函数
// 参数 tokenType: token类型
// 参数 fn: 对应的解析函数
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
