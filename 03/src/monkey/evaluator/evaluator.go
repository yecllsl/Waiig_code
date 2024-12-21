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

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

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

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

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

func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
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

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
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
