package lexer

import (
	"testing"
)

func TestUTokenize(t *testing.T) {
	cases := []struct {
		usage    string
		expected []*Token
	}{
		{"OPTIONS", []*Token{{TTOptions, "OPTIONS", 0}}},

		{"XOPTIONS", []*Token{{TTPos, "XOPTIONS", 0}}},
		{"OPTIONSX", []*Token{{TTPos, "OPTIONSX", 0}}},
		{"ARG", []*Token{{TTPos, "ARG", 0}}},
		{"ARG42", []*Token{{TTPos, "ARG42", 0}}},
		{"ARG_EXTRA", []*Token{{TTPos, "ARG_EXTRA", 0}}},

		{"ARG1 ARG2", []*Token{{TTPos, "ARG1", 0}, {TTPos, "ARG2", 5}}},
		{"ARG1  ARG2", []*Token{{TTPos, "ARG1", 0}, {TTPos, "ARG2", 6}}},

		{"[ARG]", []*Token{{TTOpenSq, "[", 0}, {TTPos, "ARG", 1}, {TTCloseSq, "]", 4}}},
		{"[ ARG ]", []*Token{{TTOpenSq, "[", 0}, {TTPos, "ARG", 2}, {TTCloseSq, "]", 6}}},
		{"ARG [ARG2 ]", []*Token{{TTPos, "ARG", 0}, {TTOpenSq, "[", 4}, {TTPos, "ARG2", 5}, {TTCloseSq, "]", 10}}},
		{"ARG [ ARG2]", []*Token{{TTPos, "ARG", 0}, {TTOpenSq, "[", 4}, {TTPos, "ARG2", 6}, {TTCloseSq, "]", 10}}},

		{"...", []*Token{{TTRep, "...", 0}}},
		{"ARG...", []*Token{{TTPos, "ARG", 0}, {TTRep, "...", 3}}},
		{"ARG ...", []*Token{{TTPos, "ARG", 0}, {TTRep, "...", 4}}},
		{"[ARG...]", []*Token{{TTOpenSq, "[", 0}, {TTPos, "ARG", 1}, {TTRep, "...", 4}, {TTCloseSq, "]", 7}}},

		{"|", []*Token{{TTChoice, "|", 0}}},
		{"ARG|ARG2", []*Token{{TTPos, "ARG", 0}, {TTChoice, "|", 3}, {TTPos, "ARG2", 4}}},
		{"ARG |ARG2", []*Token{{TTPos, "ARG", 0}, {TTChoice, "|", 4}, {TTPos, "ARG2", 5}}},
		{"ARG| ARG2", []*Token{{TTPos, "ARG", 0}, {TTChoice, "|", 3}, {TTPos, "ARG2", 5}}},

		{"[OPTIONS]", []*Token{{TTOpenSq, "[", 0}, {TTOptions, "OPTIONS", 1}, {TTCloseSq, "]", 8}}},

		{"-p", []*Token{{TTShortOpt, "-p", 0}}},
		{"-X", []*Token{{TTShortOpt, "-X", 0}}},

		{"--force", []*Token{{TTLongOpt, "--force", 0}}},
		{"--sig-proxy", []*Token{{TTLongOpt, "--sig-proxy", 0}}},

		{"-aBc", []*Token{{TTOptSeq, "aBc", 1}}},
		{"--", []*Token{{TTDoubleDash, "--", 0}}},
		{"=<bla>", []*Token{{TTOptValue, "=<bla>", 0}}},
		{"=<bla-bla>", []*Token{{TTOptValue, "=<bla-bla>", 0}}},
		{"=<bla--bla>", []*Token{{TTOptValue, "=<bla--bla>", 0}}},
		{"-p=<file-path>", []*Token{{TTShortOpt, "-p", 0}, {TTOptValue, "=<file-path>", 2}}},
		{"--path=<absolute-path>", []*Token{{TTLongOpt, "--path", 0}, {TTOptValue, "=<absolute-path>", 6}}},
	}
	for _, c := range cases {
		t.Logf("test %s", c.usage)
		tks, err := Tokenize(c.usage)
		if err != nil {
			t.Errorf("[Tokenize '%s']: Unexpected error: %v", c.usage, err)
			continue
		}

		t.Logf("actual: %v\n", tks)
		if len(tks) != len(c.expected) {
			t.Errorf("[Tokenize '%s']: token count mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, c.expected, tks)
			continue
		}

		for i, actual := range tks {
			expected := c.expected[i]
			switch {
			case actual.Typ != expected.Typ:
				t.Errorf("[Tokenize '%s']: token type mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
			case actual.Val != expected.Val:
				t.Errorf("[Tokenize '%s']: token text mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
			case actual.Pos != expected.Pos:
				t.Errorf("[Tokenize '%s']: token pos mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, expected, actual)
			}
		}

	}
}

func TestUTokenizeErrors(t *testing.T) {
	cases := []struct {
		usage string
		pos   int
	}{
		{"-", 1},
		{"---x", 2},
		{"-x-", 2},

		{"=", 1},
		{"=<", 2},
		{"=<dsdf", 6},
		{"=<>", 2},
	}

	for _, c := range cases {
		t.Logf("test %s", c.usage)
		tks, err := Tokenize(c.usage)
		if err == nil {
			t.Errorf("Tokenize('%s') should have failed, instead got %v", c.usage, tks)
			continue
		}
		t.Logf("Got expected error %v", err)
		if err.Pos != c.pos {
			t.Errorf("[Tokenize '%s']: error pos mismatch:\n\tExpected: %v\n\tActual  : %v", c.usage, c.pos, err.Pos)

		}
	}
}
