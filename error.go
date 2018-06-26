package chaos

import (
	"math/rand"
	"time"
)

type errorSpec struct {
	statusCode  int
	message     string
	probability float64
}

func (cs *chaosSpec) injectError() (bool, int, string) {
	if cs.es != nil {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		if p := rnd.Float64(); p > 1-cs.es.probability {
			return true, cs.es.statusCode, cs.es.message
		}
	}

	return false, 0, ""
}
