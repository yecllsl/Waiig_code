package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

const PROMPT = ">> "

// Start 启动 Monkey 语言的 REPL（Read-Eval-Print Loop）交互式解释器
// 参数:
//   - in: 输入流，用于读取用户输入（通常为 os.Stdin）
//   - out: 输出流，用于显示结果和提示信息（通常为 os.Stdout）
//
// 功能说明:
//  1. 初始化词法分析器、语法分析器和求值器环境
//  2. 进入无限循环，持续接收用户输入并执行求值
//  3. 显示提示符 ">> " 等待用户输入
//  4. 对输入的代码进行完整的词法分析、语法分析和求值过程
//  5. 处理语法错误并显示友好的错误信息
//  6. 输出求值结果或错误信息
func Start(in io.Reader, out io.Writer) {
	// 创建输入扫描器，用于逐行读取用户输入
	scanner := bufio.NewScanner(in)
	// 创建新的求值环境，用于存储变量和函数定义
	env := object.NewEnvironment()

	// REPL 主循环：持续接收、解析和求值用户输入
	for {
		// 显示提示符，等待用户输入
		fmt.Fprintf(out, PROMPT)
		// 扫描用户输入，检查是否成功读取
		scanned := scanner.Scan()
		if !scanned {
			// 输入结束（如遇到 EOF），退出 REPL
			return
		}

		// 获取用户输入的代码行
		line := scanner.Text()
		// 创建词法分析器，将源代码转换为 token 序列
		l := lexer.New(line)
		// 创建语法分析器，将 token 序列转换为抽象语法树（AST）
		p := parser.New(l)

		// 解析程序，生成抽象语法树
		program := p.ParseProgram()
		// 检查语法错误
		if len(p.Errors()) != 0 {
			// 如果存在语法错误，显示错误信息并继续下一轮循环
			printParserErrors(out, p.Errors())
			continue
		}

		// 对抽象语法树进行求值，得到结果对象
		evaluated := evaluator.Eval(program, env)
		// 检查求值结果是否非空（nil 表示没有返回值或错误）
		if evaluated != nil {
			// 输出求值结果的字符串表示
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

// printParserErrors 显示语法分析错误的友好信息
// 参数:
//   - out: 输出流，用于显示错误信息（通常为 os.Stdout）
//   - errors: 语法错误消息字符串切片，包含所有检测到的语法错误
//
// 功能说明:
//  1. 显示猴子表情符号，增加错误信息的趣味性和可识别性
//  2. 输出通用的错误提示信息，表明遇到了语法问题
//  3. 列出所有具体的语法错误消息，每个错误缩进显示
//  4. 帮助用户快速定位和修复代码中的语法错误
//
// 设计特点:
//   - 用户友好：使用生动的语言和表情符号，避免技术术语的冰冷感
//   - 信息完整：显示所有检测到的语法错误，不遗漏任何问题
//   - 格式清晰：错误消息缩进显示，便于阅读和区分
func printParserErrors(out io.Writer, errors []string) {
	// 显示猴子表情符号，增加错误信息的趣味性
	io.WriteString(out, MONKEY_FACE)
	// 输出通用的错误提示标题
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	// 标识接下来的内容是语法错误列表
	io.WriteString(out, " parser errors:\n")
	// 遍历所有语法错误消息，逐个显示
	for _, msg := range errors {
		// 每个错误消息前添加制表符缩进，提高可读性
		io.WriteString(out, "\t"+msg+"\n")
	}
}
