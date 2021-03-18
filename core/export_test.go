package core

import "time"

func (sw *StopWatch) GetContainingDuration() (map[string]time.Duration, []string) {
	return sw.getContainingDuration()
}

func SplitExponentFraction(val string) (string, string) {
	return splitExponentFraction(val)
}
