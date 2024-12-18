package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/parser"
)

const PROMPT = ">> "

// Start 函数是 REPL(Read-Eval-Print Loop) 的入口点。
// 它从 in 读取输入，并将输出写入 out。
// 参数 in 是一个 io.Reader 类型，用于读取输入。
// 参数 out 是一个 io.Writer 类型，用于写入输出。
func Start(in io.Reader, out io.Writer) {
	// 创建一个 bufio.Scanner 来读取输入。
	scanner := bufio.NewScanner(in)

	// 主循环，用于持续读取输入和输出提示符。
	for {
		// 向输出写入提示符。
		fmt.Fprintf(out, PROMPT)
		// 扫描输入。
		scanned := scanner.Scan()
		// 如果没有成功扫描到输入，退出循环。
		if !scanned {
			return
		}

		// 获取扫描到的输入行。
		line := scanner.Text()
		// 创建一个词法分析器，用于分析输入行。
		l := lexer.New(line)
		// 创建一个解析器，用于解析程序。
		p := parser.New(l)

		// 解析输入行为程序。
		program := p.ParseProgram()
		// 如果有解析错误，打印错误并继续下一次循环。
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		// 将解析后的程序写入输出。
		io.WriteString(out, program.String())
		// 写入换行符。
		io.WriteString(out, "\n")
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

// printParserErrors 输出解析错误信息到指定的输出流。
// 该函数旨在向用户提供清晰、友好的错误反馈，以便用户能够快速识别解析过程中遇到的问题。
// 参数:
//   - out: 错误信息的输出流，通常是一个文件或标准输出。
//   - errors: 包含解析错误信息的切片。
func printParserErrors(out io.Writer, errors []string) {
	// 输出一个有趣的表情符号和错误信息的开头部分，以吸引用户的注意。
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")

	// 遍历错误信息切片，逐个输出每条错误信息。
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
