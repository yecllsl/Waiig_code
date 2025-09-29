package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

// 全局常量定义，表示Monkey语言中的基本值
var (
	NULL  = &object.Null{}                // 空值对象
	TRUE  = &object.Boolean{Value: true}  // 真布尔值对象
	FALSE = &object.Boolean{Value: false} // 假布尔值对象
)

// Eval 是求值器的入口函数，负责对AST节点进行求值
// 参数 node: 要求值的AST节点
// 参数 env: 当前执行环境（变量作用域）
// 返回值: 求值结果的对象
func Eval(node ast.Node, env *object.Environment) object.Object {
	// 使用类型switch根据节点类型进行不同的求值处理
	switch node := node.(type) {

	// 语句求值
	case *ast.Program:
		// 程序节点：按顺序求值所有语句
		return evalProgram(node, env)

	case *ast.BlockStatement:
		// 语句块节点：在独立作用域中求值语句序列
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		// 表达式语句节点：求值其包含的表达式
		return Eval(node.Expression, env)

	case *ast.ReturnStatement:
		// return语句：求值返回值并包装为ReturnValue对象
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		// let语句：求值赋值表达式并在环境中设置变量
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	// 表达式求值
	case *ast.IntegerLiteral:
		// 整数字面量：直接创建Integer对象
		return &object.Integer{Value: node.Value}

	case *ast.StringLiteral:
		// 字符串字面量：直接创建String对象
		return &object.String{Value: node.Value}

	case *ast.Boolean:
		// 布尔字面量：转换为Boolean对象
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		// 前缀表达式：先求值右侧表达式，再应用前缀运算符
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		// 中缀表达式：分别求值左右表达式，再应用中缀运算符
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		// if条件表达式：根据条件求值选择不同的分支
		return evalIfExpression(node, env)

	case *ast.Identifier:
		// 标识符：在环境中查找变量值或内置函数
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		// 函数字面量：创建Function对象（闭包）
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		// 函数调用：求值函数和参数，然后应用函数
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.ArrayLiteral:
		// 数组字面量：求值所有元素并创建Array对象
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		// 索引表达式：求值左侧（数组/哈希）和索引，然后进行索引操作
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		// 哈希字面量：求值所有键值对并创建Hash对象
		return evalHashLiteral(node, env)

	}

	return nil
}

// evalProgram 求值整个程序（语句序列）
// 参数 program: 程序AST节点
// 参数 env: 执行环境
// 返回值: 最后一个语句的求值结果（遇到return或error时提前返回）
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	// 按顺序求值所有语句
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		// 检查特殊返回值类型
		switch result := result.(type) {
		case *object.ReturnValue:
			// 遇到return语句，返回其值（解除包装）
			return result.Value
		case *object.Error:
			// 遇到错误，直接返回错误
			return result
		}
	}

	return result
}

// evalBlockStatement 求值语句块（创建新的作用域）
// 参数 block: 语句块AST节点
// 参数 env: 外部执行环境
// 返回值: 语句块中最后一个语句的求值结果
func evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object

	// 在语句块作用域中求值所有语句
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			// 如果遇到return或error，提前返回（不解除包装）
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// nativeBoolToBooleanObject 将Go布尔值转换为Monkey Boolean对象
// 参数 input: Go布尔值
// 返回值: 对应的Boolean对象（TRUE或FALSE）
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// evalPrefixExpression 求值前缀表达式
// 参数 operator: 前缀运算符（"!"或"-"）
// 参数 right: 右侧表达式求值结果
// 返回值: 应用前缀运算符后的结果
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		// 逻辑非运算符
		return evalBangOperatorExpression(right)
	case "-":
		// 负号运算符
		return evalMinusPrefixOperatorExpression(right)
	default:
		// 未知运算符错误
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// evalInfixExpression 求值中缀表达式
// 参数 operator: 中缀运算符
// 参数 left: 左侧表达式求值结果
// 参数 right: 右侧表达式求值结果
// 返回值: 应用中缀运算符后的结果
func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	// 根据操作数类型选择不同的求值策略
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		// 整数运算
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		// 字符串运算（仅支持连接）
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		// 相等比较（引用相等）
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		// 不等比较（引用不等）
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		// 类型不匹配错误
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		// 未知运算符错误
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evalBangOperatorExpression 求值逻辑非运算符表达式
// 参数 right: 右侧表达式求值结果
// 返回值: 逻辑非运算结果
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		// !true = false
		return FALSE
	case FALSE:
		// !false = true
		return TRUE
	case NULL:
		// !null = true
		return TRUE
	default:
		// 其他值视为真值，!truthy = false
		return FALSE
	}
}

// evalMinusPrefixOperatorExpression 求值负号运算符表达式
// 参数 right: 右侧表达式求值结果
// 返回值: 负号运算结果
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	// 检查操作数类型
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	// 对整数值取负
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// evalIntegerInfixExpression 求值整数中缀表达式
// 参数 operator: 运算符
// 参数 left: 左侧整数对象
// 参数 right: 右侧整数对象
// 返回值: 整数运算结果
func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	// 提取整数值
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	// 根据运算符进行相应的整数运算
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evalStringInfixExpression 求值字符串中缀表达式
// 参数 operator: 运算符
// 参数 left: 左侧字符串对象
// 参数 right: 右侧字符串对象
// 返回值: 字符串运算结果
func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	// 字符串只支持连接运算
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	// 字符串连接
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

