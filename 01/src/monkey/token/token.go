package token

type TokenType string //TokenType类型定义成了字符串，这样我们就可以使用各种TokenType值，而根据TokenType值能区分不同类型的词法单元。

const (
	ILLEGAL = "ILLEGAL" //未知词法单元或字符
	EOF     = "EOF"     //表示文件结尾

	// Identifiers + literals 标识符和字面量
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"   // 1343456

	// Operators 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters 分隔符
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Keywords guanjian 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// LookupIdent 检查标识符是否为保留关键字。
// 如果标识符是保留关键字，返回相应的TokenType。
// 否则，返回IDENT类型，表示标识符。
func LookupIdent(ident string) TokenType {
	// 检查keywords字典中是否存在该标识符。
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	// 如果不是关键字，返回IDENT类型。
	return IDENT
}
