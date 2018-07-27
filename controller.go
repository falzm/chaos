package chaos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type chaosController struct {
	server *http.Server
	routes map[string]*spec

	sync.RWMutex
}

func (c *chaosController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var (
		method string
		path   string
	)

	if method = r.URL.Query().Get("method"); method == "" {
		http.Error(rw, "Missing value for method parameter", http.StatusBadRequest)
		return
	}

	if path = r.URL.Query().Get("path"); path == "" {
		http.Error(rw, "Missing value for path parameter", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		c.getRouteChaosSpec(rw, r, method, path)

	case "PUT":
		c.setRouteChaosSpec(rw, r, method, path)

	case "DELETE":
		c.delRouteChaosSpec(rw, r, method, path)
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (c *chaosController) setRouteChaosSpec(rw http.ResponseWriter, r *http.Request, method, path string) {
	var cs spec

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Invalid request body: %s", err), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(data, &cs); err != nil {
		http.Error(rw, fmt.Sprintf("Invalid request body: %s", err), http.StatusBadRequest)
		return
	}

	c.Lock()
	c.routes[method+path] = &cs
	c.Unlock()

	rw.WriteHeader(http.StatusNoContent)
}

func (c *chaosController) getRouteChaosSpec(rw http.ResponseWriter, r *http.Request, method, path string) {
	c.RLock()
	spec, ok := c.routes[method+path]
	c.RUnlock()
	if !ok {
		http.Error(rw, "No such route", http.StatusNotFound)
		return
	}

	if spec.delay != nil {
		fmt.Fprintf(rw, "Delay: %s (probability: %.1f)\n", spec.delay.duration, spec.delay.probability)
	}

	if spec.err != nil {
		fmt.Fprintf(rw, "Error: %d %q (probability: %.1f)\n",
			spec.err.statusCode, spec.err.message, spec.err.probability)
	}

	if !spec.until.IsZero() {
		fmt.Fprintf(rw, "Until: %s\n", spec.until)
	}
}

func (c *chaosController) delRouteChaosSpec(rw http.ResponseWriter, r *http.Request, method, path string) {
	c.RLock()
	_, ok := c.routes[method+path]
	c.RUnlock()
	if !ok {
		http.Error(rw, "No such endpoint", http.StatusNotFound)
		return
	}

	c.Lock()
	delete(c.routes, method+path)
	c.Unlock()

	rw.WriteHeader(http.StatusNoContent)
}
