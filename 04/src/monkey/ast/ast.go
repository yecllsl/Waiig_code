package ast

import (
	"bytes"
	"monkey/token"
	"strings"
)

// The base Node interface
// Node 接口是 Monkey 编程语言抽象语法树（AST）的基础接口
// 所有 AST 节点类型都必须实现这个接口，为 AST 提供统一的访问方式
type Node interface {
	// TokenLiteral 方法返回节点对应的词法单元的字面量
	// 该方法提供对节点原始源代码文本的访问，用于调试和错误报告
	TokenLiteral() string

	// String 方法返回节点的字符串表示形式
	// 该方法用于将 AST 节点转换回可读的源代码格式，便于调试和测试
	String() string
}

// All statement nodes implement this
// Statement 接口是所有语句节点的基接口
// 语句是编程语言中执行操作的语法结构，不产生值但可能产生副作用
type Statement interface {
	// 继承 Node 接口，获得 TokenLiteral() 和 String() 方法
	Node

	// statementNode 方法是语句节点的标记方法
	// 该方法没有实际逻辑，仅用于类型断言和接口区分
	statementNode()
}

// All expression nodes implement this
// Expression 接口是所有表达式节点的基接口
// 表达式是编程语言中产生值的语法结构，可以参与计算和赋值操作
type Expression interface {
	// 继承 Node 接口，获得 TokenLiteral() 和 String() 方法
	Node

	// expressionNode 方法是表达式节点的标记方法
	// 该方法没有实际逻辑，仅用于类型断言和接口区分
	expressionNode()
}

// Program 结构体表示Monkey语言的整个程序
// 作为抽象语法树（AST）的根节点，包含程序中的所有语句
type Program struct {
	Statements []Statement // 语句切片，存储程序中的所有语句节点
}

// TokenLiteral 方法实现Node接口，返回Program的标记字面量
// 如果Program包含语句，返回第一个语句的标记字面量；否则返回空字符串
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral() // 返回第一个语句的标记字面量
	} else {
		return "" // 空程序返回空字符串
	}
}

// String 方法实现Node接口，返回Program的字符串表示
// 通过遍历所有语句并调用其String()方法，拼接成完整的程序字符串
func (p *Program) String() string {
	var out bytes.Buffer // 创建字节缓冲区用于字符串拼接

	for _, s := range p.Statements {
		out.WriteString(s.String()) // 将每个语句的字符串表示写入缓冲区
	}

	return out.String() // 返回拼接后的完整程序字符串
}

// Statements
// LetStatement 结构体表示Monkey语言中的变量声明语句
// 语法格式：let <identifier> = <expression>;
type LetStatement struct {
	Token token.Token // the token.LET token - let关键字对应的词法标记
	Name  *Identifier // 变量名标识符，指向Identifier表达式节点
	Value Expression  // 赋值表达式，可以是任意类型的表达式节点
}

// statementNode 方法实现Statement接口，作为LetStatement的标记方法
// 该方法没有实际逻辑，仅用于类型断言和接口区分
func (ls *LetStatement) statementNode() {}

// TokenLiteral 方法实现Node接口，返回LetStatement的标记字面量
// 直接返回Token字段的Literal属性，即"let"关键字的字面量
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// String 方法实现Node接口，返回LetStatement的字符串表示
// 生成格式为"let <identifier> = <expression>;"的完整变量声明语句
func (ls *LetStatement) String() string {
	var out bytes.Buffer // 创建字节缓冲区用于字符串拼接

	out.WriteString(ls.TokenLiteral() + " ") // 写入"let"关键字和空格
	out.WriteString(ls.Name.String())        // 写入变量名标识符
	out.WriteString(" = ")                   // 写入赋值运算符和空格

	if ls.Value != nil {
		out.WriteString(ls.Value.String()) // 写入赋值表达式（如果存在）
	}

	out.WriteString(";") // 写入语句结束分号

	return out.String() // 返回拼接后的完整语句字符串
}

