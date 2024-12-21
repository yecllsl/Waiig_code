package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

// ObjectType 定义了一个对象类型的别名。
// 使用类型别名可以提高代码的可读性和自文档化。
type ObjectType string

// 定义对象类型常量
const (
	NULL_OBJ  = "NULL"  // 表示空对象
	ERROR_OBJ = "ERROR" // 表示错误对象

	INTEGER_OBJ = "INTEGER" // 表示整数对象
	BOOLEAN_OBJ = "BOOLEAN" // 表示布尔对象

	RETURN_VALUE_OBJ = "RETURN_VALUE" // 表示返回值对象

	FUNCTION_OBJ = "FUNCTION" // 表示函数对象
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

// Integer 类型的方法 Type 返回整数对象的类型
// 该方法满足 ObjectType 接口，用于标识对象类型
// 参数: 无
// 返回值: ObjectType 类型，表示整数对象的类型
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// Inspect 返回当前整数对象的字符串表示。
// 该方法主要用于调试和日志记录目的，提供了一个标准的方式来观察整数对象的值。
// 返回值是一个字符串，格式为整数的值。
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

type Boolean struct {
	Value bool
}

// Boolean.Type 返回布尔类型的 ObjectType。
// 该方法实现了 ObjectType 接口，用于标识对象的类型。
// 参数: 无
// 返回值: ObjectType，表示布尔类型的类型标识。
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

// Type 返回Null对象的类型
// 该方法实现了ObjectType接口，用于标识对象类型
// 参数: 无
// 返回值: ObjectType类型，表示NULL_OBJ
func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue 是一个结构体，用于封装一个对象值。
// 它的主要作用是作为某些函数或方法的返回值，以便于统一和简化返回值的处理。
type ReturnValue struct {
	Value Object // Value 字段用于存储函数或方法希望返回的对象值。
}

// Type 返回ReturnValue对象的类型。
// 该方法实现了ObjectType接口，用于标识对象的类型。
// 参数: 无
// 返回值: ObjectType，表示ReturnValue对象的类型。
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Inspect 返回ReturnValue内部值的字符串表示。
// 该方法主要用于获取ReturnValue实例所封装值的详细信息。
func (rv *ReturnValue) Inspect() string {
	// 调用Value的Inspect方法，获取字符串表示的值。
	return rv.Value.Inspect()
}

// Error 定义了一个包含错误信息的简单结构体。
// 该结构体用于封装错误信息，提供更详细的错误描述。
type Error struct {
	// Message 存储错误消息，描述发生了什么问题。
	Message string
}

// Error 类型的 Type 方法返回错误对象的类型
// 该方法用于标识 Error 实例属于 ERROR_OBJ 类型
// 主要用于对象类型识别和类型转换
func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Inspect 返回错误的详细信息。
// 该方法将错误消息前加上"ERROR: "前缀，以提供更清晰的错误指示。
func (e *Error) Inspect() string { return "ERROR: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
