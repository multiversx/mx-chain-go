package antiflood

// QuotaStatusHandler defines the behavior of a quota handler able to process periodic updates of peers quota measured
// by the system
type QuotaStatusHandler interface {
	ResetStatistics()
	AddQuota(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
		numProcessedMessages uint32, sizeProcessedMessages uint64)
	IsInterfaceNil() bool
}