// ReturnStatement 结构体表示 Monkey 语言中的返回语句
// 语法格式为：return <expression>;
// 该语句用于从函数中返回一个值，是函数执行流程控制的一部分
type ReturnStatement struct {
	Token       token.Token // 存储 'return' 关键字的词法标记，用于标识语句类型和位置信息
	ReturnValue Expression  // 存储要返回的表达式，可以是任意类型的表达式节点（如标识符、字面量、函数调用等）
}

// statementNode 方法是 ReturnStatement 结构体实现 Statement 接口的标记方法
// 该方法为空实现，不包含任何实际逻辑，主要用于类型标记和接口实现验证
// 作用：通过实现此方法，ReturnStatement 正式声明自己是一个语句节点，支持类型断言和接口区分
func (rs *ReturnStatement) statementNode() {}

// TokenLiteral 方法实现 Node 接口，返回 ReturnStatement 的标记字面量
// 该方法直接访问 Token 字段的 Literal 属性，返回 "return" 关键字的字面量字符串
// 作用：为 AST 节点提供标记字面量访问接口，支持调试输出、错误报告和代码生成
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// String 方法实现 Node 接口，返回 ReturnStatement 的完整字符串表示
// 该方法通过 bytes.Buffer 拼接 "return" 关键字、返回值表达式和分号，重建源代码格式
// 作用：为调试输出、代码生成和测试验证提供语句的完整字符串表示
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	// 写入 "return" 关键字和空格
	out.WriteString(rs.TokenLiteral() + " ")

	// 检查返回值表达式是否存在，避免空指针异常
	if rs.ReturnValue != nil {
		// 递归调用返回值表达式的 String 方法，获取其字符串表示
		out.WriteString(rs.ReturnValue.String())
	}

	// 写入语句结束的分号
	out.WriteString(";")

	// 返回拼接后的完整语句字符串
	return out.String()
}

// ExpressionStatement 结构体表示 Monkey 语言中的表达式语句
// 语法格式为：<expression>;
// 该语句将表达式包装为独立的语句，允许表达式在语句上下文中使用
type ExpressionStatement struct {
	Token      token.Token // 存储表达式第一个词法标记，用于标识语句类型和位置信息
	Expression Expression  // 存储被包装的表达式节点，可以是任意类型的表达式
}

// statementNode 方法是 ExpressionStatement 结构体实现 Statement 接口的标记方法
// 该方法为空实现，不包含任何实际逻辑，主要用于类型标记和接口实现验证
// 作用：通过实现此方法，ExpressionStatement 正式声明自己是一个语句节点，支持类型断言和接口区分
func (es *ExpressionStatement) statementNode() {}

// TokenLiteral 方法实现 Node 接口，返回 ExpressionStatement 的标记字面量
// 该方法直接访问 Token 字段的 Literal 属性，返回表达式第一个标记的字面量字符串
// 作用：为 AST 节点提供标记字面量访问接口，支持调试输出、错误报告和代码生成
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// String 方法实现 Node 接口，返回 ExpressionStatement 的字符串表示
// 该方法通过递归调用 Expression 字段的 String 方法，返回表达式的完整字符串表示
// 作用：为调试输出、代码生成和测试验证提供表达式语句的字符串表示
func (es *ExpressionStatement) String() string {
	// 检查表达式是否存在，避免空指针异常
	if es.Expression != nil {
		// 递归调用表达式的 String 方法，获取其完整字符串表示
		return es.Expression.String()
	}
	// 如果表达式为 nil，返回空字符串
	return ""
}

// BlockStatement 表示 Monkey 语言中的代码块语句
// 代码块语句由一对花括号 {} 包围，包含零个或多个语句序列
// 语法格式：{ <statement1>; <statement2>; ... }
type BlockStatement struct {
	Token      token.Token // 代码块的起始标记，通常是左花括号 '{'
	Statements []Statement // 代码块中包含的语句序列，可以为空
}

