package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`^[[:space:]]*([-]?[0-9]+)(.+?)([-]?[0-9]+)[[:space:]]*$`)

func parseBigInt(s string) (x *big.Int, success bool) {
	x = new(big.Int)
	_, success = x.SetString(s, 10)

	if !success {
		fmt.Printf("Unable to parse %s as an integer.\n", s)
	}
	return
}

type operator struct {
	name string
	op   func(x, y *big.Int) *big.Int
}

func (oprtr *operator) matches(s string) bool {
	return oprtr.name == s
}

// Addition
func makePlusOperator() operator {
	return operator{"+", func(x, y *big.Int) (z *big.Int) {
		z = new(big.Int)
		z.Add(x, y)
		return
	}}
}

// Subraction
func makeMinusOperator() operator {
	return operator{"-", func(x, y *big.Int) (z *big.Int) {
		z = new(big.Int)
		z.Sub(x, y)
		return
	}}
}

// Multiplication
func makeTimesOperator() operator {
	return operator{"*", func(x, y *big.Int) (z *big.Int) {
		z = new(big.Int)
		z.Mul(x, y)
		return
	}}
}

// Exponentiation
func makePowerOperator() operator {
	return operator{"^", func(x, y *big.Int) (z *big.Int) {
		z = new(big.Int)
		z.Exp(x, y, nil)
		return
	}}
}

func main() {
	input_scanner := bufio.NewScanner(os.Stdin)
	operators := [4]operator{makePlusOperator(), makeMinusOperator(),
		makeTimesOperator(), makePowerOperator()}

	for {
		fmt.Print("Enter an expression of the form <integer> <operator> <integer>: ")

		input_scanner.Scan()
		line := input_scanner.Text()
		parts := re.FindStringSubmatch(line)

		if len(parts) != 4 {
			fmt.Println("Unable to parse that as <integer> <operator> <integer>")
			continue
		}

		x, success := parseBigInt(parts[1])
		y, success := parseBigInt(parts[3])

		if !success {
			continue
		}

		operator_string := strings.Trim(parts[2], " \t\n")

		var z *big.Int
		for _, oprtr := range operators {
			if oprtr.matches(operator_string) {
				z = oprtr.op(x, y)
			}
		}

		if z == nil {
			fmt.Printf("unknown operation \"%s\"\n", operator_string)
			continue
		}
		fmt.Println(z)
	}

	if err := input_scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
