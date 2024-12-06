package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/token"
)

const PROMPT = ">> "

// Start函数用于启动一个交互循环，从in读取输入，并将处理结果写入out。
// 参数in是一个io.Reader，用于读取输入。
// 参数out是一个io.Writer，用于输出结果。
func Start(in io.Reader, out io.Writer) {
	// 创建一个bufio.Scanner用于高效地读取输入。
	scanner := bufio.NewScanner(in)

	// 无限循环，直到输入结束。
	for {
		// 输出提示符。
		fmt.Fprintf(out, PROMPT)
		// 尝试扫描一行输入。
		scanned := scanner.Scan()
		// 如果扫描失败（例如输入结束），则退出循环。
		if !scanned {
			return
		}

		// 获取扫描到的输入行。
		line := scanner.Text()
		// 使用lexer包创建一个新的词法分析器，对输入行进行分析。
		l := lexer.New(line)

		// 循环获取输入行的下一个词(token)，直到达到EOF。
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			// 输出词的信息。
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}
