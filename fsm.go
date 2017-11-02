package cli

import (
	"sort"
	"strings"

	"fmt"

	"github.com/jawher/mow.cli/internal/matcher"
	"github.com/jawher/mow.cli/internal/values"
)

type state struct {
	id          int
	terminal    bool
	transitions transitions
}

type transition struct {
	matcher matcher.Matcher
	next    *state
}

type transitions []*transition

func (t transitions) Len() int      { return len(t) }
func (t transitions) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t transitions) Less(i, j int) bool {
	a, b := t[i].matcher, t[j].matcher
	return a.Priority() < b.Priority()
	//switch a.(type) {
	//case matcher.shortcut:
	//	return false
	//case matcher.optsEnd:
	//	return false
	//case *matcher.Arg:
	//	return false
	//default:
	//	return true
	//}
}

var _id = 0

func newState() *state {
	_id++
	return &state{id: _id, transitions: []*transition{}}
}

func (s *state) t(matcher matcher.Matcher, next *state) *state {
	s.transitions = append(s.transitions, &transition{matcher: matcher, next: next})
	return next
}

func (s *state) has(tr *transition) bool {
	for _, t := range s.transitions {
		if t.next == tr.next && t.matcher == tr.matcher {
			return true
		}
	}
	return false
}

func removeTransitionAt(idx int, arr transitions) transitions {
	res := make([]*transition, len(arr)-1)
	copy(res, arr[:idx])
	copy(res[idx:], arr[idx+1:])
	return res
}

func (s *state) simplify() {
	simplify(s, s, map[*state]bool{})
}

func simplify(start, s *state, visited map[*state]bool) {
	if visited[s] {
		return
	}
	visited[s] = true
	for _, tr := range s.transitions {
		simplify(start, tr.next, visited)
	}
	for s.simplifySelf(start) {
	}
}

func (s *state) simplifySelf(start *state) bool {
	for idx, tr := range s.transitions {
		if matcher.IsShortcut(tr.matcher) {
			next := tr.next
			s.transitions = removeTransitionAt(idx, s.transitions)
			for _, tr := range next.transitions {
				if !s.has(tr) {
					s.transitions = append(s.transitions, tr)
				}
			}
			if next.terminal {
				s.terminal = true
			}
			return true
		}
	}
	return false
}

func (s *state) dot() string {
	trs := dot(s, map[*state]bool{})
	return fmt.Sprintf("digraph G {\n\trankdir=LR\n%s\n}\n", strings.Join(trs, "\n"))
}

func dot(s *state, visited map[*state]bool) []string {
	res := []string{}
	if visited[s] {
		return res
	}
	visited[s] = true

	for _, tr := range s.transitions {
		res = append(res, fmt.Sprintf("\tS%d -> S%d [label=\"%v\"]", s.id, tr.next.id, tr.matcher))
		res = append(res, dot(tr.next, visited)...)
	}
	if s.terminal {
		res = append(res, fmt.Sprintf("\tS%d [peripheries=2]", s.id))
	}
	return res
}

func (s *state) parse(args []string) error {
	pc := matcher.NewParseContext()
	ok, err := s.apply(args, pc)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("incorrect usage")
	}

	for opt, vs := range pc.Opts {
		if multiValued, ok := opt.Value.(values.MultiValued); ok {
			multiValued.Clear()
			opt.ValueSetFromEnv = false
		}
		for _, v := range vs {
			if err := opt.Value.Set(v); err != nil {
				return err
			}
		}

		if opt.ValueSetByUser != nil {
			*opt.ValueSetByUser = true
		}
	}

	for arg, vs := range pc.Args {
		if multiValued, ok := arg.Value.(values.MultiValued); ok {
			multiValued.Clear()
			arg.ValueSetFromEnv = false
		}
		for _, v := range vs {
			if err := arg.Value.Set(v); err != nil {
				return err
			}
		}

		if arg.ValueSetByUser != nil {
			*arg.ValueSetByUser = true
		}
	}

	return nil
}

func (s *state) apply(args []string, pc matcher.ParseContext) (bool, error) {
	if s.terminal && len(args) == 0 {
		return true, nil
	}
	sort.Sort(s.transitions)

	if len(args) > 0 {
		arg := args[0]

		if !pc.RejectOptions && arg == "--" {
			pc.RejectOptions = true
			args = args[1:]
		}
	}

	type match struct {
		tr  *transition
		rem []string
		pc  matcher.ParseContext
	}

	matches := []*match{}
	for _, tr := range s.transitions {
		fresh := matcher.NewParseContext()
		fresh.RejectOptions = pc.RejectOptions
		if ok, rem := tr.matcher.Match(args, &fresh); ok {
			matches = append(matches, &match{tr, rem, fresh})
		}
	}

	for _, m := range matches {
		ok, err := m.tr.next.apply(m.rem, m.pc)
		if err != nil {
			return false, err
		}
		if ok {
			pc.Merge(m.pc)
			return true, nil
		}
	}
	return false, nil
}