// statementNode 是 BlockStatement 实现 Statement 接口的标记方法
// 该方法为空实现，仅用于类型标记和接口实现验证
// 作用：声明 BlockStatement 是一个语句节点类型，支持类型断言和接口区分
func (bs *BlockStatement) statementNode() {}

// TokenLiteral 返回 BlockStatement 的标记字面量字符串
// 实现 Node 接口，直接返回 Token 字段的 Literal 属性
// 对于代码块语句，通常返回左花括号 '{' 的字面量
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// String 返回 BlockStatement 的字符串表示
// 实现 Node 接口，通过遍历 Statements 切片并递归调用每个语句的 String() 方法
// 将代码块中的所有语句拼接成完整的字符串表示
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	// 遍历代码块中的所有语句
	for _, s := range bs.Statements {
		// 递归调用每个语句的 String() 方法，将结果写入缓冲区
		out.WriteString(s.String())
	}

	return out.String()
}

// Expressions
// Identifier 表示 Monkey 语言中的标识符表达式
// 标识符是变量名、函数名等用户定义的名称
// 语法格式：<variable_name> 或 <function_name>
type Identifier struct {
	Token token.Token // 标识符的词法标记，类型为 token.IDENT
	Value string      // 标识符的实际字符串值，如变量名 "x" 或函数名 "add"
}

// expressionNode 是 Identifier 实现 Expression 接口的标记方法
// 该方法为空实现，仅用于类型标记和接口实现验证
// 作用：声明 Identifier 是一个表达式节点类型，支持类型断言和接口区分
func (i *Identifier) expressionNode() {}

// TokenLiteral 实现 Node 接口，返回标识符的词法标记字面量
// 该方法直接访问 Identifier 的 Token 字段的 Literal 属性
// 返回值：标识符对应的词法标记字面量字符串，通常是变量名或函数名
// 作用：提供标识符的词法标记访问接口，支持调试输出和代码生成
// 设计特点：简单直接、高效访问、类型安全、保持接口一致性
// 使用场景：调试输出、错误报告、代码生成、测试验证、REPL环境
// 设计意义：体现 AST 节点基本特性，支持源代码逆向转换
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// String 实现 Node 接口，返回标识符的字符串表示
// 该方法直接返回 Identifier 的 Value 字段，即标识符的实际名称
// 返回值：标识符的字符串值，如变量名 "x" 或函数名 "add"
// 作用：提供标识符的字符串表示，支持调试输出和代码生成
// 设计特点：简单直接、高效访问、类型安全、保持接口一致性
// 使用场景：调试输出、代码生成、测试验证、REPL环境、AST遍历
// 设计意义：体现标识符的本质特性，支持源代码逆向转换
// 特殊意义：在 Identifier 中，String() 方法返回的是标识符的实际值，而非词法标记
func (i *Identifier) String() string { return i.Value }

