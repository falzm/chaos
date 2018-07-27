package chaos

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type delaySpec struct {
	duration    time.Duration
	probability float64
}

func (s *delaySpec) UnmarshalJSON(data []byte) error {
	delaySpec := struct {
		Duration    int     `json:"duration"`
		Probability float64 `json:"p"`
	}{}

	if err := json.Unmarshal(data, &delaySpec); err != nil {
		return err
	}

	s.duration = time.Duration(delaySpec.Duration) * time.Millisecond
	s.probability = delaySpec.Probability

	if delaySpec.Duration <= 0 {
		return fmt.Errorf("delay duration parameter value must be greater than 0 ")
	}

	if s.probability < 0 || s.probability > 1 {
		return fmt.Errorf("probability parameter value must be 0 < p < 1 ")
	}

	return nil
}

func (s *spec) injectDelay() bool {
	if s.delay != nil {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		if p := rnd.Float64(); p > 1-s.delay.probability {
			time.Sleep(s.delay.duration)
			return true
		}
	}

	return false
}
