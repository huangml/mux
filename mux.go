package mux

import (
	"regexp"
	"strings"
	"sync"
)

type entry struct {
	val   interface{}
	index int
}

type Config struct {
	TrimPattern TrimFunc
	TrimString  TrimFunc
	Matcher     MatchFunc
}

type Mux struct {
	trimPattern TrimFunc
	trimString  TrimFunc
	matcher     MatchFunc

	m     map[string]*entry
	mtx   sync.RWMutex
	index int
}

func (m *Mux) SetStringTrimmer(f TrimFunc) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.trimString = f
}

func (m *Mux) SetMatcher(f MatchFunc) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.matcher = f
}

func (m *Mux) Map(pattern string, val interface{}) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	pattern = m.trimPattern(pattern)

	if e, ok := m.m[pattern]; ok {
		e.val = val
	} else {
		m.index++
		m.m[pattern] = &entry{
			val:   val,
			index: m.index,
		}
	}
}

func (m *Mux) Delete(pattern string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	delete(m.m, m.trimPattern(pattern))
}

func (m *Mux) Match(s string) (val interface{}) {
	val, _, _ = m.MatchWithPatternScore(s)
	return
}

func (m *Mux) MatchWithPattern(s string) (val interface{}, pattern string) {
	val, pattern, _ = m.MatchWithPatternScore(s)
	return
}

func (m *Mux) MatchWithPatternScore(s string) (val interface{}, pattern string, maxScore int) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	hasOK := false
	s = m.trimString(s)
	for p, e := range m.m {
		if ok, score := m.matcher(p, s, e.index); ok && (!hasOK || score > maxScore) {
			hasOK, maxScore = true, score
			val, pattern = e.val, p
		}
	}
	return
}

func (m *Mux) MatchAll(s string) (vals []interface{}) {
	vals, _, _ = m.MatchAllWithPatternScore(s)
	return
}

func (m *Mux) MatchAllWithPattern(s string) (vals []interface{}, patterns []string) {
	vals, patterns, _ = m.MatchAllWithPatternScore(s)
	return
}

func (m *Mux) MatchAllWithPatternScore(s string) (vals []interface{}, patterns []string, scores []int) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	s = m.trimString(s)
	for p, e := range m.m {
		if ok, score := m.matcher(p, s, e.index); ok {
			vals = append(vals, e.val)
			patterns = append(patterns, p)
			scores = append(scores, score)
		}
	}
	return
}

type TrimFunc func(s string) string

var NoTrim = func(s string) string {
	return s
}

var PathTrim = func(s string) string {
	if s == "" {
		return "/"
	}

	if s[0] != '/' {
		s = "/" + s
	}

	return s
}

type MatchFunc func(pattern, s string, index int) (ok bool, score int)

var StrictMatch = func(pattern, s string, index int) (ok bool, score int) {
	ok = pattern == s
	return
}

var PathMatch = func(pattern, s string, index int) (ok bool, score int) {
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == s, n
	} else {
		return len(s) >= n && s[:n] == pattern, n
	}
}

var PrefixMatch = func(pattern, s string, index int) (ok bool, score int) {
	return strings.HasPrefix(s, pattern), len(pattern)
}

var SuffixMatch = func(pattern, s string, index int) (ok bool, score int) {
	return strings.HasSuffix(s, pattern), len(pattern)
}

var RegexMatch = func(pattern, s string, index int) (ok bool, score int) {
	return regexp.MustCompile(pattern).MatchString(s), index
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

func New(c Config) *Mux {
	if c.TrimPattern == nil {
		c.TrimPattern = NoTrim
	}
	if c.TrimString == nil {
		c.TrimString = NoTrim
	}
	if c.Matcher == nil {
		c.Matcher = StrictMatch
	}

	return &Mux{
		trimPattern: c.TrimPattern,
		trimString:  c.TrimString,
		matcher:     c.Matcher,

		m: make(map[string]*entry),
	}
}

func NewStrictMux() *Mux {
	return New(Config{})
}

func NewPathMux() *Mux {
	return New(Config{
		TrimPattern: PathTrim,
		TrimString:  PathTrim,
		Matcher:     PathMatch,
	})
}
