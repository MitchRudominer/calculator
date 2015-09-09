package scanner

import (
	"fmt"
	"math/big"
)

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
		return "("
	case TOKEN_RPAREN:
		return ")"
	case TOKEN_PLUS:
		return "+"
	case TOKEN_MINUS:
		return "-"
	case TOKEN_TIMES:
		return "*"
	case TOKEN_POWER:
		return "^"
	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_UNKNOWN:
		return "UNKNOWN"
	default:
		panic(fmt.Sprintf("Invalid TokenKind: %v", tokenKind))
	}
}

type Token struct {
	Kind           TokenKind
	Value          *big.Int // The value of a number token
	SourceString   string   // Only populated for TOKEN_UNKNOWN
	SourcePosition int
}

func (token Token) String() string {
	switch token.Kind {
	case TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_PLUS, TOKEN_MINUS, TOKEN_TIMES, TOKEN_POWER:
		return token.Kind.String()
	case TOKEN_NUMBER:
		return fmt.Sprintf("NUMBER(%v)", token.Value.Int64())
	case TOKEN_UNKNOWN:
		return fmt.Sprintf("'%v'", token.SourceString)
	default:
		panic(fmt.Sprintf("Invalid TokenKind: %v", token.Kind))
	}
}

func (token Token) DebugString() string {
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

// This code is copied from golang.org
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

type Scanner struct {
	stream                   []Token
	tokenCount               int
	currentlyParsingInt      bool
	currentInt               *big.Int
	currentIntSourcePosition int
}

func NewScanner() *Scanner {
	scanner := new(Scanner)
	scanner.currentInt = new(big.Int)
	return scanner
}

func (s *Scanner) startIntParse(position int, value *big.Int) {
	s.currentlyParsingInt = true
	s.currentIntSourcePosition = position
	s.currentInt.Set(value)
}

func (s *Scanner) continueIntParse(value *big.Int) bool {
	if s.currentlyParsingInt {
		s.currentInt.Add(s.currentInt.Mul(s.currentInt, bigTen), value)
		return true
	}
	return false
}

func (s *Scanner) endIntParse() bool {
	if s.currentlyParsingInt {
		s.currentlyParsingInt = false
		s.stream[s.tokenCount] = numberToken(s.currentInt)
		s.stream[s.tokenCount].SourcePosition = s.currentIntSourcePosition
		s.tokenCount++
		return true
	}
	return false
}

func (s *Scanner) Scan(input string) ScanResult {
	s.stream = make([]Token, len(input))
	// We use a range loop because it returns runes instead of bytes
	for runePosition, nextRune := range input {
		if isSpace(nextRune) {
			s.endIntParse()
			continue
		}
		isDigit, digitValue := isDigit(nextRune)
		if isDigit {
			bigDigitValue := big.NewInt(int64(digitValue))
			if !s.continueIntParse(bigDigitValue) {
				s.startIntParse(runePosition, bigDigitValue)
			}

		} else {
			// Handle a non-digit rune.
			s.endIntParse()
			s.stream[s.tokenCount] = nonNumberToken(nextRune)
			token := &s.stream[s.tokenCount]
			token.SourcePosition = runePosition
			if token.Kind == TOKEN_UNKNOWN {
				token.SourceString = string(nextRune)
			}
			s.tokenCount++
		}
	}
	s.endIntParse()
	s.stream = s.stream[0:s.tokenCount]
	return ScanResult{Success: true, Stream: s.stream}
}
