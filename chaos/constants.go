package chaos

const (
	roundDivisor_maybeCorruptSignature                      = 5
	roundDivisor_shouldSkipWaitingForSignatures             = 7
	roundDivisor_shouldReturnErrorInCheckSignaturesValidity = 11
	blockNonceDivisor_shouldPanic                           = 123
	numCallsDivisor_processTransaction_shouldReturnError    = 30_001
)
