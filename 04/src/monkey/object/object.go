package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"monkey/ast"
	"strings"
)

// BuiltinFunction 定义内置函数的函数签名类型
// 用于表示 Monkey 语言中的内置函数，这些函数由解释器直接实现而非用户定义
// 函数签名：接受可变数量的 Object 参数，返回一个 Object 结果
type BuiltinFunction func(args ...Object) Object

type ObjectType string

// 定义 Monkey 语言中所有对象类型的常量标识符
// 这些常量用于标识和区分不同类型的对象，在类型检查和运行时类型判断中使用
const (
	NULL_OBJ  = "NULL"  // 空值对象类型标识符
	ERROR_OBJ = "ERROR" // 错误对象类型标识符

	INTEGER_OBJ = "INTEGER" // 整数对象类型标识符
	BOOLEAN_OBJ = "BOOLEAN" // 布尔值对象类型标识符
	STRING_OBJ  = "STRING"  // 字符串对象类型标识符

	RETURN_VALUE_OBJ = "RETURN_VALUE" // 返回值包装对象类型标识符

	FUNCTION_OBJ = "FUNCTION" // 用户定义函数对象类型标识符
	BUILTIN_OBJ  = "BUILTIN"  // 内置函数对象类型标识符

	ARRAY_OBJ = "ARRAY" // 数组对象类型标识符
	HASH_OBJ  = "HASH"  // 哈希表对象类型标识符
)

// HashKey 结构体用于表示哈希表的键
// 在 Monkey 语言的哈希表实现中，用于唯一标识和快速查找键值对
type HashKey struct {
	Type  ObjectType // 对象类型标识符，确保类型安全
	Value uint64     // 哈希值，基于对象内容计算得出
}

// Hashable 接口定义了可哈希对象的行为规范
// 实现此接口的对象可以作为哈希表的键使用，支持哈希表的键值对存储和快速查找
type Hashable interface {
	HashKey() HashKey // 返回对象的哈希键，包含类型标识符和哈希值
}

// Object 接口是 Monkey 语言对象系统的核心接口
// 所有 Monkey 语言中的值类型都必须实现此接口，提供统一的类型检查和值表示机制
type Object interface {
	Type() ObjectType // 返回对象的类型标识符，用于运行时类型检查
	Inspect() string  // 返回对象的可读字符串表示，用于调试和 REPL 环境显示
}

// Integer 结构体表示 Monkey 语言中的整数对象
// 用于存储和操作整数值，支持算术运算和哈希表键功能
type Integer struct {
	Value int64 // 存储整数值，使用 int64 类型支持大整数运算
}

