package estimator

import "sync/atomic"

type execResSizeLimitChecker struct {
	execResSize    uint32
	numExecRes     uint32
	maxExecResSize uint32
}

func newExecResultSizeLimitChecker(execResSize uint32, maxExecResSize uint32) *execResSizeLimitChecker {
	return &execResSizeLimitChecker{
		execResSize:    execResSize,
		maxExecResSize: maxExecResSize,
	}
}

// AddNumExecRes adds the provided value to numExecRes in a concurrent safe manner
func (ers *execResSizeLimitChecker) AddNumExecRes(numExecRes int) {
	atomic.AddUint32(&ers.numExecRes, uint32(numExecRes))
}

// IsMaxExecResSizeReached returns true if the provided number of execution results exceeds maximum allowed size
func (ers *execResSizeLimitChecker) IsMaxExecResSizeReached(numNewExecRes int) bool {
	totalExecRes := atomic.LoadUint32(&ers.numExecRes) + uint32(numNewExecRes)
	execResSize := ers.execResSize * totalExecRes

	// TODO: evaluate adding an execution results size throttler as for blocks

	return execResSize > ers.maxExecResSize
}

// IsInterfaceNil returns true if there is no value under the interface
func (ers *execResSizeLimitChecker) IsInterfaceNil() bool {
	return ers == nil
}
