package parser

import "math/big"
import "github.com/rudominer/calculator/scanner"
import "fmt"
import "strings"

// EXPRESSION -> EXPRESSION + TERM | EXPRESSION - TERM | TERM
// TERM       -> TERM * FACTOR | FACTOR
// FACTOR     -> NUMBER | (EXPRESSION)

// Eliminating left recursion
//
// EXPRESSION -> TERM EXPRSUFFIX
// EXPRSUFFIX -> + TERM EXPRSUFFIX | - TERM EXPRSUFFIX | epsilon
// TERM       -> FACTOR TERMSUFFIX
// TERMSUFFIX -> * FACTOR TERMSUFFIX | epsilon
/// FACTOR    -> NUMBER | (EXPRESSION)

type ParseNode struct {
	name     string
	value    *big.Int
	children []*ParseNode
}

func (node *ParseNode) String() string {
	return toString(node, 0)
}

func toString(node *ParseNode, indentLevel int) string {
	var s = "\n" + strings.Repeat(".", indentLevel)
	s += fmt.Sprintf("^%s[%d]", node.name, node.value.Int64())
	if node.children != nil {
		for _, child := range node.children {
			s += toString(child, indentLevel+3)
		}
	}
	return s
}

func newParseNode(name string) *ParseNode {
	node := new(ParseNode)
	node.name = name
	return node
}

func (node *ParseNode) appendChild(name string) *ParseNode {
	child := newParseNode(name)
	node.children = append(node.children, child)
	return child
}

func (node *ParseNode) appendExpressionChild() *ParseNode {
	return node.appendChild("expression")
}

func (node *ParseNode) appendExpressionSuffixChild() *ParseNode {
	return node.appendChild("expressionSuffix")
}

func (node *ParseNode) appendTermChild() *ParseNode {
	return node.appendChild("term")
}

func (node *ParseNode) appendTermSuffixChild() *ParseNode {
	return node.appendChild("termSuffix")
}

func (node *ParseNode) appendFactorChild() *ParseNode {
	return node.appendChild("factor")
}

type ParseResult struct {
	Success       bool
	Result        *big.Int
	ErrorMessage  string
	ParseTreeRoot *ParseNode
}

