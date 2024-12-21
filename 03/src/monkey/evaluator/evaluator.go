package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

// 定义全局常量NULL、TRUE和FALSE，用于表示特定的布尔值和空值。
// 这些常量在程序中用作标准的真理值和空值引用，以提高代码的可读性和一致性。
// 布尔值只有true和false两种可能性，所以可以使用true和false的引用来代替每次都新建实例。
var (
	// NULL代表空值，通常用于表示没有值或者空对象。
	NULL = &object.Null{}
	// TRUE代表布尔值真，用于在布尔上下文中表示真值。
	TRUE = &object.Boolean{Value: true}
	// FALSE代表布尔值假，用于在布尔上下文中表示假值。
	FALSE = &object.Boolean{Value: false}
)

// Eval 函数是Monkey编程语言的解释器，负责遍历AST树并计算每个节点的值。
// 它接受一个AST节点和一个环境对象作为参数，返回计算结果对象。
// 参数:
//
//	node ast.Node: AST树的当前节点。
//	env *object.Environment: 当前执行环境，用于存储变量和函数。
//
// 返回值:
//
//	object.Object: 节点计算后的结果。
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// 处理语句类型
	case *ast.Program:
		// 评估程序中的所有语句
		return evalProgram(node, env)

	case *ast.BlockStatement:
		// 评估代码块中的所有语句
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		// 评估表达式语句的值
		return Eval(node.Expression, env)

	case *ast.ReturnStatement:
		// 评估返回语句的值，并创建ReturnValue对象
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		// 评估变量的值，并在环境中设置变量
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	// 处理表达式类型
	case *ast.IntegerLiteral:
		// 创建Integer对象，表示整数值
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		// 将布尔值转换为Boolean对象
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		// 评估前缀表达式的值
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		// 评估中缀表达式的值
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
		// 评估if表达式的值
		return evalIfExpression(node, env)

	case *ast.Identifier:
		// 评估标识符的值，从环境中获取变量
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		// 创建Function对象，表示函数字面量
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		// 评估调用表达式的值，包括函数和参数
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	}

	// 如果节点类型不匹配上述任何情况，返回nil
	return nil
}

// evalProgram 函数用于执行一个编译器程序，并返回执行结果。
// 它接受一个解析后的抽象语法树（AST）程序和一个环境对象作为参数。
// 参数 env 是执行程序时使用的环境，用于存储变量和函数定义等信息。
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	// 初始化结果变量，用于存储每条语句的执行结果。
	var result object.Object

	// 遍历程序中的所有语句，并依次执行。
	for _, statement := range program.Statements {
		// 对当前语句进行求值，并存储结果。
		result = Eval(statement, env)

		// 根据执行结果的类型，决定是否提前终止程序执行。
		switch result := result.(type) {
		case *object.ReturnValue:
			// 如果结果是返回值类型，则返回实际的返回值。
			return result.Value
		case *object.Error:
			// 如果结果是错误类型，则返回错误对象。
			return result
		}
	}

	// 如果所有语句执行完毕，没有遇到返回值或错误，则返回最后一条语句的执行结果。
	return result
}

// evalBlockStatement 评估一个代码块中的所有语句，并返回最后一个语句的评估结果。
// 如果遇到返回值或错误，则提前终止评估并返回。
// 参数:
//
//	block: 一个包含多个语句的代码块。
//	env: 执行语句的环境上下文。
//
// 返回值:
//
//	最后一个语句的评估结果，或者遇到的返回值/错误。
func evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object

	// 遍历代码块中的所有语句并评估它们。
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		// 检查评估结果是否为非空。
		if result != nil {
			rt := result.Type()
			// 如果结果类型为返回值或错误，则立即返回结果。
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	// 返回最后一个语句的评估结果。
	return result
}

