package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

// 定义了一组常量，用于表示不同操作符的优先级
const (
	_ int = iota //这里使用iota为这些常量设置逐个递增的数值。空白标识符_为0，其余的常量值是1到7
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// 定义一个precedences的map，用于存储token.TokenType和对应的优先级
var precedences = map[token.TokenType]int{
	// token.EQ和token.NOT_EQ的优先级为EQUALS
	token.EQ:     EQUALS,
	token.NOT_EQ: EQUALS,
	// token.LT和token.GT的优先级为LESSGREATER
	token.LT: LESSGREATER,
	token.GT: LESSGREATER,
	// token.PLUS和token.MINUS的优先级为SUM
	token.PLUS:  SUM,
	token.MINUS: SUM,
	// token.SLASH和token.ASTERISK的优先级为PRODUCT
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	// token.LPAREN的优先级为CALL
	token.LPAREN: CALL,
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

// noPrefixParseFnError 只是将格式化的错误消息添加到语法分析器的errors字段中
// 参数:
//
//	t - 找不到前缀解析函数的令牌类型。
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	// 格式化错误消息
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	// 将错误消息添加到错误列表中
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
		// 当前词法符号类型不属于上述情况时（只有let和return语句），解析表达式语句。
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

// parseExpressionStatement 解析表达式语句
// 表达式语句是程序中的一种基本语句，它由一个表达式构成
// 该函数主要负责创建表达式语句的节点，并解析其中的表达式
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	// 初始化表达式语句节点
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// 解析表达式，表达式的优先级由LOWEST常量指定
	stmt.Expression = p.parseExpression(LOWEST)

	// 如果下一个token是分号，则表明语句结束，移动到下一个token
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// 返回解析完成的表达式语句节点
	return stmt
}

// parseExpression 解析表达式，根据运算符的优先级递归地构建表达式的AST节点。
// 参数 precedence 表达式当前的优先级。
// 返回值是解析得到的表达式节点。
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 获取当前令牌类型的前缀解析函数。
	prefix := p.prefixParseFns[p.curToken.Type]
	// 如果前缀解析函数为空，则表示不支持当前令牌类型的解析，抛出错误。
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// 调用前缀解析函数解析当前令牌为表达式节点。
	leftExp := prefix()

	// 循环解析中缀表达式，直到遇到分号或当前优先级低于下一看见的令牌优先级。
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// 获取下一看见的令牌类型的中缀解析函数。
		infix := p.infixParseFns[p.peekToken.Type]
		// 如果中缀解析函数为空，则返回当前解析的表达式节点。
		if infix == nil {
			return leftExp
		}

		// 移动到下一看见的令牌。
		p.nextToken()

		// 调用中缀解析函数，将当前表达式节点作为参数，解析得到新的表达式节点。
		leftExp = infix(leftExp)
	}

	// 返回最终解析的表达式节点。
	return leftExp
}

// peekPrecedence 返回当前查看的token的优先级
// 当前函数用于在解析过程中确定操作符或关键字的优先级，以便正确地构建解析树
// 如果当前查看的token在precedences映射中没有对应的优先级，那么函数将返回LOWEST，表示最低优先级
func (p *Parser) peekPrecedence() int {
	// 检查当前查看的token类型在precedences映射中是否有对应的优先级
	if p, ok := precedences[p.peekToken.Type]; ok {
		// 如果有，返回对应的优先级
		return p
	}

	// 如果没有，返回最低优先级
	return LOWEST
}

