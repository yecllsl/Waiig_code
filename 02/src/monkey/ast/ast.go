package ast

import (
	"bytes"
	"monkey/token"
	"strings"
)

// The base Node interface
// Node接口定义了文本节点的通用方法
// 它被设计用于各种类型的节点，以提供统一的操作方式
type Node interface {
	// TokenLiteral方法返回节点的字面字符串值
	// 这通常用于获取节点的原始文本表示，未经过任何解释或处理
	TokenLiteral() string

	// String方法返回节点的字符串表示
	// 与TokenLiteral不同，这个方法可能返回经过处理或格式化的文本
	String() string
}

// All statement nodes implement this
// Statement 接口定义了一个语句节点的标准结构。
// 它继承了 Node 接口，意味着所有语句节点都是节点树的一部分。
// 该接口主要用于标识和统一处理各类语句节点。
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
// Expression 是一个接口，用于定义表达式节点应具备的行为。
// 它继承了 Node 接口，意味着 Expression 接口的实现者也必须是 Node 接口的实现者。
// Expression 接口的存在允许表达式被统一处理和操作，例如在解析、优化或执行表达式时。
type Expression interface {
	Node
	expressionNode()
}

// 定义一个Program结构体，包含一个Statement类型的切片
/* 这个Program节点将成为语法分析器生成的每个AST的根节点。每个有效的Monkey程序都是
一系列位于Program.Statements中的语句。Program.Statements是一个切片，其中有实现
Statement接口的AST节点。 */
type Program struct {
	Statements []Statement
}

