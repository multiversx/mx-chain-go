package floodPreventers

// QuotaStatusHandler defines the behavior of a quota handler able to process periodic updates of peers quota measured
// by the system
type QuotaStatusHandler interface {
	ResetStatistics()
	AddQuota(identifier string, numReceived uint32, sizeReceived uint64, numProcessed uint32, sizeProcessed uint64)
	SetGlobalQuota(numReceived uint32, sizeReceived uint64, numProcessed uint32, sizeProcessed uint64)
	IsInterfaceNil() bool
}
