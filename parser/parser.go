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
	name       string
	firstToken *scanner.Token
	value      *big.Int
	children   []*ParseNode
}

func (node *ParseNode) String() string {
	return toString(node, 0)
}

// Recursively generates a string representing a tree of nodes
// where indentLevel indicates the level in the tree
func toString(node *ParseNode, indentLevel int) string {
	var s = "\n" + strings.Repeat(".", indentLevel)
	value := ""
	if node.value != nil {
		value = fmt.Sprintf("%d", node.value.Int64())
	}
	first := ""
	if node.firstToken != nil {
		first = node.firstToken.String()
	}
	s += fmt.Sprintf("^%s(%s)[%s]", node.name, first, value)
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

// Type Parser
type Parser struct {
	input         []scanner.Token
	err           error
	parseTreeRoot *ParseNode
}

func NewParser(scanResult scanner.ScanResult) *Parser {
	parser := new(Parser)
	parser.err = scanResult.Error
	parser.input = scanResult.Stream
	parser.parseTreeRoot = newParseNode("root expression")
	return parser
}

func (p *Parser) peekNextToken(errMsg string) (nextToken scanner.Token) {
	if len(p.input) == 0 {
		p.err = fmt.Errorf(errMsg)
		return
	}
	nextToken = (p.input)[0]
	return
}

// This method is similar to peekNextToken except that it returns a bool instead
// of setting the global error state when there is no next token. This is
// useful for case in which there being no next token is not an error.
func (p *Parser) checkNextToken() (nextToken scanner.Token, success bool) {
	if len(p.input) == 0 {
		success = false
		return
	}
	nextToken = (p.input)[0]
	success = true
	return
}

func (p *Parser) consumeNextToken() {
	p.input = p.input[1:] // Advance the input reader by one token
}

func (p *Parser) parseNumber(numberNode *ParseNode) {
	nextToken := p.peekNextToken("Unexpected end-of-input. Expecting a number.")
	if p.err != nil {
		return
	}
	switch nextToken.Kind {
	case scanner.TOKEN_NUMBER:
		numberNode.firstToken = &nextToken
		numberNode.value = nextToken.Value
		p.consumeNextToken()
		return
	default:
		p.err = fmt.Errorf("Unexpected token at position %d: %v. Expecting a number here.", nextToken.SourcePosition, nextToken)
	}
	return
}

func (p *Parser) parseFactor(factorNode *ParseNode) {
	nextToken := p.peekNextToken("Unexpected end-of-input. Expecting something to multiply.")
	if p.err != nil {
		return
	}
	factorNode.firstToken = &nextToken
	switch nextToken.Kind {
	case scanner.TOKEN_NUMBER:
		p.parseNumber(factorNode)
		return
	case scanner.TOKEN_MINUS:
		p.consumeNextToken()
		p.parseNumber(factorNode)
		if p.err == nil {
			factorNode.value.Neg(factorNode.value)
		}
		return
	case scanner.TOKEN_LPAREN:
		p.consumeNextToken()
		expressionNode := factorNode.appendExpressionChild()
		p.parseExpression(expressionNode)
		factorNode.value = expressionNode.value
		if p.err == nil {
			nextToken = p.peekNextToken("Unexpected end-of-input. Expecting a right parentheses at the end.")
			if p.err != nil {
				return
			}
			switch nextToken.Kind {
			case scanner.TOKEN_RPAREN:
				p.consumeNextToken()
				return
			default:
				p.err = fmt.Errorf("Expecting a closing paren ')' at position %d and instead found %v", nextToken.SourcePosition, nextToken)
			}
		}
	default:
		factorNode.firstToken = nil
		p.err = fmt.Errorf("Unexpected token at position %d: %v. Expecting something to multiply: a number or '('.", nextToken.SourcePosition, nextToken)
	}
	return
}

func (p *Parser) parseTermSuffix(termHead *ParseNode, node *ParseNode) {
	node.value = big.NewInt(1)
	nextToken, success := p.checkNextToken()
	if !success {
		return
	}
	switch nextToken.Kind {
	case scanner.TOKEN_TIMES:
		node.firstToken = &nextToken
		p.consumeNextToken()
		factorNode := node.appendFactorChild()
		termSuffixNode := node.appendTermSuffixChild()
		p.parseFactor(factorNode)
		if p.err == nil {
			p.parseTermSuffix(termHead, termSuffixNode)
		}
		if p.err == nil {
			node.value.Mul(factorNode.value, termSuffixNode.value)
		}
		return
	// FOLLOW(TERMSUFFIX) = {-, +, )}
	case scanner.TOKEN_MINUS, scanner.TOKEN_PLUS, scanner.TOKEN_RPAREN:
		// Take the epsilon transition.
		return
	default:
		p.err = fmt.Errorf("Extraneous token at position %d: %v,"+
			" while parsing the term that begins with %v at position %d.",
			nextToken.SourcePosition, nextToken, termHead.firstToken, termHead.firstToken.SourcePosition)
	}
	return
}

func (p *Parser) parseTerm(termNode *ParseNode) {
	nextToken := p.peekNextToken("Unexpected end-of-input. Expecting a term.")
	switch nextToken.Kind {
	case scanner.TOKEN_LPAREN, scanner.TOKEN_NUMBER, scanner.TOKEN_MINUS:
		termNode.firstToken = &nextToken
		factorNode := termNode.appendFactorChild()
		termSuffixNode := termNode.appendTermSuffixChild()
		p.parseFactor(factorNode)
		if p.err == nil {
			p.parseTermSuffix(termNode, termSuffixNode)
		}
		if p.err == nil {
			termNode.value = big.NewInt(1)
			termNode.value.Mul(factorNode.value, termSuffixNode.value)
		}
	default:
		p.err = fmt.Errorf("Unexpected token at position %d: %v. Expecting something to add or subtract: a number or '('.", nextToken.SourcePosition, nextToken)
	}
	return
}

func (p *Parser) parseExpressionSuffix(expressionHead *ParseNode, node *ParseNode) {
	node.value = big.NewInt(0)
	nextToken, success := p.checkNextToken()
	if !success {
		return
	}
	switch nextToken.Kind {
	case scanner.TOKEN_PLUS, scanner.TOKEN_MINUS:
		node.firstToken = &nextToken
		p.consumeNextToken()
		termNode := node.appendTermChild()
		expressionSuffixNode := node.appendExpressionSuffixChild()
		p.parseTerm(termNode)
		if p.err == nil {
			p.parseExpressionSuffix(expressionHead, expressionSuffixNode)
		}
		if p.err == nil {
			if nextToken.Kind == scanner.TOKEN_PLUS {
				node.value.Add(expressionSuffixNode.value, termNode.value)
			} else {
				node.value.Sub(expressionSuffixNode.value, termNode.value)
			}
		}

	// FOLLOW(EXPRSUFFIX) = {*, )}
	case scanner.TOKEN_TIMES, scanner.TOKEN_RPAREN:
		// Take the epsilon transition
		return
	default:
		panic("ASSERT: This line never reached because any extraneous tokens were already noticed by an instance of parseTokenSuffix.")
	}
	return
}

func (p *Parser) parseExpression(expressionNode *ParseNode) {
	nextToken := p.peekNextToken("Unexpected end-of-input. Expecting an expression.")
	if p.err != nil {
		return
	}
	switch nextToken.Kind {
	case scanner.TOKEN_LPAREN, scanner.TOKEN_NUMBER, scanner.TOKEN_MINUS:
		expressionNode.firstToken = &nextToken
		termNode := expressionNode.appendTermChild()
		expressionSuffixNode := expressionNode.appendExpressionSuffixChild()
		p.parseTerm(termNode)
		if p.err == nil {
			p.parseExpressionSuffix(expressionNode, expressionSuffixNode)
		}
		if p.err == nil {
			expressionNode.value = new(big.Int)
			expressionNode.value.Add(termNode.value, expressionSuffixNode.value)
		}
	default:
		p.err = fmt.Errorf("Unexpected token at position %d: %v. Expecting a number or '('.", nextToken.SourcePosition, nextToken)
	}
	return
}

func (p *Parser) parse() ParseResult {
	if p.err == nil {
		p.parseExpression(p.parseTreeRoot)
	}

	// Check if there are any extraneous tokens left in the stream.
	if p.err == nil {
		nextToken, success := p.checkNextToken()
		if success {
			p.err = fmt.Errorf("Extraneous token at position %d: %v", nextToken.SourcePosition, nextToken)
		}
	}

	return ParseResult{p.err, p.parseTreeRoot.value, p.parseTreeRoot}
}

type ParseResult struct {
	Error         error
	Result        *big.Int
	ParseTreeRoot *ParseNode
}

func Parse(input string) ParseResult {
	scanner := scanner.NewScanner()
	scanResult := scanner.Scan(input)
	parser := NewParser(scanResult)
	return parser.parse()
}
