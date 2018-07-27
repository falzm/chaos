package chaos

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Default network address and port to bind the chaos management HTTP controller to.
const DefaultBindAddr = "127.0.0.1:8666"

// Chaos represents an instance of a Chaos middleware.
type Chaos struct {
	controller *chaosController
}

// NewChaos returns a new Chaos middleware instance with management HTTP controller listening on bindAddr
// (fallback to DefaultBindAddr if empty), or a non-nil error if middleware initialization failed. If bindAddr starts
// with "unix:", the controller will be bound to a UNIX socket at the path described after the "unix:" prefix (e.g.
// "unix:/var/run/http-chaos.sock").
func NewChaos(bindAddr string) (*Chaos, error) {
	var (
		c = Chaos{
			controller: &chaosController{
				routes: make(map[string]*spec),
			},
		}
		listener net.Listener
		err      error
	)

	if bindAddr == "" {
		bindAddr = DefaultBindAddr
	}

	c.controller.server = &http.Server{Handler: c.controller}

	if strings.HasPrefix(bindAddr, "unix:") {
		if listener, err = net.Listen("unix", strings.TrimPrefix(bindAddr, "unix:")); err != nil {
			return nil, fmt.Errorf("unable to bind UNIX socket: %s", err)
		}
	} else {
		if listener, err = net.Listen("tcp", bindAddr); err != nil {
			return nil, fmt.Errorf("unable to bind TCP socket: %s", err)
		}
	}

	go c.controller.server.Serve(listener)

	return &c, nil
}

// Handler is the middleware method implementing the standard net/http Handler interface type.
func (c *Chaos) Handler(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if c.inject(rw, r) {
			h.ServeHTTP(rw, r)
		}
	})
}

// ServeHTTP is the middleware method implementing the Negroni HTTP middleware Handler interface type.
func (c *Chaos) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if c.inject(rw, r) {
		next(rw, r)
	}
}

// inject is the actual chaos injection code, it returns a booleaon value false to signal the calling handler that it
// must not continue the middleware chain if an injected error interrupted the request processing.
func (c *Chaos) inject(rw http.ResponseWriter, r *http.Request) (cont bool) {
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
			return false
		}
	}

	return true
}
