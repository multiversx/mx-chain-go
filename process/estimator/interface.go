package estimator

// ExecResSizeComputationHandler defines the behaviour of a component that is able to spin new encapsulated execution results size limit checker
type ExecResSizeComputationHandler interface {
	NewComputation() ExecResSizeLimitCheckerHandler
	IsInterfaceNil() bool
}

// ExecResSizeLimitCheckerHandler defines the behaviour of a size limit checker
type ExecResSizeLimitCheckerHandler interface {
	AddNumExecRes(numExecRes int)
	IsMaxExecResSizeReached(numNewExecRes int) bool
	IsInterfaceNil() bool
}
