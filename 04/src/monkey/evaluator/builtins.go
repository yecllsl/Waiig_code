// Package evaluator 实现 Monkey 语言的求值器功能
// 包含内置函数的定义和实现，为 Monkey 语言提供标准库功能
package evaluator

import (
	"fmt"
	"monkey/object"
)

// builtins 映射定义了 Monkey 语言的所有内置函数
// 每个内置函数都是一个 *object.Builtin 对象，包含实际的函数实现
var builtins = map[string]*object.Builtin{
	// len 内置函数：返回数组或字符串的长度
	// 支持数组和字符串类型，返回整数类型的长度值
	"len": &object.Builtin{Fn: func(args ...object.Object) object.Object {
		// 参数数量检查：len 函数只接受一个参数
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1",
				len(args))
		}

		// 根据参数类型进行不同的处理
		switch arg := args[0].(type) {
		case *object.Array:
			// 处理数组：返回数组元素的个数
			return &object.Integer{Value: int64(len(arg.Elements))}
		case *object.String:
			// 处理字符串：返回字符串的字符数
			return &object.Integer{Value: int64(len(arg.Value))}
		default:
			// 不支持的类型：返回错误信息
			return newError("argument to `len` not supported, got %s",
				args[0].Type())
		}
	},
	},

	// puts 内置函数：输出所有参数到标准输出
	// 支持任意数量的参数，每个参数都会被转换为字符串输出
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 遍历所有参数，逐个输出到标准输出
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			// 返回 NULL 表示函数执行成功
			return NULL
		},
	},

	// first 内置函数：返回数组的第一个元素
	// 如果数组为空，返回 NULL
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 参数数量检查：first 函数只接受一个参数
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			// 参数类型检查：参数必须是数组类型
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			// 类型断言获取数组对象
			arr := args[0].(*object.Array)
			// 检查数组是否包含元素
			if len(arr.Elements) > 0 {
				// 返回第一个元素
				return arr.Elements[0]
			}

			// 空数组返回 NULL
			return NULL
		},
	},

	// last 内置函数：返回数组的最后一个元素
	// 如果数组为空，返回 NULL
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 参数数量检查：last 函数只接受一个参数
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			// 参数类型检查：参数必须是数组类型
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			// 类型断言获取数组对象
			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			// 检查数组是否包含元素
			if length > 0 {
				// 返回最后一个元素
				return arr.Elements[length-1]
			}

			// 空数组返回 NULL
			return NULL
		},
	},

	// rest 内置函数：返回除第一个元素外的数组剩余部分
	// 如果数组为空或只有一个元素，返回空数组
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 参数数量检查：rest 函数只接受一个参数
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			// 参数类型检查：参数必须是数组类型
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			// 类型断言获取数组对象
			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			// 检查数组是否包含多个元素
			if length > 0 {
				// 创建新数组，包含除第一个元素外的所有元素
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			// 空数组返回 NULL
			return NULL
		},
	},

	// push 内置函数：向数组末尾添加一个元素
	// 返回包含新元素的新数组，原数组保持不变
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			// 参数数量检查：push 函数需要两个参数（数组和要添加的元素）
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			// 参数类型检查：第一个参数必须是数组类型
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			// 类型断言获取数组对象
			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			// 创建新数组，长度比原数组多1
			newElements := make([]object.Object, length+1, length+1)
			// 复制原数组的所有元素
			copy(newElements, arr.Elements)
			// 在末尾添加新元素
			newElements[length] = args[1]

			// 返回包含新元素的新数组
			return &object.Array{Elements: newElements}
		},
	},
}