// TokenLiteral 返回程序中第一个语句的字面量字符串。
// 如果程序中没有语句，则返回空字符串。
// 该方法主要用于调试和日志记录，提供程序开始部分的快速查看。
func (p *Program) TokenLiteral() string {
	// 检查程序是否包含至少一个语句
	if len(p.Statements) > 0 {
		// 如果有语句，返回第一个语句的字面量字符串
		return p.Statements[0].TokenLiteral()
	} else {
		// 如果没有语句，返回空字符串
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// LetStatement represents a 'let' statement in the program.
// It contains three parts: Token, Name, and Value.
// The purpose of this structure is to define the syntax and structure of a 'let' statement, which is used to declare variables in the program.
type LetStatement struct {
	Token token.Token // token.LET 词法单元，标识声明的类型为 'let'
	Name  *Identifier // Variable name, represented by an Identifier structure
	Value Expression  // Variable value, represented by an Expression interface
}

// statementNode 方法满足 Node 接口的要求。
// 该方法在 LetStatement 类型中没有具体的操作，因为它是一个空方法。
// 主要用于将 LetStatement 类型标识为一个语句节点。
func (ls *LetStatement) statementNode() {}

// TokenLiteral 返回 LetStatement 结构体中 Token 字段的 Literal 属性值。
// 该方法用于获取与 Token 相关联的字面字符串值。
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// LetStatement 的 String 方法返回 LetStatement 的字符串表示形式。
// 该方法主要用于调试和日志记录目的，通过拼接 TokenLiteral、变量名和变量值的字符串表示，
// 以一种人类可读的格式输出 LetStatement 的内容。
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	// 写入 LetStatement 的 TokenLiteral，后面跟一个空格，用于分隔。
	out.WriteString(ls.TokenLiteral() + " ")
	// 写入变量名的字符串表示。
	out.WriteString(ls.Name.String())
	// 写入等号和一个空格，表示赋值操作。
	out.WriteString(" = ")

	// 如果变量值不为空，则写入变量值的字符串表示。
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	// 写入分号，表示语句结束。
	out.WriteString(";")

	// 返回拼接后的字符串表示。
	return out.String()
}

// ReturnStatement 结构体表示一个"返回语句"
type ReturnStatement struct {
	// Token 保存返回语句的 token
	Token token.Token
	// ReturnValue 保存返回语句的返回值
	ReturnValue Expression
}

// statementNode 方法是一个接口实现方法，用于将 ReturnStatement 类型的对象标识为一个语句节点。
// 这个方法没有参数，也没有返回值。它的主要作用是满足某个接口的要求，以便 ReturnStatement 类型的对象
// 可以被视作语句节点处理。在这个特定的实现中，方法体为空，因为它主要用于类型标识而非功能执行。
func (rs *ReturnStatement) statementNode() {}

// TokenLiteral 返回返回语句的字面量字符串
// 该方法主要用于获取与 ReturnStatement 相关的令牌的字面量值
// 没有输入参数
// 返回值: string 类型，表示令牌的字面量字符串
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement 结构体表示一个表达式语句
type ExpressionStatement struct {
	Token      token.Token // 该表达式中的第一个词法单元
	Expression Expression
}

// statementNode 实现了 ExpressionStatement 的 statementNode 方法，
// 该方法用于使 ExpressionStatement 符合 statement 节点接口。
func (es *ExpressionStatement) statementNode() {}

// TokenLiteral 返回表达式语句的字面量字符串。
// 该方法主要用于获取与表达式语句关联的令牌的字面量值。
// 没有输入参数。
// 返回值是字符串类型，表示令牌的字面量值。
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// String 方法返回表达式语句的字符串表示。
// 如果 Expression 属性不为 nil，则调用该表达式的 String 方法并返回其结果。
// 如果 Expression 属性为 nil，则返回空字符串。
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Expressions
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// Boolean 结构体，用于表示布尔类型的值
type Boolean struct {
	// Token 字段，用于存储布尔值对应的 Token
	Token token.Token
	// Value 字段，用于存储布尔值
	Value bool
}

// expressionNode 方法定义了 Boolean 类型的节点作为表达式节点的接口。
// 该方法没有参数和返回值，当前实现为空，可能预留用于未来扩展或特定场景下的操作定义。
func (b *Boolean) expressionNode() {}

// TokenLiteral 返回Boolean类型的字面量字符串。
// 该方法主要用于获取原始的字符串表示，通常用于解析或打印原始输入值。
// 没有输入参数。
// 返回值为字符串类型，表示Boolean类型的字面量。
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// String 实现了 fmt.Stringer 接口，用于将 Boolean 类型的值转换为字符串表示形式。
// 这个方法直接返回原始字面值，即 Token.Literal。
func (b *Boolean) String() string {
	return b.Token.Literal
}

// IntegerLiteral 表示源代码中的整数字面量。
// 它包括字面量的标记信息及其实际的整数值。
type IntegerLiteral struct {
	Token token.Token // Token 表示整数字面量的标记信息，如类型和在源代码中的位置。
	Value int64       // Value 表示整数字面量的实际整数值。
}

// expressionNode 方法是一个接口实现方法，用于将 IntegerLiteral 类型的实例标记为表达式节点。
// 这个方法没有参数，也没有返回值。它的主要作用是满足某个接口的要求，使得 IntegerLiteral 类型的实例可以被视为表达式节点。
// 该方法当前为空，是因为在设计上不需要执行任何操作，仅仅是为了实现接口的编译时要求。
func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral 返回IntegerLiteral类型的字面量字符串表示。
// 该方法主要用于获取存储在Token结构中的Literal字段值，
// 即原始的、未经过解释或编译处理的字符串形式的整数。
// 这对于解析或错误处理阶段需要直接访问原始输入字符串的情况非常有用。
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// String 实现了 fmt.Stringer 接口，用于将整数字面量转换为字符串表示形式。
// 这个方法直接返回存储在 Token 中的字面量字符串，不进行额外的格式化处理。
// 主要用途包括调试和日志记录，其中需要以字符串形式显示整数字面量的原始表示。
func (il *IntegerLiteral) String() string { return il.Token.Literal }

// PrefixExpression 表示一个前缀表达式结构。
// 它包含一个标记、一个运算符以及运算符右边的表达式。
type PrefixExpression struct {
	Token    token.Token // 前缀标记，例如 !
	Operator string      // 前缀表达式的运算符，如 +, -, ! 等
	Right    Expression  // 运算符右边的表达式
}

// expressionNode 是 PrefixExpression 类型实现的一个接口方法。
// 该方法用于标识 PrefixExpression 是一种表达式节点。
// 它不执行任何操作，主要用于满足接口要求或类型断言。
func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral 返回前缀表达式的令牌字面值。
// 该方法用于获取与表达式相关联的令牌的字面值字符串。
// 没有输入参数。
// 返回值是字符串类型，表示令牌的字面值。
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// PrefixExpression 的 String 方法用于生成前缀表达式的字符串表示。
// 该方法对于显示表达式树或调试非常有用。
// 返回值是一个字符串，表示整个前缀表达式。
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	// 写入左括号，开始构建前缀表达式的字符串形式。
	out.WriteString("(")
	// 写入操作符，这是前缀表达式的关键部分。
	out.WriteString(pe.Operator)
	// 递归写入右侧表达式的字符串表示，完成表达式树的遍历。
	out.WriteString(pe.Right.String())
	// 写入右括号，标志着一个完整前缀表达式的结束。
	out.WriteString(")")

	// 返回构建好的前缀表达式字符串。
	return out.String()
}

