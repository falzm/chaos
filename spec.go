package chaos

import (
	"encoding/json"
	"fmt"
	"time"
)

type spec struct {
	delay *delaySpec
	err   *errorSpec

	until time.Time
}

func (s *spec) UnmarshalJSON(data []byte) error {
	chaosSpec := struct {
		Delay    *delaySpec `json:"delay,omitempty"`
		Error    *errorSpec `json:"error,omitempty"`
		Duration string     `json:"duration,omitempty"`
	}{}

	if err := json.Unmarshal(data, &chaosSpec); err != nil {
		return err
	}

	s.delay = chaosSpec.Delay
	s.err = chaosSpec.Error

	if chaosSpec.Duration != "" {
		duration, err := time.ParseDuration(chaosSpec.Duration)
		if err != nil {
			return fmt.Errorf("invalid value for duration parameter: %s", err)
		}

		s.until = time.Now().Add(duration)
	}

	return nil
}
