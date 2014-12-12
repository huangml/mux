package mux

import (
	"strings"
	"sync"
)

type entry struct {
	val   interface{}
	index int
}

type rawMux struct {
	sync.RWMutex
	m map[string]*entry
}

func (m *rawMux) bind(pattern string, val interface{}) {
	m.Lock()
	defer m.Unlock()

	if e, ok := m.m[pattern]; ok {
		e.val = val
	} else {
		m.m[pattern] = &entry{
			val:   val,
			index: len(m.m),
		}
	}
}

func (m *rawMux) match(s string, f MatchFunc) (val interface{}, pattern string) {
	m.RLock()
	defer m.RUnlock()

	var (
		hasOK    bool
		maxScore int
	)

	for p, e := range m.m {
		if ok, score := f(p, s, e.index); ok {
			if !hasOK || score > maxScore {
				hasOK, maxScore = true, score
				val, pattern = e.val, p
			}
		}
	}
	return
}

type Mux struct {
	PatternTrimer TrimFunc
	StringTrimer  TrimFunc
	Matcher       MatchFunc

	raw rawMux
}

func New() *Mux {
	return &Mux{
		raw: rawMux{
			m: make(map[string]*entry),
		},
	}
}

func (m *Mux) Bind(pattern string, val interface{}) {
	if m.PatternTrimer != nil {
		pattern = m.PatternTrimer(pattern)
	}
	m.raw.bind(pattern, val)
}

func (m *Mux) Match(s string) (val interface{}, pattern string) {
	if m.StringTrimer != nil {
		s = m.StringTrimer(s)
	}

	if m.Matcher != nil {
		return m.raw.match(s, m.Matcher)
	} else {
		return m.raw.match(s, StrictMatch)
	}
}

type TrimFunc func(s string) string

func PathTrim(s string) string {
	if s == "" {
		return "/"
	}

	if s[0] != '/' {
		s = "/" + s
	}

	return s
}

type MatchFunc func(pattern, s string, index int) (ok bool, score int)

func StrictMatch(pattern, s string, index int) (ok bool, score int) {
	ok = pattern == s
	return
}

func PathMatch(pattern, s string, index int) (ok bool, score int) {
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == s, n
	} else {
		return len(s) >= n && s[:n] == pattern, n
	}
}

func PrefixMatch(pattern, s string, index int) (ok bool, score int) {
	return strings.HasPrefix(s, pattern), len(pattern)
}

func SuffixMatch(pattern, s string, index int) (ok bool, score int) {
	return strings.HasSuffix(s, pattern), len(pattern)
}

func CombineTrimFn(f1, f2 TrimFunc) TrimFunc {
	return func(s string) string {
		return f1(f2(s))
	}
}

func FirstMatchFn(f MatchFunc) MatchFunc {
	return func(pattern, s string, index int) (ok bool, score int) {
		ok, _ = f(pattern, s, index)
		score = -index
		return
	}
}

func LastMatchFn(f MatchFunc) MatchFunc {
	return func(pattern, s string, index int) (ok bool, score int) {
		ok, _ = f(pattern, s, index)
		score = index
		return
	}
}

func ShortestPatternMatchFn(f MatchFunc) MatchFunc {
	return func(pattern, s string, index int) (ok bool, score int) {
		ok, _ = f(pattern, s, index)
		score = -len(pattern)
		return
	}
}

func LongestPatternMatchFn(f MatchFunc) MatchFunc {
	return func(pattern, s string, index int) (ok bool, score int) {
		ok, _ = f(pattern, s, index)
		score = len(pattern)
		return
	}
}

func NewPathMux() *Mux {
	m := New()
	m.PatternTrimer = PathTrim
	m.StringTrimer = PathTrim
	m.Matcher = PathMatch
	return m
}
