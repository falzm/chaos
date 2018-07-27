package chaos

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type errorSpec struct {
	statusCode  int
	message     string
	probability float64
}

func (s *errorSpec) UnmarshalJSON(data []byte) error {
	spec := struct {
		StatusCode  int     `json:"status_code"`
		Message     string  `json:"message"`
		Probability float64 `json:"p"`
	}{}

	if err := json.Unmarshal(data, &spec); err != nil {
		return err
	}

	s.statusCode = spec.StatusCode
	s.message = spec.Message
	s.probability = spec.Probability

	if s.statusCode < 100 || s.statusCode > 600 {
		return fmt.Errorf("error status code parameter value must be 100 < p < 600 ")
	}

	if s.probability < 0 || s.probability > 1 {
		return fmt.Errorf("probability parameter value must be 0 < p < 1 ")
	}

	return nil
}

func (s *spec) injectError() (bool, int, string) {
	if s.err != nil {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		if p := rnd.Float64(); p > 1-s.err.probability {
			return true, s.err.statusCode, s.err.message
		}
	}

	return false, 0, ""
}
