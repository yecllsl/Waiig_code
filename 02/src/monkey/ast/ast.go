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

// ReturnStatement 结构体表示一个返回语句
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

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
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

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
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

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
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

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
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
