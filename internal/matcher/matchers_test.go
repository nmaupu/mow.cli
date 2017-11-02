package matcher

import (
	"testing"

	"time"

	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/values"
	"github.com/stretchr/testify/require"
)

//TODO: test priority
//TODO: test that all matchers are comparable

func TestShortcut(t *testing.T) {
	pc := &ParseContext{}
	args := []string{"a", "b"}
	ok, nargs := theShortcut.Match(args, pc)
	require.True(t, ok, "shortcut always matches")
	require.Equal(t, args, nargs, "shortcut doesn't touch the passed args")
}

func TestOptsEnd(t *testing.T) {
	pc := &ParseContext{}
	args := []string{"a", "b"}
	ok, nargs := theOptsEnd.Match(args, pc)
	require.True(t, ok, "optsEnd always matches")
	require.Equal(t, args, nargs, "optsEnd doesn't touch the passed args")
	require.True(t, pc.RejectOptions, "optsEnd sets the rejectOptions flag")
}

func TestArgMatcher(t *testing.T) {
	arg := &container.Container{Name: "X"}
	argMatcher := arg{arg: arg}

	{
		pc := NewParseContext()
		args := []string{"a", "b"}
		ok, nargs := argMatcher.Match(args, &pc)
		require.True(t, ok, "arg should match")
		require.Equal(t, []string{"b"}, nargs, "arg should consume the matched value")
		require.Equal(t, []string{"a"}, pc.Args[arg], "arg should stored the matched value")
	}
	{
		pc := NewParseContext()
		ok, _ := argMatcher.Match([]string{"-v"}, &pc)
		require.False(t, ok, "arg should not match options")
	}
	{
		pc := NewParseContext()
		pc.RejectOptions = true
		ok, _ := argMatcher.Match([]string{"-v"}, &pc)
		require.True(t, ok, "arg should match options when the reject flag is set")
	}
}

func TestBoolOptMatcher(t *testing.T) {
	forceOpt := &container.Container{Names: []string{"-f", "--force"}, Value: values.NewBool(new(bool), false)}
	optMatcher := opt{
		theOne: forceOpt,
		index: map[string]*container.Container{
			"-f":      forceOpt,
			"--force": forceOpt,
			"-g":      {Names: []string{"-g"}, Value: values.NewBool(new(bool), false)},
			"-x":      {Names: []string{"-x"}, Value: values.NewBool(new(bool), false)},
			"-y":      {Names: []string{"-y"}, Value: values.NewBool(new(bool), false)},
		},
	}
	cases := []struct {
		args  []string
		nargs []string
		val   []string
	}{
		{[]string{"-f", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"-f=true", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"-f=false", "x"}, []string{"x"}, []string{"false"}},
		{[]string{"--force", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"--force=true", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"--force=false", "x"}, []string{"x"}, []string{"false"}},
		{[]string{"-fgxy", "x"}, []string{"-gxy", "x"}, []string{"true"}},
		{[]string{"-gfxy", "x"}, []string{"-gxy", "x"}, []string{"true"}},
		{[]string{"-gxfy", "x"}, []string{"-gxy", "x"}, []string{"true"}},
		{[]string{"-gxyf", "x"}, []string{"-gxy", "x"}, []string{"true"}},
	}
	for _, cas := range cases {
		t.Logf("Testing case: %#v", cas)
		pc := NewParseContext()
		ok, nargs := optMatcher.Match(cas.args, &pc)
		require.True(t, ok, "opt should match")
		require.Equal(t, cas.nargs, nargs, "opt should consume the option name")
		require.Equal(t, cas.val, pc.Opts[forceOpt], "true should stored as the option's value")

		pc = NewParseContext()
		pc.RejectOptions = true
		nok, _ := optMatcher.Match(cas.args, &pc)
		require.False(t, nok, "opt shouldn't match when rejectOptions flag is set")
	}
}