// InfixExpression 表示中缀表示法中的二元表达式，例如 "1 + 2"。
// 该结构体包含操作符的标记（如 +, -, *, /），以及操作符的左操作数和右操作数。
type InfixExpression struct {
	Token    token.Token // 操作符标记，例如 +
	Left     Expression  // 表达式的左操作数
	Operator string      // 操作符，例如 "+", "-", "*", "/"
	Right    Expression  // 表达式的右操作数
}

func (ie *InfixExpression) expressionNode() {}

// TokenLiteral 返回InfixExpression类型的对象的字面量字符串表示。
// 该方法主要用于获取表达式开头的Token的字面量值，通常用于打印或显示目的。
// 参数: 无
// 返回值: string 类型，表示表达式开头Token的字面量值。
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// InfixExpression 的 String 方法返回该结构体的字符串表示形式。
// 该方法实现了 fmt.Stringer 接口，用于自定义对象的字符串表示。
// 主要用于调试和日志记录目的，提供了一种标准的方式来显示 InfixExpression 对象的内容。
func (ie *InfixExpression) String() string {
	// 创建一个可写缓冲区来构建最终的字符串表示。
	var out bytes.Buffer

	// 写入左括号，标志着表达式的开始。
	out.WriteString("(")
	// 递归调用 Left 节点的 String 方法，构建左子表达式的字符串表示。
	out.WriteString(ie.Left.String())
	// 写入操作符，并在其前后添加空格，以符合常规的表达式格式。
	out.WriteString(" " + ie.Operator + " ")
	// 递归调用 Right 节点的 String 方法，构建右子表达式的字符串表示。
	out.WriteString(ie.Right.String())
	// 写入右括号，标志着表达式的结束。
	out.WriteString(")")

	// 返回构建好的表达式字符串。
	return out.String()
}

// IfExpression 表示抽象语法树中的 'if' 表达式结构。
// 它包括 'if' 标记、条件表达式、结果代码块和备选代码块。
type IfExpression struct {
	Token       token.Token     // 'if' 标记
	Condition   Expression      // 'if' 语句的条件表达式
	Consequence *BlockStatement // 条件为真时执行的代码块
	Alternative *BlockStatement // 条件为假时执行的代码块
}

// expressionNode 方法是 IfExpression 类型的一个空白方法，没有参数和返回值。
// 它的存在可能是为了满足某个接口的要求，或者是 IfExpression 类型的一个占位方法。
// 该方法没有实现任何逻辑，因此不执行任何操作。
func (ie *IfExpression) expressionNode() {}

// TokenLiteral 返回 IfExpression 结构体中 Token 的字面值字符串
// 该方法主要用于获取 If 表达式开头的 if 关键字的字面值
// 它没有输入参数，返回值为字符串类型
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// IfExpression 的 String 方法返回该如果表达式的字符串表示。
// 该方法通过拼接 "if"、条件表达式、后续执行的表达式以及可选的 "else" 和替代执行的表达式来构建整个字符串。
// 这主要用于调试和日志记录目的，以便开发者可以以人类可读的形式查看表达式的结构。
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	// 写入 "if" 开头，紧接着是条件表达式的字符串表示。
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	// 如果存在替代执行的表达式（else 分支），则也写入其字符串表示。
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	// 返回构建好的字符串。
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
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

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
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
