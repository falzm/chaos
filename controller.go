package chaos

import (
	"fmt"
	"net/http"
	"time"

	"github.com/facette/httputil"
)

type chaosController struct {
	server *http.Server
	chaos  *Chaos
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
	var (
		cs   chaosSpec
		data struct {
			Delay *struct {
				Duration    int     `json:"duration"`
				Probability float64 `json:"p"`
			} `json:"delay,omitempty"`
			Error *struct {
				StatusCode  int     `json:"status_code"`
				Message     string  `json:"message"`
				Probability float64 `json:"p"`
			} `json:"error,omitempty"`
		}
	)

	if err := httputil.BindJSON(r, &data); err != nil {
		http.Error(rw, fmt.Sprintf("Invalid request body: %s", err), http.StatusBadRequest)
		return
	}

	if data.Error != nil {
		cs.es = &errorSpec{
			statusCode:  data.Error.StatusCode,
			message:     data.Error.Message,
			probability: data.Error.Probability,
		}

		if cs.es.statusCode < 100 || cs.es.statusCode > 600 {
			http.Error(rw, "Error status code parameter value must be 100 < p < 600 ", http.StatusBadRequest)
			return
		}
		if cs.es.probability < 0 || cs.es.probability > 1 {
			http.Error(rw, "Probability parameter value must be 0 < p < 1 ", http.StatusBadRequest)
			return
		}
	}

	if data.Delay != nil {
		cs.ds = &delaySpec{
			duration:    time.Duration(data.Delay.Duration) * time.Millisecond,
			probability: data.Delay.Probability,
		}

		if data.Delay.Duration <= 0 {
			http.Error(rw, "Delay duration parameter value must be greater than 0 ", http.StatusBadRequest)
			return
		}
		if cs.ds.probability < 0 || cs.ds.probability > 1 {
			http.Error(rw, "Probability parameter value must be 0 < p < 1 ", http.StatusBadRequest)
			return
		}
	}

	c.chaos.routes[method+path] = &cs

	rw.WriteHeader(http.StatusNoContent)
}

func (c *chaosController) getRouteChaosSpec(rw http.ResponseWriter, r *http.Request, method, path string) {
	cs, ok := c.chaos.routes[method+path]
	if !ok {
		http.Error(rw, "No such route", http.StatusNotFound)
		return
	}

	if cs.ds != nil {
		fmt.Fprintf(rw, "Delay: %s (probability: %.1f)\n", cs.ds.duration, cs.ds.probability)
	}

	if cs.es != nil {
		fmt.Fprintf(rw, "Error: %d %q (probability: %.1f)\n",
			cs.es.statusCode, cs.es.message, cs.es.probability)
	}
}

func (c *chaosController) delRouteChaosSpec(rw http.ResponseWriter, r *http.Request, method, path string) {
	if _, ok := c.chaos.routes[method+path]; !ok {
		http.Error(rw, "No such endpoint", http.StatusNotFound)
		return
	}

	delete(c.chaos.routes, method+path)

	rw.WriteHeader(http.StatusNoContent)
}