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
		if !parsed.Success {
			t.Errorf("Parse(%q) failed with message: %q.", c.in, parsed.ErrorMessage)
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
	EXPECTING_FACTOR_AT_END     = "Unexpected end-of-input. Expecting a factor."
)

func unexpectedToken(token string, position int) string {
	return fmt.Sprintf("Unexpected token at postion %d: %s", position, token)
}

func unrecognizedToken(token string, position int) string {
	return fmt.Sprintf("Unexpected token at postion %d: '%s'", position, token)
}

func TestParseFailure(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", EXPECTING_EXPRESSION_AT_END},
		{"     ", EXPECTING_EXPRESSION_AT_END},
		{"a", unrecognizedToken("a", 0)},
		{"     a", unrecognizedToken("a", 5)},
		{"+", unexpectedToken("+", 0)},
		{"     +", unexpectedToken("+", 5)},
		{"(", EXPECTING_EXPRESSION_AT_END},
		{")", unexpectedToken(")", 0)},
		{"()", unexpectedToken(")", 1)},
		{"((1)", EXPECTING_RPAREN_AT_END},
		{"((1)))", unexpectedToken(")", 5)},
		{"1---1", unexpectedToken("-", 3)},
		{"111 * -5 *", EXPECTING_FACTOR_AT_END},
		{"111 * *5", unexpectedToken("*", 6)},
		{"(666 + -4) * -11 + 17 + -3) * 5", unexpectedToken(")", 26)},
	}
	for _, c := range cases {
		parsed := Parse(c.in)
		if parsed.Success {
			t.Errorf("Parse(%q) unexpectedly succeeded.", c.in)
			continue
		}
		got := parsed.ErrorMessage
		if got != c.want {
			t.Errorf("Parse(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
