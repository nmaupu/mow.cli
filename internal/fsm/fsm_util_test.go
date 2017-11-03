package fsm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jawher/mow.cli/internal/matcher"
)

var (
	testMatchers = map[string]matcher.Matcher{
		"*":         matcher.NewShortcut(),
		"!":         NopeMatcher{},
		"-a":        matcher.NewOpt(nil, nil),
		"ARG":       matcher.NewArg(nil),
		"--":        matcher.NewOptsEnd(),
		"[OPTIONS]": matcher.NewOptions(nil, nil),
	}
)

type NopeMatcher struct{}

func (NopeMatcher) Match(args []string, c *matcher.ParseContext) (bool, []string) {
	return false, args
}

func (NopeMatcher) Priority() int {
	return 666
}

type YepMatcher struct{}

func (YepMatcher) Match(args []string, c *matcher.ParseContext) (bool, []string) {
	return true, args
}

func (YepMatcher) Priority() int {
	return 666
}

type TestMatcher struct {
	match    func(args []string, c *matcher.ParseContext) (bool, []string)
	priority int
}

func (t TestMatcher) Match(args []string, c *matcher.ParseContext) (bool, []string) {
	return t.match(args, c)
}

func (t TestMatcher) Priority() int {
	return t.priority
}

func mkFsm(spec string, matchers map[string]matcher.Matcher) *State {
	states := map[string]*State{}
	lines := strings.FieldsFunc(spec, func(r rune) bool { return r == '\n' })

	var res *State = nil

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			panic(fmt.Sprintf("Invalid line %q: syntax: START TR END", line))
		}
		sn, tn, en := parts[0], parts[1], parts[2]
		sn, sterm := stateNameTerm(sn)
		en, eterm := stateNameTerm(en)

		s, ok := states[sn]
		if !ok {
			s = newStateWithName(sn)
			states[sn] = s
		}
		s.Terminal = s.Terminal || sterm

		if res == nil {
			res = s
		}

		e, ok := states[en]
		if !ok {
			e = newStateWithName(en)
			states[en] = e
		}
		e.Terminal = e.Terminal || eterm

		t, ok := matchers[tn]
		if !ok {
			panic(fmt.Sprintf("Unknown matcher %q in line %q", tn, line))
		}

		s.T(t, e)
	}
	return res
}

func mkTransitionStrs(trs transitions) []string {
	res := []string{}
	for _, tr := range trs {
		res = append(res, trName(tr.Matcher))
	}
	return res
}

func stateNameTerm(name string) (string, bool) {
	if strings.HasPrefix(name, "(") {
		if strings.HasSuffix(name, ")") {
			name = name[1 : len(name)-1]
			return name, true
		}
		panic(fmt.Sprintf("Invalid state name %q", name))
	}
	return name, false
}
func stateName(s *State) string {
	if !s.Terminal {
		return fmt.Sprintf("S%d", s.id)
	}
	return fmt.Sprintf("(S%d)", s.id)
}

func fsmStr(s *State) string {
	lines := fsmStrVis(s, map[*State]struct{}{})
	sort.Sort(lines)
	return strings.Join(lines, "\n")
}

func fsmStrVis(s *State, visited map[*State]struct{}) fsmStrings {
	if _, ok := visited[s]; ok {
		return nil
	}
	visited[s] = struct{}{}

	res := fsmStrings{}
	for _, tr := range s.Transitions {
		res = append(res, fmt.Sprintf("%s %s %s", stateName(s), trName(tr.Matcher), stateName(tr.Next)))
		res = append(res, fsmStrVis(tr.Next, visited)...)
	}

	return res
}

func trName(m matcher.Matcher) string {
	for name, x := range testMatchers {
		if x == m {
			return name
		}
	}
	panic(fmt.Sprintf("No such matcher %v", m))
}

type fsmStrings []string

func (t fsmStrings) Len() int      { return len(t) }
func (t fsmStrings) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t fsmStrings) Less(i, j int) bool {
	a := strings.TrimFunc(t[i], isParen)
	b := strings.TrimFunc(t[j], isParen)
	return strings.Compare(a, b) < 0
}

func isParen(r rune) bool {
	return r == '(' || r == ')'
}

func newStateWithName(name string) *State {
	res := NewState()
	sid := strings.TrimPrefix(name, "S")
	id, err := strconv.Atoi(sid)
	if err != nil {
		panic(fmt.Sprintf("State name must be S%d", name))
	}
	res.id = id
	return res
}
