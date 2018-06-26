package chaos

import (
	"math/rand"
	"time"
)

type delaySpec struct {
	duration    time.Duration
	probability float64
}

func (cs *chaosSpec) injectDelay() bool {
	if cs.ds != nil {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		if p := rnd.Float64(); p > 1-cs.ds.probability {
			time.Sleep(cs.ds.duration)
			return true
		}
	}

	return false
}
