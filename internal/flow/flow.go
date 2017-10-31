package flow

import (
	"fmt"
	"strings"
)

type ExitCode int

type Step struct {
	Do      func()
	Success *Step
	Error   *Step
	Desc    string
	Exiter  func(code int)
}

func (s *Step) Run(p interface{}) {
	s.callDo(p)

	switch {
	case s.Success != nil:
		s.Success.Run(p)
	case p == nil:
		return
	default:
		if code, ok := p.(ExitCode); ok {
			if s.Exiter != nil {
				s.Exiter(int(code))
			}
			return
		}
		panic(p)
	}
}

func (s *Step) callDo(p interface{}) {
	if s.Do == nil {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			if s.Error == nil {
				panic(p)
			}
			s.Error.Run(e)
		}
	}()
	s.Do()
}

func (s *Step) Dot() string {
	trs := flowDot(s, map[*Step]bool{})
	return fmt.Sprintf("digraph G {\n\trankdir=LR\n%s\n}\n", strings.Join(trs, "\n"))
}

func flowDot(s *Step, visited map[*Step]bool) []string {
	res := []string{}
	if visited[s] {
		return res
	}
	visited[s] = true

	if s.Success != nil {
		res = append(res, fmt.Sprintf("\t\"%s\" -> \"%s\" [label=\"ok\"]", s.Desc, s.Success.Desc))
		res = append(res, flowDot(s.Success, visited)...)
	}
	if s.Error != nil {
		res = append(res, fmt.Sprintf("\t\"%s\" -> \"%s\" [label=\"ko\"]", s.Desc, s.Error.Desc))
		res = append(res, flowDot(s.Error, visited)...)
	}
	return res
}