func parseNumber(numberNode *ParseNode, input *[]scanner.Token) (success bool, errorMessage string) {
	if len(*input) == 0 {
		success = false
		errorMessage = fmt.Sprintf("Unexpected end-of-input. Expecting a number.")
		return
	}
	nextToken := (*input)[0]
	switch nextToken.Kind {
	case scanner.TOKEN_NUMBER:
		success = true
		numberNode.value = nextToken.Value
		*input = (*input)[1:] // Advance the input reader by one token
		return
	default:
		success = false
		errorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	return
}

func parseFactor(factorNode *ParseNode, input *[]scanner.Token) (success bool, errorMessage string) {
	if len(*input) == 0 {
		success = false
		errorMessage = fmt.Sprintf("Unexpected end-of-input. Expecting a factor.")
		return
	}
	nextToken := (*input)[0]
	switch nextToken.Kind {
	case scanner.TOKEN_NUMBER:
		success, errorMessage = parseNumber(factorNode, input)
		return
	case scanner.TOKEN_MINUS:
		*input = (*input)[1:] // Advance the input reader by one token
		success, errorMessage = parseNumber(factorNode, input)
		if success {
			factorNode.value.Neg(factorNode.value)
		}
		return
	case scanner.TOKEN_LPAREN:
		*input = (*input)[1:] // Advance the input reader by one token
		expressionNode := factorNode.appendExpressionChild()
		success, errorMessage = parseExpression(expressionNode, input)
		factorNode.value = expressionNode.value
		if success {
			if len(*input) == 0 {
				success = false
				errorMessage = fmt.Sprintf("Unexpected end-of-input. Expecting a right parentheses at the end.")
				return
			}
			nextToken := (*input)[0]
			switch nextToken.Kind {
			case scanner.TOKEN_RPAREN:
				success = true
				*input = (*input)[1:] // Advance the input reader by one token
				return
			default:
				success = false
				errorMessage = fmt.Sprintf("Expecting a closing paren ')' at position %d and instead found %v", nextToken.SourcePosition, nextToken)
			}
		}
	default:
		success = false
		errorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	return
}

func parseTermSuffix(node *ParseNode, input *[]scanner.Token) (success bool, errorMessage string) {
	node.value = big.NewInt(1)
	if len(*input) == 0 {
		success = true
		return
	}
	nextToken := (*input)[0]
	switch nextToken.Kind {
	case scanner.TOKEN_TIMES:
		*input = (*input)[1:] // Advance the input reader by one token
		factorNode := node.appendFactorChild()
		termSuffixNode := node.appendTermSuffixChild()
		success, errorMessage = parseFactor(factorNode, input)
		if success {
			success, errorMessage = parseTermSuffix(termSuffixNode, input)
		}
		if success {
			node.value.Mul(factorNode.value, termSuffixNode.value)
		}
		return
	case scanner.TOKEN_MINUS, scanner.TOKEN_PLUS, scanner.TOKEN_RPAREN:
		success = true
		return
	default:
		success = false
		errorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	return
}

func parseTerm(termNode *ParseNode, input *[]scanner.Token) (success bool, errorMessage string) {
	termNode.value = big.NewInt(1)
	if len(*input) == 0 {
		success = false
		errorMessage = fmt.Sprintf("Unexpected end-of-input. Expecting a term.")
		return
	}
	nextToken := (*input)[0]
	switch nextToken.Kind {
	case scanner.TOKEN_LPAREN, scanner.TOKEN_NUMBER, scanner.TOKEN_MINUS:
		factorNode := termNode.appendFactorChild()
		termSuffixNode := termNode.appendTermSuffixChild()
		success, errorMessage = parseFactor(factorNode, input)
		if success {
			success, errorMessage = parseTermSuffix(termSuffixNode, input)
		}
		if success {
			termNode.value.Mul(factorNode.value, termSuffixNode.value)
		}
	default:
		success = false
		errorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	return
}

func parseExpressionSuffix(node *ParseNode, input *[]scanner.Token) (success bool, errorMessage string) {
	node.value = big.NewInt(0)
	if len(*input) == 0 {
		success = true
		return
	}
	nextToken := (*input)[0]
	switch nextToken.Kind {
	case scanner.TOKEN_PLUS, scanner.TOKEN_MINUS:
		*input = (*input)[1:] // Advance the input reader by one token
		termNode := node.appendTermChild()
		expressionSuffixNode := node.appendExpressionSuffixChild()
		success, errorMessage = parseTerm(termNode, input)
		if success {
			success, errorMessage = parseExpressionSuffix(expressionSuffixNode, input)
		}
		if success {
			if nextToken.Kind == scanner.TOKEN_PLUS {
				node.value.Add(expressionSuffixNode.value, termNode.value)
			} else {
				node.value.Sub(expressionSuffixNode.value, termNode.value)
			}
		}

	case scanner.TOKEN_TIMES, scanner.TOKEN_RPAREN:
		success = true
		return
	default:
		success = false
		errorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	return
}

func parseExpression(expressionNode *ParseNode, input *[]scanner.Token) (success bool, errorMessage string) {
	expressionNode.value = big.NewInt(0)
	if len(*input) == 0 {
		success = false
		errorMessage = fmt.Sprintf("Unexpected end-of-input. Expecting an expression.")
		return
	}
	nextToken := (*input)[0]
	switch nextToken.Kind {
	case scanner.TOKEN_LPAREN, scanner.TOKEN_NUMBER, scanner.TOKEN_MINUS:
		termNode := expressionNode.appendTermChild()
		expressionSuffixNode := expressionNode.appendExpressionSuffixChild()
		success, errorMessage = parseTerm(termNode, input)
		if success {
			success, errorMessage = parseExpressionSuffix(expressionSuffixNode, input)
		}
		if success {
			expressionNode.value.Add(termNode.value, expressionSuffixNode.value)
		}
	default:
		success = false
		errorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	return
}

func Parse(input string) (parseResult ParseResult) {
	scanResult := scanner.Scan(input)
	parseResult.Success = scanResult.Success
	parseResult.ErrorMessage = scanResult.ErrorMessage
	if scanResult.Success {
		parseResult.ParseTreeRoot = newParseNode("root expression")
		parseResult.Success, parseResult.ErrorMessage = parseExpression(parseResult.ParseTreeRoot, &scanResult.Stream)
	}
	if len(scanResult.Stream) != 0 {
		parseResult.Success = false
		nextToken := scanResult.Stream[0]
		parseResult.ErrorMessage = fmt.Sprintf("Unexpected token at postion %d: %v", nextToken.SourcePosition, nextToken)
	}
	parseResult.Result = parseResult.ParseTreeRoot.value
	return
}
