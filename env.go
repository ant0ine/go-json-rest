package rest

import (
	"sync"
)

// inpired by https://groups.google.com/forum/#!msg/golang-nuts/teSBtPvv1GQ/U12qA9N51uIJ
type env struct {
	envLock sync.Mutex
	envMap  map[*Request]map[string]interface{}
}

func (e *env) setVar(r *Request, key string, value interface{}) {
	e.envLock.Lock()
	defer e.envLock.Unlock()
	if e.envMap == nil {
		e.envMap = make(map[*Request]map[string]interface{})
	}
	if e.envMap[r] == nil {
		e.envMap[r] = make(map[string]interface{})
	}
	e.envMap[r][key] = value
}

func (e *env) getVar(r *Request, key string) interface{} {
	e.envLock.Lock()
	defer e.envLock.Unlock()
	if e.envMap == nil {
		return nil
	}
	if e.envMap[r] == nil {
		return nil
	}
	return e.envMap[r][key]
}

func (e *env) clear(r *Request) {
	e.envLock.Lock()
	defer e.envLock.Unlock()
	delete(e.envMap, r)
}
