package scanner

import (
	"fmt"
	"math/big"
	"regexp"
)

var lexRegExp = regexp.MustCompile(`^([ \(, \), \+, \-, \*, \^, 0-9, [[:space:]] ])*$`)

var bigTen = big.NewInt(10)

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
	Kind           TokenKind
	Value          *big.Int // The value of a number token
	SourceString   string   // Only populated for TOKEN_UNKNOWN
	SourcePosition int
}

func (token Token) String() string {
	s := fmt.Sprintf("{%d}", token.SourcePosition)
	s += token.Kind.String()
	if token.Kind == TOKEN_NUMBER {
		s += fmt.Sprintf("(%v)", token.Value.Int64())
	}
	if token.Kind == TOKEN_UNKNOWN {
		s += fmt.Sprintf("(%v)", token.SourceString)
	}
	return s
}

func numberToken(value *big.Int) (t Token) {
	t.Kind = TOKEN_NUMBER
	t.Value = new(big.Int)
	t.Value.Set(value)
	return
}

func nonNumberToken(x rune) (t Token) {
	switch x {
	case '(':
		t.Kind = TOKEN_LPAREN
	case ')':
		t.Kind = TOKEN_RPAREN
	case '+':
		t.Kind = TOKEN_PLUS
	case '-':
		t.Kind = TOKEN_MINUS
	case '*':
		t.Kind = TOKEN_TIMES
	case '^':
		t.Kind = TOKEN_POWER
	default:
		t.Kind = TOKEN_UNKNOWN
	}
	return
}

type ScanResult struct {
	Success      bool
	ErrorMessage string
	Stream       []Token
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
				stream[tokenCount].SourcePosition = currentIntSourcePosition
				tokenCount++
			}
			continue
		}
		isDigit, digitValue := isDigit(nextRune)
		if isDigit {
			bigDigitValue := big.NewInt(int64(digitValue))
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
				stream[tokenCount].SourcePosition = currentIntSourcePosition
				tokenCount++
			}
			stream[tokenCount] = nonNumberToken(nextRune)
			stream[tokenCount].SourcePosition = runePosition
			if stream[tokenCount].Kind == TOKEN_UNKNOWN {
				stream[tokenCount].SourceString = string(nextRune)
			}
			tokenCount++
		}
	}
	if currentlyParsingInt {
		stream[tokenCount] = numberToken(currentInt)
		stream[tokenCount].SourcePosition = currentIntSourcePosition
		tokenCount++
	}
	stream = stream[0:tokenCount]
	return ScanResult{Success: true, Stream: stream}
}
