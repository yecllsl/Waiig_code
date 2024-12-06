package lexer

import "monkey/token"

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)所输入字符串的当前位置（指向当前字符）
	readPosition int  // current reading position in input (after current char)所输入字符串中的当前读取位置（在当前字符之后的一个字符）
	ch           byte // current char under examination当前正在查看的字符
}

/* 在New()函数中使用readChar，初始化l.ch、l.position和l.readPosition，
以便在调用NextToken()之前让*Lexer完全就绪： */
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

/* 这就是NextToken()方法的基本结构。它首先检查了当前正在查看的字符l.ch，
根据具体的字符来返回对应的词法单元。在返回词法单元之前，位于所输入字符串中
的指针会前移，所以之后再次调用NextToken()时，l.ch字段就已经更新过了。
最后，名为newToken的小型函数可以帮助初始化这些词法单元。 */
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// 跳过输入中的空白字符
	l.skipWhitespace()

	// 根据当前字符确定词法标记的类型和字面值
	switch l.ch {
	case '=':
		// 检查是否为等号（==）或赋值号（=）
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		// 加号
		tok = newToken(token.PLUS, l.ch)
	case '-':
		// 减号
		tok = newToken(token.MINUS, l.ch)
	case '!':
		// 检查是否为不等号（!=）或感叹号（!）
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		// 除号
		tok = newToken(token.SLASH, l.ch)
	case '*':
		// 乘号
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		// 小于号
		tok = newToken(token.LT, l.ch)
	case '>':
		// 大于号
		tok = newToken(token.GT, l.ch)
	case ';':
		// 分号
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		// 逗号
		tok = newToken(token.COMMA, l.ch)
	case '{':
		// 左大括号
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		// 右大括号
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		// 左圆括号
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		// 右圆括号
		tok = newToken(token.RPAREN, l.ch)
	case 0:
		// 文件结束
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		// 处理字母、数字或非法字符
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else { //将不知道如何处理的字符声明成类型为token.ILLEGAL的词法单元。
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	// 读取下一个字符
	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readChar() {
	/* readChar的目的是读取input中的下一个字符，并前移其在input中的位置。
	这个过程的第一件事就是检查是否已经到达input的末尾。如果是，则将l.ch设置为0，
	这是NUL字符的ASCII编码，用来表示“尚未读取任何内容”或“文件结尾”​。
	如果还没有到达input的末尾，则将l.ch设置为下一个字符，即l.input[l.readPosition]指向的字符。 */
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	/* 之后，将l.position更新为刚用过的l.readPosition，然后将l.readPosition加1。
		这样一来，l.readPosition就始终指向下一个将读取的字符位置，而l.position始终指向
		// New 函数用于创建并初始化一个新的 Lexer 实例
	func New(input string) *Lexer {
		// 创建一个新的 Lexer 实例 l，并将输入字符串 input 赋值给 l 的 input 字段
		l := &Lexer{input: input}
		// 调用 l 的 readChar 方法，读取输入字符串的第一个字符，并将其赋值给 l 的 ch 字段
		l.readChar()
		// 返回新创建的 Lexer 实例 l
		return l
	}
	// New 函数用于创建并初始化一个新的 Lexer 实例
	func New(input string) *Lexer {
		// 创建一个新的 Lexer 实例 l，并将输入字符串 input 赋值给 l 的 input 字段
		l := &Lexer{input: input}
		// 调用 l 的 readChar 方法，读取输入字符串的第一个字符，并将其赋值给 l 的 ch 字段
		l.readChar()
		// 返回新创建的 Lexer 实例 l
		return l
	}
	// New 函数用于创建并初始化一个新的 Lexer 实例
	func New(input string) *Lexer {
		// 创建一个新的 Lexer 实例 l，并将输入字符串 input 赋值给 l 的 input 字段
		l := &Lexer{input: input}
		// 调用 l 的 readChar 方法，读取输入字符串的第一个字符，并将其赋值给 l 的 ch 字段
		l.readChar()
		// 返回新创建的 Lexer 实例 l
		return l
	}
	// New 函数用于创建并初始化一个新的 Lexer 实例
	func New(input string) *Lexer {
		// 创建一个新的 Lexer 实例 l，并将输入字符串 input 赋值给 l 的 input 字段
		l := &Lexer{input: input}
		// 调用 l 的 readChar 方法，读取输入字符串的第一个字符，并将其赋值给 l 的 ch 字段
		l.readChar()
		// 返回新创建的 Lexer 实例 l
		return l
	}
	刚刚读取的位置。这个特性很快就会派上用场。 */
	l.position = l.readPosition
	l.readPosition += 1
}

// peekChar 返回当前位置之后的一个字符，而不改变读取位置。
// 这个函数用于在不移动读取位置的情况下预览下一个字符。
// 如果读取位置已经到达或超过输入字符串的末尾，则返回0。
// 否则，返回读取位置之后的字符。
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// isLetter 判断给定的字符是否为字母或下划线。
// 参数:
//   ch byte: 待检查的字符。
// 返回值:
//   bool: 如果字符是字母或下划线，则返回true；否则返回false。
// 该函数通过比较字符的ASCII值来判断其是否在字母的范围内或为下划线。
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// newToken 创建并返回一个新的token实例。
// 该函数接收一个token类型(tokenType)和一个字节(ch)，并使用这些参数构建一个token对象。
// 参数tokenType指定新token的类型，参数ch被转换为字符串后作为token的字面值。
// 返回值是创建的token对象实例。
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
