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

// Start函数是 REPL(Read-Eval-Print Loop)的入口点。
// 它接受一个输入流(in)和一个输出流(out)作为参数，用于读取用户输入和写入输出结果。
func Start(in io.Reader, out io.Writer) {
	// 创建一个扫描器，用于从输入流中读取数据。
	scanner := bufio.NewScanner(in)
	// 初始化一个新的环境对象，用于存储变量和函数等。
	env := object.NewEnvironment()

	for {
		// 向输出流写入提示符，提示用户输入。
		fmt.Fprintf(out, PROMPT)
		// 使用扫描器扫描用户输入。
		scanned := scanner.Scan()
		// 如果没有成功扫描到用户输入，意味着输入流已结束，此时退出循环。
		if !scanned {
			return
		}

		// 获取扫描到的用户输入文本。
		line := scanner.Text()
		// 创建一个新的词法分析器，用于将用户输入的文本转换为词法符号。
		l := lexer.New(line)
		// 创建一个新的语法分析器，用于将词法符号转换为抽象语法树。
		p := parser.New(l)

		// 使用语法分析器解析程序代码。
		program := p.ParseProgram()
		// 检查解析过程中是否有错误产生。
		if len(p.Errors()) != 0 {
			// 如果有错误，打印解析错误信息，并继续下一次循环，等待新的用户输入。
			printParserErrors(out, p.Errors())
			continue
		}

		// 使用解释器评估解析后的程序代码，并获取评估结果。
		evaluated := evaluator.Eval(program, env)
		// 如果评估结果不为空，将其转换为字符串并写入输出流。
		if evaluated != nil {
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

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
