package rest

import (
	"net/http"
	"sync"
)

// inpired by https://groups.google.com/forum/#!msg/golang-nuts/teSBtPvv1GQ/U12qA9N51uIJ
type env struct {
	envLock sync.Mutex
	envMap  map[*http.Request]map[string]interface{}
}

func (self *env) setVar(r *http.Request, key string, value interface{}) {
	self.envLock.Lock()
	defer self.envLock.Unlock()
	if self.envMap == nil {
		self.envMap = make(map[*http.Request]map[string]interface{})
	}
	if self.envMap[r] == nil {
		self.envMap[r] = make(map[string]interface{})
	}
	self.envMap[r][key] = value
}

func (self *env) getVar(r *http.Request, key string) interface{} {
	self.envLock.Lock()
	defer self.envLock.Unlock()
	if self.envMap == nil {
		return nil
	}
	if self.envMap[r] == nil {
		return nil
	}
	return self.envMap[r][key]
}

func (self *env) clear(r *http.Request) {
	self.envLock.Lock()
	defer self.envLock.Unlock()
	delete(self.envMap, r)
}