// curPrecedence 返回当前标记类型的运算符优先级。
// 该方法通过查找 precedences 字典来获取当前标记类型的优先级值。
// 如果当前标记类型在字典中没有对应的优先级，则返回最低优先级 LOWEST。
func (p *Parser) curPrecedence() int {
	// 检查当前标记类型是否在 precedences 字典中定义
	if p, ok := precedences[p.curToken.Type]; ok {
		// 如果是，返回对应的优先级值
		return p
	}

	// 如果不是，返回最低优先级
	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析整数字面量。
// 该函数创建一个整数字面量节点，并尝试将当前记号的字面值解析为64位整数。
// 如果解析失败，它会记录一个错误并返回nil。
func (p *Parser) parseIntegerLiteral() ast.Expression {
	// 初始化整数字面量节点。
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// 尝试将当前记号的字面值解析为64位整数。
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		// 如果解析失败，格式化错误信息并添加到错误列表中。
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	// 将成功解析的整数值赋给字面量节点。
	lit.Value = value

	// 返回整数字面量节点。
	return lit
}

// parsePrefixExpression 解析前缀表达式。
// 该函数创建一个前缀表达式结构体，设置当前令牌为表达式的令牌，
// 并将当前令牌的字面值作为操作符。然后，获取下一个令牌，并解析
// 表达式的右侧部分。这适用于处理如“-5”或“!true”等前缀表达式。
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression 解析中缀表达式，如加法、乘法等。
// 它根据当前解析的左表达式和当前的令牌构建一个中缀表达式结构体。
// 参数:
//   - left: 左表达式，即中缀表达式的左侧操作数。
//
// 返回值:
//   - ast.Expression: 解析后的中缀表达式。
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// 创建一个中缀表达式结构体，设置其基本属性。
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// 获取当前运算符的优先级。
	precedence := p.curPrecedence()
	// 移动到下一个令牌。
	p.nextToken()
	// 解析右侧表达式，使用当前运算符的优先级来决定解析的深度。
	expression.Right = p.parseExpression(precedence)

	// 返回构建好的中缀表达式。
	return expression
}

// parseBoolean 解析布尔表达式。
//
// 该函数创建并返回一个布尔表达式的AST节点，表示一个布尔值。
// 它使用当前的标记作为布尔表达式的标记，并通过检查当前标记是否为token.TRUE来确定布尔值。
// 如果当前标记是token.TRUE，则返回的布尔表达式的Value字段为true；否则为false。
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseGroupedExpression 解析分组表达式，如圆括号包围的表达式。
// 这个函数主要用于处理优先级最高的表达式解析。
func (p *Parser) parseGroupedExpression() ast.Expression {
	// 移动到下一个token，即左圆括号后的第一个token。
	p.nextToken()

	// 解析当前的表达式，由于分组表达式内部的表达式优先级最低，
	// 所以这里使用LOWEST优先级进行解析。
	exp := p.parseExpression(LOWEST)

	// 检查下一个token是否为右圆括号，如果不是，则解析失败，返回nil。
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 成功解析完分组表达式，返回解析得到的表达式。
	return exp
}

// parseIfExpression 解析if表达式
// 该函数负责解析if语句，包括条件、执行后果和可选的else部分
func (p *Parser) parseIfExpression() ast.Expression {
	// 初始化IfExpression结构体，并将当前token作为标识
	expression := &ast.IfExpression{Token: p.curToken}

	// 检查是否有一个左括号跟随在if后面
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// 移动到下一个token，并解析条件表达式
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// 确保条件表达式以右括号结束
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 检查是否有一个左大括号跟随在条件表达式后面
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析if条件下的执行后果代码块
	expression.Consequence = p.parseBlockStatement()

	// 检查是否有else部分
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		// 确保else部分以左大括号开始
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		// 解析else部分的执行后果代码块
		expression.Alternative = p.parseBlockStatement()
	}

	// 返回解析完成的if表达式
	return expression
}

