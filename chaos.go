package chaos

import (
	"fmt"
	"net/http"
)

// Default network address and port to bind the chaos management HTTP controller to.
const DefaultBindAddr = "127.0.0.1:8666"

type chaosSpec struct {
	ds *delaySpec
	es *errorSpec

	// TODO: implement chaos effect duration
}

// Chaos represents an instance of a Chaos middleware.
type Chaos struct {
	routes map[string]*chaosSpec

	controller *chaosController
}

// NewChaos returns a new Chaos middleware instance, with management HTTP controller listening on bindAddr (fallback to DefaultBindAddr if empty).
func NewChaos(bindAddr string) *Chaos {
	if bindAddr == "" {
		bindAddr = DefaultBindAddr
	}

	c := Chaos{
		routes:     make(map[string]*chaosSpec),
		controller: &chaosController{},
	}

	c.controller.chaos = &c
	c.controller.server = &http.Server{
		Addr:    bindAddr,
		Handler: c.controller,
	}

	go c.controller.server.ListenAndServe()

	return &c
}

// ServeHTTP is the middleware method implementing the Negroni HTTP middleware Handler interface type.
func (c *Chaos) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if cs, ok := c.routes[r.Method+r.URL.Path]; ok {
		if cs.injectDelay() {
			rw.Header().Add("X-Chaos-Injected-Delay", fmt.Sprintf("%s (probability: %.1f)",
				cs.ds.duration, cs.ds.probability))
		}

		if ok, statusCode, msg := cs.injectError(); ok {
			rw.Header().Add("X-Chaos-Injected-Error", fmt.Sprintf("%d (probability: %.1f)",
				cs.es.statusCode, cs.es.probability))
			http.Error(rw, msg, statusCode)
			return
		}
	}

	next(rw, r)
}
