package main

import (
	"fmt"
	"monkey/repl"
	"os"
	"os/user"
)

// main函数是程序的入口点。
// 它首先获取当前用户的信息，并向用户打招呼，然后启动REPL环境。
func main() {
	// 获取当前用户信息。
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	// 向用户显示欢迎信息。
	fmt.Printf("Hello %s! This is the Monkey programming language!\n",
		user.Username)
	// 提示用户可以开始输入命令。
	fmt.Printf("Feel free to type in commands\n")
	// 启动REPL环境，允许用户输入命令。
	repl.Start(os.Stdin, os.Stdout)
}
