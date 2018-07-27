package chaos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// Client represents a chaos controller management client.
type Client struct {
	http http.Client
}

// NewClient returns a client for managing chaos controller on address controllerAddr.
func NewClient(controllerAddr string) *Client {
	var (
		client          Client
		controllerProto = "tcp"
	)

	if controllerAddr == "" {
		controllerAddr = DefaultBindAddr
	}

	if strings.HasPrefix(controllerAddr, "unix:") {
		controllerProto = "unix"
		controllerAddr = strings.TrimPrefix(controllerAddr, "unix:")
	}

	client.http = http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(controllerProto, controllerAddr)
			},
		},
	}

	return &client
}

// Spec represents a chaos route specification.
type Spec struct {
	s map[string]interface{}
}

// NewSpec returns an empty chaos route specification.
func NewSpec() *Spec {
	return &Spec{s: make(map[string]interface{})}
}

// Delay sets a chaos delay injection of d milliseconds at a p probability (0 < p < 1) to chaos spec.
func (s *Spec) Delay(d int, p float64) *Spec {
	s.s["delay"] = map[string]interface{}{
		"duration": d,
		"p":        p,
	}

	return s
}

// Delay sets a chaos error injection with HTTP status code sc with an optional message msg at a p probability
// (0 < p < 1) to chaos spec.
func (s *Spec) Error(sc int, msg string, p float64) *Spec {
	s.s["error"] = map[string]interface{}{
		"status_code": sc,
		"message":     msg,
		"p":           p,
	}

	return s
}

// During specifies that the route chaos spec effects must be enforced for a duration d
// (value must be expressed using time.ParseDuration() format).
func (s *Spec) During(d string) *Spec {
	s.s["duration"] = d

	return s
}

// AddRouteChaos adds chaos effects specified by spec to the route defined by method method (e.g. "POST")
// and URL path path (e.g. "/api/foo"), and returns an error if it failed.
func (c *Client) AddRouteChaos(method, path string, spec *Spec) error {
	js, err := json.Marshal(spec.s)
	if err != nil {
		return fmt.Errorf("unable to marshal spec to JSON: %s", err)
	}

	req, err := http.NewRequest("PUT",
		fmt.Sprintf("http://controller/?method=%s&path=%s", method, path),
		bytes.NewBuffer(js))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s\n", err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %s\n", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %s", err)
	}

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("controller error: %s: %s", res.Status, body)
	}

	return nil
}

// DeleteRouteChaos delete route chaos specification applied to the route defined by method method (e.g. "POST")
// and URL path path (e.g. "/api/foo"), and returns an error if it failed.
func (c *Client) DeleteRouteChaos(method, path string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://controller/?method=%s&path=%s", method, path), nil)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s\n", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %s\n", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %s", err)
	}

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("controller error: %s: %s", res.Status, body)
	}

	return nil
}