// nativeBoolToBooleanObject 将 native bool 类型转换为 object.Boolean 类型的对象。
// 这个函数的存在是因为需要将内置的 bool 类型与系统中的 Boolean 对象类型进行桥接，
// 以便在系统内部统一处理布尔值。
// 参数:
//
//	input - 输入的 native bool 类型变量。
//
// 返回值:
//
//	*object.Boolean - 根据输入值返回对应的 Boolean 对象。
//	                 如果输入为 true，则返回 TRUE 对象；
//	                 如果输入为 false，则返回 FALSE 对象。
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// evalPrefixExpression 评估前缀表达式。
// 参数:
//
//	operator: 运算符，如 "!", "-"。
//	right: 表达式的右侧对象。
//
// 返回值:
//
//	评估后的对象，或在运算符未知时返回错误对象。
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		// 评估带 "!" 运算符的表达式。
		return evalBangOperatorExpression(right)
	case "-":
		// 评估带 "-" 前缀运算符的表达式。
		return evalMinusPrefixOperatorExpression(right)
	default:
		// 当遇到未知运算符时，生成并返回错误对象。
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// evalInfixExpression 评估中缀表达式。
// 参数:
// - operator: 表达式中的操作符，如 "+", "-", "*", "/", "==" 等。
// - left: 表达式中的左操作数。
// - right: 表达式中的右操作数。
// 返回值:
// - object.Object: 表达式的评估结果，具体类型取决于操作数和操作符。
func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	// 根据操作数的类型和操作符选择合适的处理方式。
	switch {
	// 当左右操作数均为整数时，调用evalIntegerInfixExpression进行计算。
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	// 当操作符为"=="时，比较左右操作数是否相等。
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	// 当操作符为"!="时，比较左右操作数是否不相等。
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	// 当左右操作数类型不同时，抛出类型不匹配错误。
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	// 当操作符未知或不支持时，抛出未知操作符错误。
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evalBangOperatorExpression 评估带逻辑非操作符的表达式。
// 该函数根据传入的右操作数的值，返回相应的布尔值。
// 参数:
//
//	right object.Object: 右操作数对象，其类型为 object.Object。
//
// 返回值:
//
//	object.Object: 根据右操作数的值返回 TRUE 或 FALSE。
//
// 逻辑非操作符的语义如下：
// - 如果右操作数为 TRUE，则返回 FALSE。
// - 如果右操作数为 FALSE 或 NULL，则返回 TRUE。
// - 对于其他情况，默认返回 FALSE。
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// evalMinusPrefixOperatorExpression 评估带有负号前缀操作符的表达式。
// 这个函数接受一个对象作为参数，如果该对象不是整数类型，则返回错误。
// 如果是整数类型，它将返回一个新的整数对象，其值为原值的负数。
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	// 检查 right 对象的类型是否为整数类型，如果不是，则返回错误。
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	// 将 right 对象转换为 Integer 类型，并获取其值。
	value := right.(*object.Integer).Value

	// 返回一个新的整数对象，其值为原值的负数。
	return &object.Integer{Value: -value}
}

// evalIntegerInfixExpression 评估两个整数对象的中缀表达式。
// 它根据提供的操作符执行相应的数学或比较操作。
// 参数:
// - operator: 字符串类型，定义了要执行的操作（如"+", "-", "*", "/", "<", ">", "==", "!="）。
// - left: 左侧操作数，类型为object.Object，预期为object.Integer的实例。
// - right: 右侧操作数，类型为object.Object，预期为object.Integer的实例。
// 返回值:
//   - 返回一个object.Object类型的结果，具体类型取决于操作符。
//     对于算术操作符（"+", "-", "*", "/"），返回一个object.Integer实例。
//     对于比较操作符（"<", ">", "==", "!="），返回一个布尔值的封装对象。
//     如果操作符未知，返回一个错误对象。
func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	// 提取左侧和右侧操作数的整数值。
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	// 根据操作符执行相应的操作。
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
		// 如果操作符不被支持，则返回错误。
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evalIfExpression 评估 if 表达式并返回相应的结果对象。
// 该函数首先评估条件表达式的值，如果条件表达式评估出错，则直接返回错误。
// 如果条件表达式为真，则评估并返回后果表达式（Consequence）的结果。
// 如果条件表达式为假且存在替代表达式（Alternative），则评估并返回替代表达式的结果。
// 如果条件表达式为假且不存在替代表达式，则返回空对象。
//
// 参数:
// - ie *ast.IfExpression: if 表达式的语法树表示。
// - env *object.Environment: 执行环境，包含变量和函数的定义。
//
// 返回值:
// - object.Object: 评估结果对象，可能是任何类型的对象，包括错误对象。
func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	// 评估条件表达式的值
	condition := Eval(ie.Condition, env)
	// 如果条件表达式评估出错，直接返回错误
	if isError(condition) {
		return condition
	}

	// 如果条件表达式为真，评估并返回后果表达式的结果
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		// 如果条件表达式为假且存在替代表达式，评估并返回替代表达式的结果
		return Eval(ie.Alternative, env)
	} else {
		// 如果条件表达式为假且不存在替代表达式，返回空对象
		return NULL
	}
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}

	return val
}

// isTruthy 判断给定的对象是否为"真值"。
// 在这个函数中，"真值"的定义是除了NULL和FALSE之外的任何对象。
// 参数:
//
//	obj object.Object: 待评估的对象。
//
// 返回值:
//
//	bool: 如果对象被认为是"真值"，则返回true；否则返回false。
func isTruthy(obj object.Object) bool {
	// 根据对象的类型进行判断。
	switch obj {
	case NULL:
		// NULL被定义为非"真值"。
		return false
	case TRUE:
		// TRUE被定义为"真值"。
		return true
	case FALSE:
		// FALSE被定义为非"真值"。
		return false
	default:
		// 除了NULL和FALSE之外的任何对象都被认为是"真值"。
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	extendedEnv := extendFunctionEnv(function, args)
	evaluated := Eval(function.Body, extendedEnv)
	return unwrapReturnValue(evaluated)
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}
