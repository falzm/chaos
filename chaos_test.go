package chaos

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/falzm/chaos"
	"github.com/gorilla/mux"
)

type testClient struct {
	chaos *Client
	api   *Client
}

func (c *testClient) sendRequest(method, path string) (int, string, float64, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("http://test%s", path), nil)
	if err != nil {
		return 0, "", 0.0, fmt.Errorf("error creating HTTP request: %s\n", err)
	}

	startTime := time.Now()

	res, err := c.api.http.Do(req)
	if err != nil {
		return 0, "", 0.0, fmt.Errorf("error sending HTTP request: %s\n", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, "", 0.0, fmt.Errorf("unable to read response body: %s", err)
	}
	body = bytes.TrimSpace(body)

	return res.StatusCode, string(body), time.Now().Sub(startTime).Seconds() * 1000, nil
}

func (c *testClient) testRouteChaos(method, path string, spec *Spec, testfunc func() error) error {
	if err := c.chaos.AddRouteChaos(method, path, spec); err != nil {
		return fmt.Errorf("unable to add route chaos spec: %s", err)
	}

	if err := testfunc(); err != nil {
		return err
	}

	if err := c.chaos.DeleteRouteChaos(method, path); err != nil {
		return fmt.Errorf("unable to delete chaos spec from route: %s", err)
	}

	return nil
}

func Test_Chaos(t *testing.T) {
	var testcli testClient

	tmpDir, err := ioutil.TempDir(os.TempDir(), "chaos")
	if err != nil {
		t.Errorf("unable to create temporary test directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	apiSock := path.Join(tmpDir, "api.sock")
	chaosSock := path.Join(tmpDir, "chaos.sock")

	chaos, err := chaos.NewChaos(fmt.Sprintf("unix:%s", chaosSock))
	if err != nil {
		t.Errorf("unable to bind chaos management controller UNIX socket: %s", err)
	}

	listener, err := net.Listen("unix", apiSock)
	if err != nil {
		t.Errorf("unable to bind UNIX socket: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/{action}",
		func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "ohai!")
		})

	server := http.Server{
		Handler: chaos.Handler(router.ServeHTTP),
	}

	t.Logf("startup up test server (listening on %s)\n", apiSock)

	go server.Serve(listener)

	testcli.api = NewClient("unix:" + apiSock)
	testcli.chaos = NewClient("unix:" + chaosSock)

	// Test adding route chaos with invalid error spec
	if err := testcli.testRouteChaos("POST", "/api/a", NewSpec().Error(0, "", 1.0), nil); err == nil {
		t.Errorf("route chaos test failed: %s", err)
		t.FailNow()
	}

	// Test adding route chaos with invalid delay spec
	if err := testcli.testRouteChaos("POST", "/api/a", NewSpec().Delay(0, 1.0), nil); err == nil {
		t.Errorf("route chaos test failed: %s", err)
		t.FailNow()
	}

	// Test chaos delay/error injection
	if err := testcli.testRouteChaos("POST", "/api/a", NewSpec().
		Delay(3000, 1.0).
		Error(http.StatusGatewayTimeout, "Whoopsie...", 1.0),
		func() error {
			var (
				expectedStatusCode         = http.StatusGatewayTimeout
				expectedErrMessage         = "Whoopsie..."
				expectedMinDelay   float64 = 3000
			)

			status, errMessage, latency, err := testcli.sendRequest("POST", "/api/a")
			if err != nil {
				return fmt.Errorf("error sending HTTP request: %s\n", err)
			}

			if latency <= expectedMinDelay {
				return fmt.Errorf("expected minimum request delay > %.2fms but took %.2fms", expectedMinDelay, latency)
			}

			if status != expectedStatusCode {
				return fmt.Errorf("expected status code %d but got %d", expectedStatusCode, status)
			}

			if errMessage != expectedErrMessage {
				return fmt.Errorf("expected error message %q but got %q", expectedErrMessage, errMessage)
			}

			return nil
		}); err != nil {
		t.Errorf("route chaos test failed: %s", err)
		t.FailNow()
	}

	// Test chaos effect temporary duration
	if err := testcli.testRouteChaos("PUT", "/api/b", NewSpec().
		Error(http.StatusTooManyRequests, "", 1.0).
		During("3s"),
		func() error {
			var expectedStatusCode = http.StatusTooManyRequests

			status, _, _, err := testcli.sendRequest("PUT", "/api/b")
			if err != nil {
				return fmt.Errorf("error sending HTTP request: %s\n", err)
			}

			if status != expectedStatusCode {
				return fmt.Errorf("expected status code %d but got %d", expectedStatusCode, status)
			}

			time.Sleep(5 * time.Second)
			expectedStatusCode = http.StatusOK

			status, _, _, err = testcli.sendRequest("PUT", "/api/b")
			if err != nil {
				return fmt.Errorf("error sending HTTP request: %s\n", err)
			}

			if status != expectedStatusCode {
				return fmt.Errorf("expected status code %d but got %d", expectedStatusCode, status)
			}

			return nil
		}); err != nil {
		t.Errorf("route chaos test failed: %s", err)
		t.FailNow()
	}

	t.Log("shutting down test server")

	server.Shutdown(nil)
}
