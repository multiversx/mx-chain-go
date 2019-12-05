package txcache

import (
	"fmt"
	"time"
)

type stopWatch struct {
	milliseconds int64
	checkpoint   time.Time
}

func (watch *stopWatch) start() {
	watch.checkpoint = time.Now()
}

func (watch *stopWatch) pause() {
	sinceCheckpoint := time.Now().Sub(watch.checkpoint)
	watch.milliseconds += sinceCheckpoint.Milliseconds()
}

func (watch *stopWatch) format() string {
	return fmt.Sprintf("%f", float32(watch.milliseconds)/1000.0)
}
