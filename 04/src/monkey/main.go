package main

import (
	"fmt"
	"monkey/repl"
	"os"
	"os/user"
)

// main 函数是 Monkey 编程语言的入口点
// 它启动一个 REPL（Read-Eval-Print Loop）交互式环境
func main() {
	// 获取当前系统用户信息
	user, err := user.Current()
	if err != nil {
		// 如果获取用户信息失败，直接抛出异常
		panic(err)
	}

	// 打印欢迎信息，显示当前用户名
	fmt.Printf("Hello %s! This is the Monkey programming language!\n",
		user.Username)

	// 提示用户可以开始输入命令
	fmt.Printf("Feel free to type in commands\n")

	// 启动 REPL 环境，使用标准输入和标准输出
	repl.Start(os.Stdin, os.Stdout)
}