func TestOptMatcher(t *testing.T) {
	names := []string{"-f", "--force"}
	opts := []*container.Container{
		{Names: names, Value: values.NewString(new(string), "")},
		{Names: names, Value: values.NewInt(new(int), 0)},
		{Names: names, Value: values.NewStrings(new([]string), nil)},
		{Names: names, Value: values.NewInts(new([]int), nil)},
	}

	cases := []struct {
		args  []string
		nargs []string
		val   []string
	}{
		{[]string{"-f", "x"}, []string{}, []string{"x"}},
		{[]string{"-f=x", "y"}, []string{"y"}, []string{"x"}},
		{[]string{"-fx", "y"}, []string{"y"}, []string{"x"}},
		{[]string{"-afx", "y"}, []string{"-a", "y"}, []string{"x"}},
		{[]string{"-af", "x", "y"}, []string{"-a", "y"}, []string{"x"}},
		{[]string{"--force", "x"}, []string{}, []string{"x"}},
		{[]string{"--force=x", "y"}, []string{"y"}, []string{"x"}},
	}

	for _, cas := range cases {
		for _, forceOpt := range opts {
			t.Logf("Testing case: %#v with opt: %#v", cas, forceOpt)
			optMatcher := opt{
				theOne: forceOpt,
				index: map[string]*container.Container{
					"-f":      forceOpt,
					"--force": forceOpt,
					"-a":      {Names: []string{"-a"}, Value: values.NewBool(new(bool), false)},
				},
			}

			pc := NewParseContext()
			ok, nargs := optMatcher.Match(cas.args, &pc)
			require.True(t, ok, "opt %#v should match args %v, %v", forceOpt, cas.args, values.IsBool(forceOpt.Value))
			require.Equal(t, cas.nargs, nargs, "opt should consume the option name")
			require.Equal(t, cas.val, pc.Opts[forceOpt], "true should stored as the option's value")

			pc = NewParseContext()
			pc.RejectOptions = true
			nok, _ := optMatcher.Match(cas.args, &pc)
			require.False(t, nok, "opt shouldn't match when rejectOptions flag is set")
		}
	}
}

func TestOptsMatcher(t *testing.T) {
	opts := options{
		options: []*container.Container{
			{Names: []string{"-f", "--force"}, Value: values.NewBool(new(bool), false)},
			{Names: []string{"-g", "--green"}, Value: values.NewString(new(string), "")},
		},
		index: map[string]*container.Container{},
	}

	for _, o := range opts.options {
		for _, n := range o.Names {
			opts.index[n] = o
		}
	}

	cases := []struct {
		args  []string
		nargs []string
		val   [][]string
	}{
		{[]string{"-f", "x"}, []string{"x"}, [][]string{{"true"}, nil}},
		{[]string{"-f=false", "y"}, []string{"y"}, [][]string{{"false"}, nil}},
		{[]string{"--force", "x"}, []string{"x"}, [][]string{{"true"}, nil}},
		{[]string{"--force=false", "y"}, []string{"y"}, [][]string{{"false"}, nil}},

		{[]string{"-g", "x"}, []string{}, [][]string{nil, {"x"}}},
		{[]string{"-g=x", "y"}, []string{"y"}, [][]string{nil, {"x"}}},
		{[]string{"-gx", "y"}, []string{"y"}, [][]string{nil, {"x"}}},
		{[]string{"--green", "x"}, []string{}, [][]string{nil, {"x"}}},
		{[]string{"--green=x", "y"}, []string{"y"}, [][]string{nil, {"x"}}},

		{[]string{"-f", "-g", "x", "y"}, []string{"y"}, [][]string{{"true"}, {"x"}}},
		{[]string{"-g", "x", "-f", "y"}, []string{"y"}, [][]string{{"true"}, {"x"}}},
		{[]string{"-fg", "x", "y"}, []string{"y"}, [][]string{{"true"}, {"x"}}},
		{[]string{"-fgxxx", "y"}, []string{"y"}, [][]string{{"true"}, {"xxx"}}},
	}

	for _, cas := range cases {
		t.Logf("testing with args %v", cas.args)
		pc := NewParseContext()
		ok, nargs := opts.Match(cas.args, &pc)
		require.True(t, ok, "opts should match")
		require.Equal(t, cas.nargs, nargs, "opts should consume the option name")
		for i, opt := range opts.options {
			require.Equal(t, cas.val[i], pc.Opts[opt], "the option value for %v should be stored", opt)
		}

		pc = NewParseContext()
		pc.RejectOptions = true
		nok, _ := opts.Match(cas.args, &pc)
		require.False(t, nok, "opts shouldn't match when rejectOptions flag is set")
	}
}

// Issue 55
func TestOptsMatcherInfiniteLoop(t *testing.T) {
	opts := options{
		options: []*container.Container{
			{Names: []string{"-g"}, Value: values.NewString(new(string), ""), ValueSetFromEnv: true},
		},
		index: map[string]*container.Container{},
	}

	for _, o := range opts.options {
		for _, n := range o.Names {
			opts.index[n] = o
		}
	}

	done := make(chan struct{}, 1)
	pc := NewParseContext()
	go func() {
		opts.Match([]string{"-x"}, &pc)
		done <- struct{}{}
	}()

	select {
	case <-done:
		// nop, everything is good
	case <-time.After(5 * time.Second):
		t.Fatalf("Timed out after 5 seconds. Infinite loop in optsMatcher.")
	}

}
