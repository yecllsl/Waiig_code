package lexer

import (
	"testing"

	"monkey/token"
)

// TestNextToken 函数是 Lexer 的主要测试函数
// 它测试词法分析器能否正确地将源代码字符串转换为 Token 序列
func TestNextToken(t *testing.T) {
	// 定义一个包含 Monkey 语言各种语法元素的测试输入字符串
	// 这个字符串包含了变量声明、函数定义、条件语句、运算符、字符串、数组、哈希等
	input := `let five = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
`

	// 定义期望的 Token 序列，包含每个 Token 的类型和字面值
	// 这个测试用例覆盖了 Monkey 语言的所有语法特性
	tests := []struct {
		expectedType    token.TokenType // 期望的 Token 类型
		expectedLiteral string          // 期望的 Token 字面值
	}{
		// 第一个变量声明：let five = 5;
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		// 第二个变量声明：let ten = 10;
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},

		// 函数定义：let add = fn(x, y) { x + y; };
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},

		// 函数调用：let result = add(five, ten);
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},

		// 运算符测试：!-/*5;
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		// 比较运算符测试：5 < 10 > 5;
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},

		// 条件语句测试：if (5 < 10) { return true; } else { return false; }
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},

		// 相等性运算符测试：10 == 10; 10 != 9;
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},

		// 字符串字面量测试
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},

		// 数组测试：[1, 2];
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},

		// 哈希表测试：{"foo": "bar"}
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},

		// 文件结束标记
		{token.EOF, ""},
	}

	// 创建新的 Lexer 实例，传入测试输入
	l := New(input)

	// 遍历所有期望的 Token，逐个验证
	for i, tt := range tests {
		// 获取下一个 Token
		tok := l.NextToken()

		// 验证 Token 类型是否正确
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		// 验证 Token 字面值是否正确
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