// parseBlockStatement 解析代码块语句
//
// 该函数负责解析一个代码块，代码块以当前解析器位置的花括号（{）开始，
// 并一直解析到遇到右花括号（}）或文件结束（EOF）为止。在代码块内，
// 该函数会尝试解析每一个语句，并将其添加到代码块的语句列表中。
//
// 返回值是一个指向ast.BlockStatement的指针，表示解析完成的代码块。
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	// 初始化一个BlockStatement结构体，并将当前解析器位置的Token与其关联
	block := &ast.BlockStatement{Token: p.curToken}
	// 初始化代码块的语句列表为空
	block.Statements = []ast.Statement{}

	// 移动到代码块内的第一个Token
	p.nextToken()

	// 循环解析语句，直到遇到右花括号或文件结束
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// 尝试解析当前的语句
		stmt := p.parseStatement()
		// 如果解析成功，则将语句添加到代码块的语句列表中
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		// 移动到下一个Token
		p.nextToken()
	}

	// 返回解析完成的代码块
	return block
}

// parseFunctionLiteral 解析函数字面量。
// 该函数首先检查当前标记是否为左括号，然后解析函数的参数，
// 接着检查下一个标记是否为左大括号，最后解析函数的主体。
// 如果解析过程中遇到预期的标记不符合，则返回nil。
// 返回值是解析完成的函数字面量，如果解析失败则返回nil。
func (p *Parser) parseFunctionLiteral() ast.Expression {
	// 初始化函数字面量，并将当前标记作为函数字面量的标记。
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// 检查下一个标记是否为左括号，如果不是则返回nil。
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// 解析函数的参数。
	lit.Parameters = p.parseFunctionParameters()

	// 检查下一个标记是否为左大括号，如果不是则返回nil。
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析函数的主体。
	lit.Body = p.parseBlockStatement()

	// 返回解析完成的函数字面量。
	return lit
}

// parseFunctionParameters 解析函数的参数列表。
// 该函数在解析遇到右括号 ')' 时结束。
// 如果参数列表为空，则直接返回空的标识符列表。
// 如果解析过程中遇到预期之外的令牌，则返回nil。
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// 如果下一个令牌是右括号')'，表明参数列表为空，直接返回。
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// 处理第一个参数。
	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// 继续处理其余的参数，直到遇到的下一个令牌不是逗号','。
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// 期望下一个令牌是右括号')'，如果不是，则表明参数列表解析失败，返回nil。
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 返回解析到的参数列表。
	return identifiers
}

// parseCallExpression 解析调用表达式
//
// 参数:
//
//	function ast.Expression: 被调用的函数表达式
//
// 返回值:
//
//	ast.Expression: 解析后的调用表达式
//
// 该函数创建并返回一个调用表达式对象，同时解析并设置其参数
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// 创建调用表达式对象，设置当前令牌和被调用的函数
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	// 解析并设置调用表达式的参数
	exp.Arguments = p.parseCallArguments()
	return exp
}

// parseCallArguments 解析函数调用的参数列表。
// 该函数在解析到函数调用时被调用，以解析括号内的参数表达式。
// 返回值是解析到的表达式数组，如果解析失败或没有参数，则返回空数组或nil。
func (p *Parser) parseCallArguments() []ast.Expression {
	// 初始化参数列表为空数组。
	args := []ast.Expression{}

	// 如果下一个token是右括号，说明没有参数，直接返回空数组。
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	// 解析第一个参数表达式。
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	// 如果下一个token是逗号，说明还有更多参数。
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()                                  // 移动到逗号
		p.nextToken()                                  // 移动到下一个参数的开始
		args = append(args, p.parseExpression(LOWEST)) // 解析下一个参数表达式
	}

	// 如果下一个token不是右括号，说明参数列表解析失败，返回nil。
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 返回解析到的参数列表。
	return args
}

// registerPrefix 为给定的 tokenType 注册一个前缀解析函数。
// 当解析器遇到对应类型的令牌时，将使用该函数来解析。
// 参数:
//
//	tokenType - 需要注册解析函数的令牌类型。
//	fn - 对应令牌类型的前缀解析函数。
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix 用于注册中缀操作符的解析函数。
// 参数 tokenType 是中缀操作符的类型，fn 是对应的解析函数。
// 这个方法将解析函数与特定的中缀操作符类型关联起来，使得解析器在遇到相应的操作符时能够调用正确的解析逻辑。
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
