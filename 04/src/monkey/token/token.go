package token

// TokenType 定义了 Monkey 编程语言中所有可能的词法单元类型
type TokenType string

// 定义所有 Token 类型的常量
const (
	// 特殊 Token 类型
	ILLEGAL = "ILLEGAL" // 非法字符或无法识别的 Token
	EOF     = "EOF"     // 文件结束标记 (End of File)

	// 标识符和字面量
	IDENT  = "IDENT"  // 标识符：变量名、函数名等（如：add, foobar, x, y, ...）
	INT    = "INT"    // 整数字面量（如：1343456）
	STRING = "STRING" // 字符串字面量（如："foobar"）

	// 运算符
	ASSIGN   = "=" // 赋值运算符
	PLUS     = "+" // 加法运算符
	MINUS    = "-" // 减法运算符
	BANG     = "!" // 逻辑非运算符
	ASTERISK = "*" // 乘法运算符
	SLASH    = "/" // 除法运算符

	// 比较运算符
	LT = "<" // 小于运算符
	GT = ">" // 大于运算符

	// 相等性运算符
	EQ     = "==" // 等于运算符
	NOT_EQ = "!=" // 不等于运算符

	// 分隔符
	COMMA     = "," // 逗号分隔符
	SEMICOLON = ";" // 分号分隔符
	COLON     = ":" // 冒号分隔符

	// 括号和分组符号
	LPAREN   = "(" // 左圆括号
	RPAREN   = ")" // 右圆括号
	LBRACE   = "{" // 左花括号
	RBRACE   = "}" // 右花括号
	LBRACKET = "[" // 左方括号
	RBRACKET = "]" // 右方括号

	// 关键字
	FUNCTION = "FUNCTION" // 函数定义关键字
	LET      = "LET"      // 变量声明关键字
	TRUE     = "TRUE"     // 布尔真值关键字
	FALSE    = "FALSE"    // 布尔假值关键字
	IF       = "IF"       // 条件语句关键字
	ELSE     = "ELSE"     // 条件语句关键字
	RETURN   = "RETURN"   // 返回值关键字
)

// Token 结构体表示 Monkey 编程语言中的一个词法单元
// 每个 Token 包含类型信息和字面值，用于词法分析和语法分析
type Token struct {
	// Type 表示 Token 的类型，使用预定义的 TokenType 常量
	// 例如：IDENT（标识符）、INT（整数）、LET（关键字）等
	Type TokenType

	// Literal 存储 Token 的实际字符串值
	// 对于标识符：存储变量名（如 "x"、"add"）
	// 对于字面量：存储具体的值（如 "123"、"hello"）
	// 对于运算符：存储运算符字符（如 "+"、"=="）
	// 对于关键字：存储关键字字符串（如 "let"、"if"）
	Literal string
}

// keywords 是一个映射表，用于将 Monkey 语言的关键字字符串映射到对应的 Token 类型
// 这个映射表在词法分析阶段用于区分关键字和普通标识符
var keywords = map[string]TokenType{
	"fn":     FUNCTION, // 函数定义关键字 -> FUNCTION Token 类型
	"let":    LET,      // 变量声明关键字 -> LET Token 类型
	"true":   TRUE,     // 布尔真值关键字 -> TRUE Token 类型
	"false":  FALSE,    // 布尔假值关键字 -> FALSE Token 类型
	"if":     IF,       // 条件语句关键字 -> IF Token 类型
	"else":   ELSE,     // 条件语句关键字 -> ELSE Token 类型
	"return": RETURN,   // 返回值关键字 -> RETURN Token 类型
}

// LookupIdent 函数用于查找标识符对应的 Token 类型
// 它检查给定的标识符是否是关键字，如果是则返回对应的关键字 Token 类型
// 如果不是关键字，则返回 IDENT 类型（普通标识符）
func LookupIdent(ident string) TokenType {
	// 检查标识符是否存在于关键字映射中
	if tok, ok := keywords[ident]; ok {
		// 如果是关键字，返回对应的 Token 类型
		return tok
	}
	// 如果不是关键字，返回 IDENT 类型（普通标识符）
	return IDENT
}
