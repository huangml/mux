package mux

import "sync"

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
	PatternTrimFn TrimFunc
	TrimFn        TrimFunc
	MatchFn       MatchFunc

	rawMux
}

func New() *Mux {
	return &Mux{
		rawMux: rawMux{
			m: make(map[string]*entry),
		},
	}
}

func (m *Mux) Bind(pattern string, val interface{}) {
	if m.PatternTrimFn != nil {
		pattern = m.PatternTrimFn(pattern)
	}
	m.rawMux.bind(pattern, val)
}

func (m *Mux) Match(s string) (val interface{}, pattern string) {
	if m.TrimFn != nil {
		s = m.TrimFn(s)
	}

	if m.MatchFn != nil {
		return m.rawMux.match(s, m.MatchFn)
	} else {
		return m.rawMux.match(s, StrictMatchFn)
	}
}

type TrimFunc func(s string) string

func PathTrimFn(s string) string {
	if s == "" {
		return "/"
	}

	if s[0] != '/' {
		s = "/" + s
	}

	return s
}

type MatchFunc func(pattern, s string, index int) (ok bool, score int)

func StrictMatchFn(pattern, s string, index int) (ok bool, score int) {
	ok = pattern == s
	return
}

func PathMatchFn(pattern, s string, index int) (ok bool, score int) {
	n := len(pattern)
	if pattern[n-1] != '/' {
		ok = pattern == s
	} else {
		ok = len(s) >= n && s[:n] == pattern
	}

	score = n

	return
}

func FirstMatch(f MatchFunc) MatchFunc {
	return func(pattern, s string, index int) (ok bool, score int) {
		ok, _ = f(pattern, s, index)
		score = -index
		return
	}
}

func LastMatch(f MatchFunc) MatchFunc {
	return func(pattern, s string, index int) (ok bool, score int) {
		ok, _ = f(pattern, s, index)
		score = index
		return
	}
}

var PathMux = New()

func init() {
	PathMux.PatternTrimFn = PathTrimFn
	PathMux.TrimFn = PathTrimFn
	PathMux.MatchFn = PathMatchFn
}
