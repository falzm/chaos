package chaos

import (
	"fmt"
	"net/http"
	"time"
)

// Default network address and port to bind the chaos management HTTP controller to.
const DefaultBindAddr = "127.0.0.1:8666"

// Chaos represents an instance of a Chaos middleware.
type Chaos struct {
	controller *chaosController
}

// NewChaos returns a new Chaos middleware instance, with management HTTP controller listening on bindAddr
// (fallback to DefaultBindAddr if empty).
func NewChaos(bindAddr string) *Chaos {
	if bindAddr == "" {
		bindAddr = DefaultBindAddr
	}

	c := Chaos{
		controller: &chaosController{
			routes: make(map[string]*spec),
		},
	}

	c.controller.server = &http.Server{
		Addr:    bindAddr,
		Handler: c.controller,
	}

	go c.controller.server.ListenAndServe()

	return &c
}

// Handler is the middleware method implementing the standard net/http Handler interface type.
func (c *Chaos) Handler(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c.controller.RLock()
		spec, ok := c.controller.routes[r.Method+r.URL.Path]
		c.controller.RUnlock()

		if ok && (spec.until.IsZero() || time.Now().Before(spec.until)) {
			if spec.injectDelay() {
				rw.Header().Add("X-Chaos-Injected-Delay", fmt.Sprintf("%s (probability: %.1f)",
					spec.delay.duration, spec.delay.probability))
			}

			if ok, statusCode, msg := spec.injectError(); ok {
				rw.Header().Add("X-Chaos-Injected-Error", fmt.Sprintf("%d (probability: %.1f)",
					spec.err.statusCode, spec.err.probability))
				http.Error(rw, msg, statusCode)
				return
			}
		}

		h.ServeHTTP(rw, r)
	})
}

// ServeHTTP is the middleware method implementing the Negroni HTTP middleware Handler interface type.
func (c *Chaos) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	c.controller.RLock()
	spec, ok := c.controller.routes[r.Method+r.URL.Path]
	c.controller.RUnlock()

	if ok && (spec.until.IsZero() || time.Now().Before(spec.until)) {
		if spec.injectDelay() {
			rw.Header().Add("X-Chaos-Injected-Delay", fmt.Sprintf("%s (probability: %.1f)",
				spec.delay.duration, spec.delay.probability))
		}

		if ok, statusCode, msg := spec.injectError(); ok {
			rw.Header().Add("X-Chaos-Injected-Error", fmt.Sprintf("%d (probability: %.1f)",
				spec.err.statusCode, spec.err.probability))
			http.Error(rw, msg, statusCode)
			return
		}
	}

	next(rw, r)
}
