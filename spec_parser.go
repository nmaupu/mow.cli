package cli

import (
	"github.com/jawher/mow.cli/internal/fsm"
	"github.com/jawher/mow.cli/internal/lexer"
	"github.com/jawher/mow.cli/internal/parser"
)

func uParse(c *Cmd) (*fsm.State, error) {
	tokens, err := lexer.Tokenize(c.Spec)
	if err != nil {
		return nil, err
	}

	params := parser.Params{
		Spec:       c.Spec,
		Options:    c.options,
		OptionsIdx: c.optionsIdx,
		Args:       c.args,
		ArgsIdx:    c.argsIdx,
	}
	return parser.Parse(tokens, params)
}
