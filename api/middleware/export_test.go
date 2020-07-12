package middleware

import (
	"sync"
	"sync/atomic"
	"time"
)

func SetMaxSecondsBetweenReset(value int64) {
	atomic.StoreInt64(&maxSecondsBetweenReset, value)
}

func NewTestMiddleware(
	facade ElrondHandler,
	maxConcurrentRequests uint32,
	maxNumRequestsPerAddress uint32,
	resetFailedActionFn func(int64, int64),
) *middleWare {
	mw := &middleWare{
		queue:              make(chan struct{}, maxConcurrentRequests),
		facade:             facade,
		mutRequests:        sync.Mutex{},
		sourceRequests:     make(map[string]uint32),
		maxNumRequests:     maxNumRequestsPerAddress,
		lastResetTimestamp: time.Now().Unix(),
	}
	mw.resetFailedActionFn = resetFailedActionFn

	return mw
}
