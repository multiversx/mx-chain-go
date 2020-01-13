package core

import "time"

func (sw *stopWatch) GetContainingDuration() (map[string]time.Duration, []string) {
	return sw.getContainingDuration()
}
