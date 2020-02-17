package core

import "time"

func (sw *StopWatch) GetContainingDuration() (map[string]time.Duration, []string) {
	return sw.getContainingDuration()
}
