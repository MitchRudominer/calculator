package parser

import (
	"fmt"
	"testing"
)

func TestParseSuccess(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"0", "0"},
		{"  0", "0"},
		{"0  ", "0"},
		{"  0  ", "0"},
		{"1", "1"},
		{"-1", "-1"},
		{"1+1", "2"},
		{"1+-1", "0"},
		{"1--1", "2"},
		{"-1+1", "0"},
		{"-1+-1", "-2"},
		{"(-5)", "-5"},
		{"(((-987654321098765432109876543210)))", "-987654321098765432109876543210"},
		{"11111111111111111111 * -5", "-55555555555555555555"},
		{"5 + 6 * 7", "47"},
		{"6 * 7 + 5", "47"},
		{"5 + (6 * 7)", "47"},
		{"(5 + (6 * 7))", "47"},
		{"((5) + (6 * (7)))", "47"},
		{"(5 + 6) * 7", "77"},
		{"5+6 * 7+8", "55"},
		{"(666 + -4) * -11 + (17 + -3) * 5", "-7212"},
	}
	for _, c := range cases {
		parsed := Parse(c.in)
		if parsed.Error != nil {
			t.Errorf("Parse(%q) failed with message: %v.", c.in, parsed.Error)
			continue
		}
		got := fmt.Sprintf("%v", parsed.Result)
		if got != c.want {
			t.Errorf("Parse(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

const (
	EXPECTING_EXPRESSION_AT_END = "Unexpected end-of-input. Expecting an expression."
	EXPECTING_RPAREN_AT_END     = "Unexpected end-of-input. Expecting a right parentheses at the end."
	EXPECTING_FACTOR_AT_END     = "Unexpected end-of-input. Expecting something to multiply."
)

func expectingNumber(token string, position int) string {
	return fmt.Sprintf("Unexpected token at position %d: %s. Expecting a number here.", position, token)
}

func expectingFactor(token string, position int) string {
	return fmt.Sprintf("Unexpected token at position %d: %s. Expecting something to multiply: a number or '('.", position, token)
}

func expectingExpression(token string, position int) string {
	return fmt.Sprintf("Unexpected token at position %d: %s. Expecting a number or '('.", position, token)
}

func expectingTerm(token string, position int) string {
	return fmt.Sprintf("Unexpected token at position %d: %s. Expecting something to add or subtract: a number or '('.", position, token)
}

func expectingEndOfExpression(token string, position int) string {
	return fmt.Sprintf("Unexpected token at position %d: %s. Expecting a number or ')'.", position, token)
}

func unexpectedToken(token string, position int) string {
	return fmt.Sprintf("Unexpected token at position %d: %s", position, token)
}

func extraneousToken(token string, position int) string {
	return fmt.Sprintf("Extraneous token at position %d: %s", position, token)
}

func extraneousTokenInTerm(token string, position, termStart, termStartPosition int) string {
	return fmt.Sprintf("Extraneous token at position %d: %v,"+
		" while parsing the term that begins with NUMBER(%d) at position %d.",
		position, token, termStart, termStartPosition)
}

func matches(err error, expected string) bool {
	return err.Error() == expected
}

func TestParseFailure(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", EXPECTING_EXPRESSION_AT_END},
		{"     ", EXPECTING_EXPRESSION_AT_END},
		{"a", expectingExpression("a", 0)},
		{"     a", expectingExpression("a", 5)},
		{"+", expectingExpression("+", 0)},
		{"     +", expectingExpression("+", 5)},
		{"(", EXPECTING_EXPRESSION_AT_END},
		{")", expectingExpression(")", 0)},
		{"()", expectingExpression(")", 1)},
		{"(67) + &", expectingTerm("&", 7)},
		{"((1)", EXPECTING_RPAREN_AT_END},
		{"((1)))", extraneousToken(")", 5)},
		{"1---1", expectingNumber("-", 3)},
		{"5 + 3 * 2 a + 7", extraneousTokenInTerm("a", 10, 3, 4)},
		{"7 * (1 + 2 a)", extraneousTokenInTerm("a", 11, 2, 9)},
		{"2 *", EXPECTING_FACTOR_AT_END},
		{"111 * -5 *", EXPECTING_FACTOR_AT_END},
		{"111 * -5 * +", expectingFactor("+", 11)},
		{"111 * -5 * )", expectingFactor(")", 11)},
		{"111 * *5", expectingFactor("*", 6)},
		{"(666 + -4) * -11 + 17 + -3) * 5", extraneousToken(")", 26)},
		{"(666 + -4) * -11 + 17 + -3 * )", expectingFactor(")", 29)},
	}
	for _, c := range cases {
		parsed := Parse(c.in)
		if parsed.Error == nil {
			t.Errorf("Parse(%q) unexpectedly succeeded.", c.in)
			continue
		}
		got := parsed.Error
		if !matches(got, c.want) {
			t.Errorf("Parse(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
