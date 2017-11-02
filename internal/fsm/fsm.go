package fsm

import (
	"sort"
	"strings"

	"fmt"

	"github.com/jawher/mow.cli/internal/matcher"
	"github.com/jawher/mow.cli/internal/values"
)

type State struct {
	Terminal    bool
	Transitions transitions
	id          int
}

type Transition struct {
	Matcher matcher.Matcher
	Next    *State
}

type transitions []*Transition

func (t transitions) Len() int      { return len(t) }
func (t transitions) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t transitions) Less(i, j int) bool {
	a, b := t[i].Matcher, t[j].Matcher
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

func NewState() *State {
	_id++
	return &State{id: _id, Transitions: []*Transition{}}
}

func (s *State) T(matcher matcher.Matcher, next *State) *State {
	s.Transitions = append(s.Transitions, &Transition{Matcher: matcher, Next: next})
	return next
}

func (s *State) has(tr *Transition) bool {
	for _, t := range s.Transitions {
		if t.Next == tr.Next && t.Matcher == tr.Matcher {
			return true
		}
	}
	return false
}

func removeTransitionAt(idx int, arr transitions) transitions {
	res := make([]*Transition, len(arr)-1)
	copy(res, arr[:idx])
	copy(res[idx:], arr[idx+1:])
	return res
}

func (s *State) Simplify() {
	simplify(s, s, map[*State]bool{})
}

func simplify(start, s *State, visited map[*State]bool) {
	if visited[s] {
		return
	}
	visited[s] = true
	for _, tr := range s.Transitions {
		simplify(start, tr.Next, visited)
	}
	for s.simplifySelf(start) {
	}
}

func (s *State) simplifySelf(start *State) bool {
	for idx, tr := range s.Transitions {
		if matcher.IsShortcut(tr.Matcher) {
			next := tr.Next
			s.Transitions = removeTransitionAt(idx, s.Transitions)
			for _, tr := range next.Transitions {
				if !s.has(tr) {
					s.Transitions = append(s.Transitions, tr)
				}
			}
			if next.Terminal {
				s.Terminal = true
			}
			return true
		}
	}
	return false
}

func (s *State) dot() string {
	trs := dot(s, map[*State]bool{})
	return fmt.Sprintf("digraph G {\n\trankdir=LR\n%s\n}\n", strings.Join(trs, "\n"))
}

func dot(s *State, visited map[*State]bool) []string {
	res := []string{}
	if visited[s] {
		return res
	}
	visited[s] = true

	for _, tr := range s.Transitions {
		res = append(res, fmt.Sprintf("\tS%d -> S%d [label=\"%v\"]", s.id, tr.Next.id, tr.Matcher))
		res = append(res, dot(tr.Next, visited)...)
	}
	if s.Terminal {
		res = append(res, fmt.Sprintf("\tS%d [peripheries=2]", s.id))
	}
	return res
}

func (s *State) Parse(args []string) error {
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

func (s *State) apply(args []string, pc matcher.ParseContext) (bool, error) {
	if s.Terminal && len(args) == 0 {
		return true, nil
	}
	sort.Sort(s.Transitions)

	if len(args) > 0 {
		arg := args[0]

		if !pc.RejectOptions && arg == "--" {
			pc.RejectOptions = true
			args = args[1:]
		}
	}

	type match struct {
		tr  *Transition
		rem []string
		pc  matcher.ParseContext
	}

	matches := []*match{}
	for _, tr := range s.Transitions {
		fresh := matcher.NewParseContext()
		fresh.RejectOptions = pc.RejectOptions
		if ok, rem := tr.Matcher.Match(args, &fresh); ok {
			matches = append(matches, &match{tr, rem, fresh})
		}
	}

	for _, m := range matches {
		ok, err := m.tr.Next.apply(m.rem, m.pc)
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