// Type 方法实现 Object 接口，返回整数对象的类型标识符
// 用于运行时类型检查和类型安全，确保整数对象被正确识别和处理
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// Inspect 方法实现 Object 接口，返回整数对象的可读字符串表示
// 用于调试输出、REPL 环境显示和错误消息，提供人类可读的整数值表示
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// HashKey 方法实现 Hashable 接口，返回整数对象的哈希键
// 用于哈希表键值对存储和快速查找，确保整数对象可以作为哈希表的键使用
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// Boolean 结构体表示 Monkey 语言中的布尔值对象
// 用于存储和操作布尔值，支持逻辑运算和哈希表键功能
type Boolean struct {
	Value bool // 存储布尔值，支持 true 和 false 两种状态
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// Inspect 方法实现 Object 接口，返回布尔值对象的可读字符串表示
// 用于调试输出、REPL 环境显示和错误消息，提供人类可读的布尔值表示
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// HashKey 方法实现 Hashable 接口，返回布尔值对象的哈希键
// 用于哈希表键值对存储和快速查找，确保布尔值对象可以作为哈希表的键使用
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

// Null 结构体表示 Monkey 语言中的空值对象
// 用于表示空值或缺失值，是 Monkey 语言中的特殊值类型
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue 结构体表示 Monkey 语言中的返回值包装对象
// 用于包装函数返回值，支持多层嵌套返回和返回值传递机制
type ReturnValue struct {
	Value Object // 存储实际返回的对象值，可以是任何实现 Object 接口的类型
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error 结构体表示 Monkey 语言中的错误对象
// 用于表示运行时错误和异常情况，支持错误信息的存储和传递
type Error struct {
	Message string // 存储错误消息，描述具体的错误原因和上下文信息
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Function 结构体表示 Monkey 语言中的用户定义函数对象
// 用于存储和表示用户定义的函数，支持函数定义、参数列表和函数体执行
type Function struct {
	Parameters []*ast.Identifier   // 函数参数列表，存储参数标识符的指针数组
	Body       *ast.BlockStatement // 函数体，存储包含语句块的抽象语法树节点
	Env        *Environment        // 函数执行环境，存储变量作用域和闭包信息
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect 方法实现 Object 接口，返回函数对象的可读字符串表示
// 用于调试输出、REPL 环境显示和错误消息，提供人类可读的函数定义表示
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

// String 结构体表示 Monkey 语言中的字符串对象
// 用于存储和操作字符串值，支持字符串操作和哈希表键功能
type String struct {
	Value string // 存储字符串值，支持 Unicode 字符和任意长度的文本数据
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// HashKey 方法实现 Hashable 接口，返回字符串对象的哈希键
// 用于哈希表键值对存储和快速查找，确保字符串对象可以作为哈希表的键使用
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// Builtin 结构体表示 Monkey 语言中的内置函数对象
// 用于封装和表示语言内置的函数功能，提供预定义的函数实现和高效执行
type Builtin struct {
	Fn BuiltinFunction // 内置函数实现，存储实际的内置函数逻辑和功能
}

// Type 方法实现 Object 接口，返回内置函数对象的类型标识符
// 用于运行时类型检查和类型安全，确保内置函数对象被正确识别和处理
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

// Inspect 方法实现 Object 接口，返回内置函数对象的可读字符串表示
// 用于调试输出、REPL 环境显示和错误消息，提供统一的内置函数标识表示
func (b *Builtin) Inspect() string { return "builtin function" }

// Array 结构体表示 Monkey 语言中的数组对象
// 用于存储和操作对象数组，支持数组元素的存储、访问和遍历操作
type Array struct {
	Elements []Object // 存储数组元素，支持任意类型的对象元素集合
}

// Type 方法实现 Object 接口，返回数组对象的类型标识符
// 用于运行时类型检查和类型安全，确保数组对象被正确识别和处理
func (ao *Array) Type() ObjectType { return ARRAY_OBJ }

// Inspect 方法实现 Object 接口，返回数组对象的可读字符串表示
// 用于调试输出、REPL 环境显示和错误消息，提供人类可读的数组内容表示
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// HashPair 结构体表示 Monkey 语言中哈希表的键值对
// 用于存储哈希表中的键值对关系，支持键值对的存储、访问和遍历操作
type HashPair struct {
	Key   Object // 存储哈希表的键，必须是实现 Hashable 接口的对象
	Value Object // 存储哈希表的值，可以是任意实现 Object 接口的对象
}

// Hash 结构体表示 Monkey 语言中的哈希表对象
// 用于存储和操作键值对集合，支持高效的键值对存储、查找和遍历操作
type Hash struct {
	Pairs map[HashKey]HashPair // 存储哈希表的键值对映射，使用哈希键作为映射键
}

// Type 方法实现 Object 接口，返回哈希表对象的类型标识符
// 用于运行时类型检查和类型安全，确保哈希表对象被正确识别和处理
func (h *Hash) Type() ObjectType { return HASH_OBJ }

// Inspect 方法实现 Object 接口，返回哈希表对象的可读字符串表示
// 用于调试输出、REPL 环境显示和错误消息，提供人类可读的哈希表内容表示
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
