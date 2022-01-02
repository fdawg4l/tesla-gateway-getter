package gateway

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Reflector reflects local objects and provides an http server handler
type Reflector struct {
	mu *sync.RWMutex

	objs []interface{}
}

func NewReflector() *Reflector {
	return &Reflector{
		mu: new(sync.RWMutex),
	}
}

func (r *Reflector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	enc := json.NewEncoder(w)

	defer r.mu.RUnlock()

	if err := enc.Encode(r.objs); err != nil {
		enc.Encode(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (r *Reflector) Reflect(objs ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.objs = objs
}
