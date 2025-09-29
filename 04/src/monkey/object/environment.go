package object

// NewEnclosedEnvironment 创建一个新的封闭环境，用于实现嵌套作用域
// 参数 outer: 外部环境指针，新创建的环境将继承该环境的变量查找能力
// 返回值: 指向新创建的封闭环境的指针
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// NewEnvironment 创建一个新的空环境
// 返回值: 指向新创建的环境的指针，包含空的变量存储映射
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// Environment 结构体表示Monkey语言中的变量环境
// 用于存储和管理变量名到对象的映射关系，支持嵌套作用域
type Environment struct {
	// store: 当前环境的变量存储映射，键为变量名，值为对应的Object对象
	store map[string]Object
	// outer: 指向外部环境的指针，用于实现变量查找的链式搜索（作用域链）
	outer *Environment
}

// Get 从环境中获取指定名称的变量值
// 参数 name: 要查找的变量名称
// 返回值:
//   - Object: 找到的变量值，如果未找到则返回nil
//   - bool: 指示是否成功找到变量
// 查找逻辑: 先在当前环境查找，如果未找到且存在外部环境，则递归到外部环境查找
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set 在当前环境中设置或修改变量值
// 参数 name: 变量名称
// 参数 val: 要设置的Object值
// 返回值: 设置的变量值
// 注意: 该方法只在当前环境设置变量，不会影响外部环境
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
