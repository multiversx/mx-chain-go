package accumulator

import "time"

const MinimumAllowedTime = minimumAllowedTime

func (ta *timeAccumulator) Data() []interface{} {
	ta.mut.Lock()
	data := make([]interface{}, len(ta.data))
	ta.mut.Unlock()

	return data
}

func (ta *timeAccumulator) ComputeWaitTime() time.Duration {
	return ta.computeWaitTime()
}
