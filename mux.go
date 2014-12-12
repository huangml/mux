package pathmux

import "sync"

type Mux struct {
	mtx sync.RWMutex
	m   map[string]interface{}
}

func New() *Mux {
	return &Mux{
		m: make(map[string]interface{}),
	}
}

func (m *Mux) Bind(pattern string, val interface{}) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.m[clean(pattern)] = val
}

func (m *Mux) Match(path string) (val interface{}, pattern string) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	path = clean(path)

	var n int
	for k, v := range m.m {
		if pathMatch(k, path) {
			if val == nil || len(k) > n {
				n = len(k)
				val = v
				pattern = k
			}
		}
	}
	return
}

func pathMatch(pattern, path string) bool {
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[:n] == pattern
}

func clean(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return p
}