// evalIfExpression 求值if条件表达式
// 参数 ie: if表达式AST节点
// 参数 env: 执行环境
// 返回值: 选择的分支求值结果
func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	// 求值条件表达式
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	// 根据条件真值选择分支
	if isTruthy(condition) {
		// 条件为真，执行consequence分支
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		// 条件为假且有else分支，执行alternative分支
		return Eval(ie.Alternative, env)
	} else {
		// 条件为假且无else分支，返回null
		return NULL
	}
}

// evalIdentifier 求值标识符表达式
// 参数 node: 标识符AST节点
// 参数 env: 执行环境
// 返回值: 变量值或内置函数对象
func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	// 在环境中查找变量
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// 在内置函数中查找
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	// 未找到标识符，返回错误
	return newError("identifier not found: " + node.Value)
}

// isTruthy 判断对象在条件表达式中的真值
// 参数 obj: 要判断的对象
// 返回值: 对象的真值（Monkey语言的truthy/falsy规则）
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		// null为假值
		return false
	case TRUE:
		// true为真值
		return true
	case FALSE:
		// false为假值
		return false
	default:
		// 其他所有值（非null、非false）都为真值
		return true
	}
}

// newError 创建错误对象
// 参数 format: 错误消息格式字符串
// 参数 a: 格式化参数
// 返回值: Error对象
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError 检查对象是否为错误对象
// 参数 obj: 要检查的对象
// 返回值: 如果是错误对象返回true
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

// evalExpressions 求值表达式列表
// 参数 exps: 表达式切片
// 参数 env: 执行环境
// 返回值: 求值结果的对象切片
func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	// 按顺序求值所有表达式
	for _, e := range exps {
		evaluated := Eval(e, env)
		// 如果遇到错误，立即返回错误（包装在切片中）
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// applyFunction 应用函数调用
// 参数 fn: 函数对象（Function或Builtin）
// 参数 args: 参数对象切片
// 返回值: 函数调用结果
func applyFunction(fn object.Object, args []object.Object) object.Object {
	// 根据函数类型进行不同的处理
	switch fn := fn.(type) {

	case *object.Function:
		// 用户定义函数：扩展环境并求值函数体
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		// 内置函数：直接调用函数实现
		return fn.Fn(args...)

	default:
		// 非函数对象错误
		return newError("not a function: %s", fn.Type())
	}
}

// extendFunctionEnv 扩展函数环境（创建闭包环境）
// 参数 fn: 函数对象
// 参数 args: 参数对象切片
// 返回值: 扩展后的新环境
func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	// 创建封闭环境（继承函数定义时的环境）
	env := object.NewEnclosedEnvironment(fn.Env)

	// 将参数绑定到新环境中
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// unwrapReturnValue 解除ReturnValue对象的包装
// 参数 obj: 可能包装了ReturnValue的对象
// 返回值: 解除包装后的实际值
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		// 返回实际的值（解除一层包装）
		return returnValue.Value
	}

	// 不是ReturnValue对象，直接返回
	return obj
}

// evalIndexExpression 求值索引表达式
// 参数 left: 左侧表达式求值结果（数组或哈希）
// 参数 index: 索引表达式求值结果
// 返回值: 索引操作结果
func evalIndexExpression(left, index object.Object) object.Object {
	// 根据左侧对象类型选择不同的索引策略
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		// 数组索引：使用整数索引访问元素
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		// 哈希索引：使用可哈希键访问值
		return evalHashIndexExpression(left, index)
	default:
		// 不支持的索引操作错误
		return newError("index operator not supported: %s", left.Type())
	}
}

// evalArrayIndexExpression 求值数组索引表达式
// 参数 array: 数组对象
// 参数 index: 整数索引对象
// 返回值: 索引位置的元素或null（越界时）
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	// 检查索引边界
	if idx < 0 || idx > max {
		// 索引越界，返回null
		return NULL
	}

	// 返回索引位置的元素
	return arrayObject.Elements[idx]
}

// evalHashLiteral 求值哈希字面量表达式
// 参数 node: 哈希字面量AST节点
// 参数 env: 执行环境
// 返回值: 哈希对象
func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	// 遍历所有键值对，分别求值
	for keyNode, valueNode := range node.Pairs {
		// 求值键表达式
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		// 检查键是否可哈希
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		// 求值值表达式
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		// 计算哈希键并存储键值对
		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

// evalHashIndexExpression 求值哈希索引表达式
// 参数 hash: 哈希对象
// 参数 index: 索引键对象
// 返回值: 对应键的值或null（不存在时）
func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	// 检查索引键是否可哈希
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	// 在哈希中查找键值对
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		// 键不存在，返回null
		return NULL
	}

	// 返回对应的值
	return pair.Value
}
