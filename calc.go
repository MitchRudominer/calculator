package main

import (
	"bufio"
	"fmt"
	"github.com/rudominer/calculator/parser"
	"os"
)

func main() {
	input_scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter an arithmetic expression: ")

		input_scanner.Scan()
		line := input_scanner.Text()

		parsed := parser.Parse(line)

		if parsed.Error == nil {
			fmt.Println(parsed.Result)
			fmt.Println(parsed.ParseTreeRoot)
		} else {
			fmt.Println(parsed.Error)
			fmt.Println(parsed.ParseTreeRoot)
		}
	}
}
