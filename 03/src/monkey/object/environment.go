package object

// NewEnclosedEnvironment 创建一个新的环境，并将其外层环境设置为指定的环境。
// 这个函数用于在现有的环境链中添加一个新的层级，以便于实现作用域的嵌套。
// 参数:
//   outer *Environment - 指向外层环境的指针，表示新环境的外层环境。
// 返回值:
//   *Environment - 返回新创建的环境指针，该环境的外层环境被设置为指定的outer。
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// NewEnvironment 创建一个新的Environment实例。
// 该函数初始化一个空的环境，该环境不继承自任何外部环境。
// 返回值:
//
//	*Environment: 新创建的Environment指针，其内部存储被初始化为空，且没有外部环境。
func NewEnvironment() *Environment {
	// 初始化一个空的map，用于存储环境中的变量或对象。
	s := make(map[string]Object)
	// 返回一个新的Environment实例，其store字段设置为初始化的map，outer字段设置为nil。
	return &Environment{store: s, outer: nil}
}

// Environment 表示一个键值存储环境。
// 它用于通过名称存储和查找对象，并通过链式结构支持嵌套环境。
type Environment struct {
	// store 是一个映射，保存当前环境中的对象，键为对象的名称。
	store map[string]Object

	// outer 指向当前环境的外部环境，形成环境链。
	// 当在当前环境中找不到对象时，可以通过这个指针在外部环境中查找对象。
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