// Boolean 表示 Monkey 语言中的布尔值表达式
// 布尔值是逻辑值，只能是 true 或 false
// 语法格式：true 或 false
type Boolean struct {
	Token token.Token // 布尔值的词法标记，类型为 token.TRUE 或 token.FALSE
	Value bool        // 布尔值的实际逻辑值，true 或 false
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// IntegerLiteral 表示 Monkey 语言中的整数字面量表达式
// 整数字面量是表示整数值的表达式
// 语法格式：<integer_value>，如 5、100、-42 等
type IntegerLiteral struct {
	Token token.Token // 整数字面量的词法标记，类型为 token.INT
	Value int64       // 整数字面量的实际数值，存储为 64 位整数
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// PrefixExpression 表示 Monkey 语言中的前缀表达式
// 前缀表达式是操作符位于操作数之前的表达式
// 语法格式：<operator><operand>，如 !true、-5 等
type PrefixExpression struct {
	Token    token.Token // 前缀操作符的词法标记，如 !、- 等
	Operator string      // 前缀操作符的字符串表示，如 "!"、"-"
	Right    Expression  // 右侧的操作数表达式
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// InfixExpression 表示 Monkey 语言中的中缀表达式
// 中缀表达式是操作符位于两个操作数之间的表达式
// 语法格式：<left_operand> <operator> <right_operand>，如 5 + 3、x == y 等
type InfixExpression struct {
	Token    token.Token // 中缀操作符的词法标记，如 +、-、*、/、==、!= 等
	Left     Expression  // 左侧的操作数表达式
	Operator string      // 中缀操作符的字符串表示，如 "+"、"-"、"=="、"!=" 等
	Right    Expression  // 右侧的操作数表达式
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// IfExpression 表示 Monkey 语言中的条件表达式
// 条件表达式根据条件执行不同的代码块
// 语法格式：if <condition> { <consequence> } [else { <alternative> }]
type IfExpression struct {
	Token       token.Token     // 'if' 关键字的词法标记
	Condition   Expression      // 条件表达式，计算结果为布尔值
	Consequence *BlockStatement // 条件为真时执行的代码块
	Alternative *BlockStatement // 条件为假时执行的代码块（可选）
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

// FunctionLiteral 表示 Monkey 语言中的函数字面量表达式
// 函数字面量是定义匿名函数的表达式
// 语法格式：fn(<parameters>) { <body> }
type FunctionLiteral struct {
	Token      token.Token     // 'fn' 关键字的词法标记
	Parameters []*Identifier   // 函数参数列表，每个参数是一个标识符
	Body       *BlockStatement // 函数体，包含函数执行的语句序列
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// CallExpression 表示 Monkey 语言中的函数调用表达式
// 函数调用表达式用于执行函数并传递参数
// 语法格式：<function>(<arguments>)
type CallExpression struct {
	Token     token.Token  // 左括号 '(' 的词法标记
	Function  Expression   // 被调用的函数，可以是标识符或函数字面量
	Arguments []Expression // 传递给函数的参数列表
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// StringLiteral 表示 Monkey 语言中的字符串字面量表达式
// 字符串字面量用于表示文本数据，由双引号包围的字符序列组成
// 语法格式："<string_content>"
type StringLiteral struct {
	Token token.Token // 字符串标记，存储 STRING 类型的词法标记
	Value string      // 字符串的实际内容，不包含双引号
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// ArrayLiteral 表示 Monkey 语言中的数组字面量表达式
// 数组字面量用于表示有序的元素集合，由方括号包围的元素列表组成
// 语法格式：[<element1>, <element2>, ..., <elementN>]
type ArrayLiteral struct {
	Token    token.Token  // 左方括号 '[' 的词法标记
	Elements []Expression // 数组中的元素列表，支持任意表达式类型
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

// String 方法实现 Node 接口，返回数组字面量的字符串表示
// 该方法将数组元素转换为字符串并用方括号和逗号分隔符格式化
// 返回值格式：[element1, element2, ..., elementN]
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// IndexExpression 表示 Monkey 语言中的索引表达式
// 索引表达式用于访问数组或哈希表中的特定元素
// 语法格式：<left_expression>[<index_expression>]
type IndexExpression struct {
	Token token.Token // 左方括号 '[' 的词法标记
	Left  Expression  // 被索引的表达式（数组或哈希表）
	Index Expression  // 索引表达式（整数或键值）
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

// String 方法实现 Node 接口，返回索引表达式的字符串表示
// 该方法将索引表达式格式化为括号包围的完整语法结构
// 返回值格式：(left_expression[index_expression])
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

// HashLiteral 表示 Monkey 语言中的哈希表字面量表达式
// 哈希表字面量用于表示键值对集合，由花括号包围的键值对列表组成
// 语法格式：{<key1>: <value1>, <key2>: <value2>, ..., <keyN>: <valueN>}
type HashLiteral struct {
	Token token.Token               // 左花括号 '{' 的词法标记
	Pairs map[Expression]Expression // 哈希表的键值对映射
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }

// String 方法实现 Node 接口，返回哈希表字面量的字符串表示
// 该方法将哈希表的键值对转换为字符串并用花括号和逗号分隔符格式化
// 返回值格式：{key1: value1, key2: value2, ..., keyN: valueN}
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
