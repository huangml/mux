package mux

import (
	"strings"
	"sync"
)

type entry struct {
	val   interface{}
	index int
}

type Mux struct {
	PatternTrimer TrimFunc
	StringTrimer  TrimFunc
	Matcher       MatchFunc

	m   map[string]*entry
	mtx sync.RWMutex
}

func New() *Mux {
	return &Mux{
		PatternTrimer: NoTrim,
		StringTrimer:  NoTrim,
		Matcher:       StrictMatch,

		m: make(map[string]*entry),
	}
}

func (m *Mux) Bind(pattern string, val interface{}) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	pattern = m.PatternTrimer(pattern)

	if e, ok := m.m[pattern]; ok {
		e.val = val
	} else {
		m.m[pattern] = &entry{
			val:   val,
			index: len(m.m),
		}
	}
}

func (m *Mux) Match(s string) (val interface{}, pattern string) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	var (
		hasOK    bool
		maxScore int
	)

	s = m.StringTrimer(s)
	for p, e := range m.m {
		if ok, score := m.Matcher(p, s, e.index); ok && (!hasOK || score > maxScore) {
			hasOK, maxScore = true, score
			val, pattern = e.val, p
		}
	}
	return
}

type TrimFunc func(s string) string

func NoTrim(s string) string {
	return s
}

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
