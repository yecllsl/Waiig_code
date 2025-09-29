package lexer

import "monkey/token"

// Lexer 结构体是 Monkey 编程语言的词法分析器
// 它负责将源代码字符串转换为一系列 Token
type Lexer struct {
	// input 存储要分析的源代码字符串
	input string

	// position 是当前字符在输入字符串中的位置（指向当前正在检查的字符）
	position int // current position in input (points to current char)

	// readPosition 是下一个要读取的字符位置（在当前字符之后）
	readPosition int // current reading position in input (after current char)

	// ch 是当前正在检查的字符
	ch byte // current char under examination
}

// New 函数是 Lexer 的构造函数，用于创建并初始化一个新的词法分析器实例
// 参数 input 是要分析的源代码字符串
// 返回值是一个指向新创建的 Lexer 结构体的指针
func New(input string) *Lexer {
	// 创建一个新的 Lexer 实例，并设置输入字符串
	l := &Lexer{input: input}

	// 调用 readChar 方法初始化词法分析器的状态
	// 这会设置 position、readPosition 和 ch 字段的初始值
	l.readChar()

	// 返回初始化完成的 Lexer 实例
	return l
}

// NextToken 方法是 Lexer 的核心方法，负责从输入字符串中读取并返回下一个 Token
// 该方法实现了词法分析的主要逻辑，通过逐个字符分析来识别不同的 Token 类型
// 返回值是一个 token.Token 结构体，包含 Token 的类型和字面量值
func (l *Lexer) NextToken() token.Token {
	// 创建一个空的 Token 变量，用于存储将要返回的 Token
	var tok token.Token

	// 首先跳过所有空白字符（空格、制表符、换行符等）
	// 确保从非空白字符开始分析
	l.skipWhitespace()

	// 使用 switch 语句根据当前字符进行分支处理
	// 每个 case 对应一种特定的字符或字符组合
	switch l.ch {
	case '=':
		// 处理赋值运算符 '=' 和相等比较运算符 '=='
		// 使用 peekChar() 方法查看下一个字符来判断是否为双字符运算符
		if l.peekChar() == '=' {
			// 如果是 '=='，则创建 EQ Token
			ch := l.ch
			l.readChar()                         // 读取下一个字符
			literal := string(ch) + string(l.ch) // 组合字面量 "=="
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			// 如果是单个 '='，则创建 ASSIGN Token
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		// 处理加法运算符 '+'
		tok = newToken(token.PLUS, l.ch)
	case '-':
		// 处理减法运算符 '-'
		tok = newToken(token.MINUS, l.ch)
	case '!':
		// 处理逻辑非运算符 '!' 和不等于运算符 '!='
		if l.peekChar() == '=' {
			// 如果是 '!='，则创建 NOT_EQ Token
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch) // 组合字面量 "!="
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			// 如果是单个 '!'，则创建 BANG Token
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		// 处理除法运算符 '/'
		tok = newToken(token.SLASH, l.ch)
	case '*':
		// 处理乘法运算符 '*'
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		// 处理小于运算符 '<'
		tok = newToken(token.LT, l.ch)
	case '>':
		// 处理大于运算符 '>'
		tok = newToken(token.GT, l.ch)
	case ';':
		// 处理分号 ';'
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		// 处理冒号 ':'
		tok = newToken(token.COLON, l.ch)
	case ',':
		// 处理逗号 ','
		tok = newToken(token.COMMA, l.ch)
	case '{':
		// 处理左花括号 '{'
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		// 处理右花括号 '}'
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		// 处理左圆括号 '('
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		// 处理右圆括号 ')'
		tok = newToken(token.RPAREN, l.ch)
	case '"':
		// 处理字符串字面量，以双引号开头
		// 调用 readString() 方法读取完整的字符串内容
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '[':
		// 处理左方括号 '['
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		// 处理右方括号 ']'
		tok = newToken(token.RBRACKET, l.ch)
	case 0:
		// 处理文件结束符（EOF）
		// 当 readPosition 超出输入字符串长度时，ch 被设置为 0
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		// 处理默认情况：标识符、数字或非法字符
		if isLetter(l.ch) {
			// 如果是字母或下划线，则读取标识符
			tok.Literal = l.readIdentifier()
			// 使用 LookupIdent 函数判断标识符是关键字还是普通标识符
			tok.Type = token.LookupIdent(tok.Literal)
			// 直接返回，因为 readIdentifier() 已经移动了位置指针
			return tok
		} else if isDigit(l.ch) {
			// 如果是数字，则读取数字字面量
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			// 直接返回，因为 readNumber() 已经移动了位置指针
			return tok
		} else {
			// 如果是无法识别的字符，则标记为非法 Token
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	// 对于大多数单字符 Token，读取下一个字符以准备分析下一个 Token
	// 注意：字符串、标识符和数字的处理已经在各自分支中返回，不会执行到这里
	l.readChar()

	// 返回分析得到的 Token
	return tok
}

// skipWhitespace 方法用于跳过输入字符串中的所有空白字符
// 空白字符包括：空格(' ')、制表符('\t')、换行符('\n')和回车符('\r')
// 该方法在词法分析过程中被调用，确保 Token 分析从非空白字符开始
func (l *Lexer) skipWhitespace() {
	// 使用 for 循环持续检查当前字符是否为空白字符
	// 循环条件：当前字符是空格、制表符、换行符或回车符中的任意一种
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		// 调用 readChar() 方法读取下一个字符
		// 这会移动 position 和 readPosition 指针，并更新 ch 为下一个字符
		l.readChar()
	}
	// 当遇到非空白字符时，循环结束，词法分析器准备分析下一个有意义的 Token
}

// readChar 方法是 Lexer 的核心字符读取方法，负责从输入字符串中读取下一个字符
// 该方法更新词法分析器的内部状态，包括当前字符、当前位置和下一个读取位置
// 当到达输入字符串末尾时，将当前字符设置为 0（EOF 标记）
func (l *Lexer) readChar() {
	// 检查是否已经到达或超过输入字符串的末尾
	// readPosition 表示下一个要读取的字符位置
	if l.readPosition >= len(l.input) {
		// 如果已经到达末尾，将当前字符 ch 设置为 0
		// 0 在词法分析中通常表示文件结束（EOF）
		l.ch = 0
	} else {
		// 如果还有字符可读，从输入字符串中读取下一个字符
		// 使用 readPosition 作为索引获取对应位置的字符
		l.ch = l.input[l.readPosition]
	}

	// 更新当前位置 position 为当前的读取位置 readPosition
	// position 现在指向刚刚读取的字符
	l.position = l.readPosition

	// 将读取位置 readPosition 向前移动一位，指向下一个要读取的字符
	// 这为下一次读取字符做好准备
	l.readPosition += 1
}

// peekChar 方法是 Lexer 的前瞻字符查看方法，用于查看下一个字符而不移动位置指针
// 该方法提供了一种"偷看"下一个字符的能力，用于判断双字符运算符（如"=="、"!="）
// 返回值是下一个字符的 byte 值，如果到达字符串末尾则返回 0
func (l *Lexer) peekChar() byte {
	// 检查是否已经到达或超过输入字符串的末尾
	// readPosition 表示下一个要读取的字符位置
	if l.readPosition >= len(l.input) {
		// 如果已经到达末尾，返回 0 表示文件结束（EOF）
		// 这确保了方法在边界情况下的安全性
		return 0
	} else {
		// 如果还有字符可读，返回下一个字符的值
		// 注意：这里只是返回字符值，不会移动任何位置指针
		// 这使得调用者可以查看下一个字符而不影响词法分析器的状态
		return l.input[l.readPosition]
	}
}

// readIdentifier 方法用于从输入字符串中读取一个完整的标识符
// 标识符由字母、下划线组成，用于表示变量名、函数名等
// 返回值是标识符的字符串表示
func (l *Lexer) readIdentifier() string {
	// 记录标识符的起始位置
	// position 字段记录了标识符开始的位置，用于后续提取子字符串
	position := l.position

	// 使用 for 循环持续读取字符，直到遇到非字母字符
	// isLetter 函数检查当前字符是否为字母或下划线
	for isLetter(l.ch) {
		// 调用 readChar() 方法读取下一个字符
		// 这会移动 position 和 readPosition 指针
		l.readChar()
	}

	// 使用字符串切片提取标识符
	// 从记录的起始位置 position 到当前的位置 l.position
	// 返回标识符的完整字符串
	return l.input[position:l.position]
}

// readNumber 方法用于从输入字符串中读取一个完整的数字字面量
// 数字字面量由数字字符（0-9）组成，用于表示整数值
// 返回值是数字的字符串表示
func (l *Lexer) readNumber() string {
	// 记录数字的起始位置
	// position 字段记录了数字开始的位置，用于后续提取子字符串
	position := l.position

	// 使用 for 循环持续读取字符，直到遇到非数字字符
	// isDigit 函数检查当前字符是否为数字（0-9）
	for isDigit(l.ch) {
		// 调用 readChar() 方法读取下一个字符
		// 这会移动 position 和 readPosition 指针
		l.readChar()
	}

	// 使用字符串切片提取数字字面量
	// 从记录的起始位置 position 到当前的位置 l.position
	// 返回数字的完整字符串表示
	return l.input[position:l.position]
}

// readString 方法用于从输入字符串中读取一个完整的字符串字面量
// 字符串字面量以双引号(")开头和结尾，包含任意字符序列
// 返回值是字符串内容的字符串表示（不包含两端的双引号）
func (l *Lexer) readString() string {
	// 记录字符串内容的起始位置
	// position + 1 跳过开头的双引号，直接指向字符串内容
	position := l.position + 1

	// 使用无限循环持续读取字符，直到遇到结束双引号或文件结束
	for {
		// 调用 readChar() 方法读取下一个字符
		// 这会移动 position 和 readPosition 指针
		l.readChar()

		// 检查是否遇到字符串结束标记
		// l.ch == '"' 表示遇到结束双引号
		// l.ch == 0 表示遇到文件结束（EOF），防止无限循环
		if l.ch == '"' || l.ch == 0 {
			// 遇到结束标记，退出循环
			break
		}
	}

	// 使用字符串切片提取字符串内容
	// 从记录的起始位置 position（跳过开头的双引号）到当前的位置 l.position
	// 返回字符串内容的完整字符串表示（不包含两端的双引号）
	return l.input[position:l.position]
}

// isLetter 函数用于判断一个字符是否为字母或下划线
// 该函数是词法分析器的辅助函数，用于标识符的字符识别
// 参数 ch 是要检查的字符
// 返回值：如果是字母（大小写）或下划线则返回 true，否则返回 false
func isLetter(ch byte) bool {
	// 检查字符是否为小写字母：'a' 到 'z'
	// 或者检查字符是否为大写字母：'A' 到 'Z'
	// 或者检查字符是否为下划线：'_'
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit 函数用于判断一个字符是否为数字字符
// 该函数是词法分析器的辅助函数，用于数字字面量的字符识别
// 参数 ch 是要检查的字符
// 返回值：如果是数字字符（0-9）则返回 true，否则返回 false
func isDigit(ch byte) bool {
	// 检查字符是否在数字字符范围内：'0' 到 '9'
	// 使用字符比较判断字符是否在 ASCII 数字字符范围内
	return '0' <= ch && ch <= '9'
}

// newToken 函数是词法分析器的辅助函数，用于创建单字符 Token
// 该函数封装了 Token 的创建逻辑，简化了单字符运算符和分隔符的 Token 生成
// 参数 tokenType 是 Token 的类型，ch 是字符字面量
// 返回值是一个新创建的 token.Token 结构体实例
func newToken(tokenType token.TokenType, ch byte) token.Token {
	// 创建并返回一个新的 Token 实例
	// Type 字段设置为传入的 tokenType
	// Literal 字段通过 string(ch) 将字符转换为字符串
	return token.Token{Type: tokenType, Literal: string(ch)}
}
