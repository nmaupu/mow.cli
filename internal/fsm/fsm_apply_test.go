package fsm

import (
	"testing"

	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/matcher"
	"github.com/stretchr/testify/require"
)

//TODO: test Parse

func TestApplyTerminalStateNoArgs(t *testing.T) {
	s := NewState()
	s.Terminal = true

	ok := s.apply(nil, matcher.New())

	require.True(t, ok)
}

func TestApply(t *testing.T) {
	var (
		testArgs = []string{"1", "2", "3"}
		con      = &container.Container{}
		calls    = []string{}
	)
	matchers := map[string]matcher.Matcher{
		"a": TestMatcher{
			priority: 2,
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, testArgs, args)

				calls = append(calls, "a")

				c.Opts[con] = []string{"a.opt"}
				c.Args[con] = []string{"a.arg"}
				return true, args[1:]
			},
		},
		"b": TestMatcher{
			priority: 1,
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, testArgs, args)

				calls = append(calls, "b")

				c.Opts[con] = []string{"b.opt"}
				c.Args[con] = []string{"b.arg"}
				return true, args[1:]
			},
		},
		"c": TestMatcher{
			priority: 1,
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, testArgs[1:], args, "second stage matchers should be called with the reem args")

				calls = append(calls, "c")

				c.Opts[con] = []string{"c.opt"}
				c.Args[con] = []string{"c.arg"}
				return true, args[1:]
			},
		},
		"d": TestMatcher{
			priority: 1,
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, testArgs[1:], args, "second stage matchers should be called with the reem args")

				calls = append(calls, "d")

				c.Opts[con] = []string{"d.opt"}
				c.Args[con] = []string{"d.arg"}
				return false, args[1:]
			},
		},
		"e": TestMatcher{
			priority: 1,
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, testArgs[2:], args, "third stage matchers should be called with the reem args")

				calls = append(calls, "e")

				c.Opts[con] = []string{"e.opt"}
				c.Args[con] = []string{"e.arg"}
				return true, nil
			},
		},
	}
	s := mkFsm(`
		S1 a S2
		S1 b S3

		S2 c S4
		S3 d S4

		S4 e (S5)
	`, matchers)

	s.Prepare()

	context := matcher.New()
	ok := s.apply(testArgs, context)

	require.Equal(t, []string{"b", "a", "d", "c", "e"}, calls)

	require.True(t, ok)

	require.Equal(t, context.Opts[con], []string{"a.opt", "c.opt", "e.opt"})
	require.Equal(t, context.Args[con], []string{"a.arg", "c.arg", "e.arg"})
}

func TestApplyRejectOptions(t *testing.T) {
	var (
		testArgs = []string{"1", "--", "2"}
		calls    = []string{}
	)
	matchers := map[string]matcher.Matcher{
		"a": TestMatcher{
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, testArgs, args)
				require.False(t, c.RejectOptions)

				calls = append(calls, "a")
				return true, args[1:]
			},
		},
		"b": TestMatcher{
			match: func(args []string, c *matcher.ParseContext) (bool, []string) {
				require.Equal(t, []string{"2"}, args)
				require.True(t, c.RejectOptions)

				calls = append(calls, "b")
				return true, args[1:]
			},
		},
	}
	s := mkFsm(`
		S1 a S2
		S2 b S3
	`, matchers)

	s.Prepare()

	s.apply(testArgs, matcher.New())

	require.Equal(t, []string{"a", "b"}, calls)

}
