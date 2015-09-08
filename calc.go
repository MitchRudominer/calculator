package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"regexp"
)

var lexRegExp = regexp.MustCompile(`^([ \(, \), \+, \-, \*, \^, 0-9, [[:space:]] ])*$`)

var bigTen = new(big.Int)

func init() {
	bigTen.SetInt64(10)
}

type TokenKind int

const (
	TOKEN_LPAREN TokenKind = iota // "("
	TOKEN_RPAREN                  // ")"
	TOKEN_PLUS                    // "+"
	TOKEN_MINUS                   // "-"
	TOKEN_TIMES                   // "*"
	TOKEN_POWER                   // "^"
	TOKEN_NUMBER                  //
	TOKEN_UNKNOWN
)

func (tokenKind TokenKind) String() string {
	switch tokenKind {
	case TOKEN_LPAREN:
		return "LPAREN"
	case TOKEN_RPAREN:
		return "RPAREN"
	case TOKEN_PLUS:
		return "PLUS"
	case TOKEN_MINUS:
		return "MINUS"
	case TOKEN_TIMES:
		return "TIMES"
	case TOKEN_POWER:
		return "POWER"
	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_UNKNOWN:
		return "UNKNOWN"
	default:
		return "unexpected value"
	}
}

type Token struct {
	kind           TokenKind
	value          big.Int // The value of a number token
	sourceString   string  // Only populated for TOKEN_UNKNOWN
	sourcePosition int
}

func (token Token) String() string {
	s := fmt.Sprintf("{%d}", token.sourcePosition)
	s += token.kind.String()
	if token.kind == TOKEN_NUMBER {
		s += fmt.Sprintf("(%v)", token.value.Int64())
	}
	if token.kind == TOKEN_UNKNOWN {
		s += fmt.Sprintf("(%v)", token.sourceString)
	}
	return s
}

func numberToken(value *big.Int) (t Token) {
	t.kind = TOKEN_NUMBER
	t.value.Set(value)
	return
}

func nonNumberToken(x rune) (t Token) {
	switch x {
	case '(':
		t.kind = TOKEN_LPAREN
	case ')':
		t.kind = TOKEN_RPAREN
	case '+':
		t.kind = TOKEN_PLUS
	case '-':
		t.kind = TOKEN_MINUS
	case '*':
		t.kind = TOKEN_TIMES
	case '^':
		t.kind = TOKEN_POWER
	default:
		t.kind = TOKEN_UNKNOWN
	}
	return
}

type AST struct {
}

type ParseResult struct {
	success      bool
	result       big.Float
	errorMessage string
	tree         AST
}

type ScanResult struct {
	success      bool
	errorMessage string
	stream       []Token
}

func isSpace(r rune) bool {
	if r <= '\u00FF' {
		// Obvious ASCII ones: \t through \r plus space. Plus two Latin-1 oddballs.
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	// High-valued ones.
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

func isDigit(r rune) (b bool, value int) {
	if r >= '0' && r <= '9' {
		b = true
		value = int(r) - int('0')
	}
	return
}

func Scan(input string) ScanResult {
	stream := make([]Token, len(input))
	tokenCount := 0
	currentlyParsingInt := false
	currentInt := new(big.Int)
	var currentIntSourcePosition int
	// We use a range loop because it returns runes instead of bytes
	for runePosition, nextRune := range input {
		if isSpace(nextRune) {
			if currentlyParsingInt {
				currentlyParsingInt = false
				stream[tokenCount] = numberToken(currentInt)
				stream[tokenCount].sourcePosition = currentIntSourcePosition
				tokenCount++
			}
			continue
		}
		isDigit, digitValue := isDigit(nextRune)
		if isDigit {
			bigDigitValue := new(big.Int)
			bigDigitValue.SetInt64(int64(digitValue))
			if currentlyParsingInt {
				currentInt.Add(currentInt.Mul(currentInt, bigTen), bigDigitValue)
			} else {
				currentlyParsingInt = true
				currentIntSourcePosition = runePosition
				currentInt.Set(bigDigitValue)
			}

		} else {
			if currentlyParsingInt {
				currentlyParsingInt = false
				stream[tokenCount] = numberToken(currentInt)
				stream[tokenCount].sourcePosition = currentIntSourcePosition
				tokenCount++
			}
			stream[tokenCount] = nonNumberToken(nextRune)
			stream[tokenCount].sourcePosition = runePosition
			if stream[tokenCount].kind == TOKEN_UNKNOWN {
				stream[tokenCount].sourceString = string(nextRune)
			}
			tokenCount++
		}
	}
	if currentlyParsingInt {
		stream[tokenCount] = numberToken(currentInt)
		stream[tokenCount].sourcePosition = currentIntSourcePosition
		tokenCount++
	}
	stream = stream[0:tokenCount]
	return ScanResult{success: true, stream: stream}
}

func parseExpression(tree *AST, input []Token) {
	fmt.Println(input)
	if len(input) == 0 {
		return
	}
	//lookAhead := input[0]
}

func parse(input string) *ParseResult {
	scanResult := Scan(input)
	if scanResult.success {
		tree := new(AST)
		parseExpression(tree, scanResult.stream)
	} else {
		fmt.Println(scanResult.errorMessage)
	}

	return new(ParseResult)
}

func main() {
	input_scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter an arithmetic expression: ")

		input_scanner.Scan()
		line := input_scanner.Text()

		parsed := parse(line)

		if parsed.success {
			fmt.Println(parsed.result)
		} else {
			fmt.Println(parsed.errorMessage)
		}
	}
}
