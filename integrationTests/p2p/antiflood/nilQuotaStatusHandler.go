package antiflood

type nilQuotaStatusHandler struct {
}

// ResetStatistics is not implemented
func (nqsh *nilQuotaStatusHandler) ResetStatistics() {
}

// AddQuota is not implemented
func (nqsh *nilQuotaStatusHandler) AddQuota(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
	numProcessedMessages uint32, sizeProcessedMessages uint64) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (nqsh *nilQuotaStatusHandler) IsInterfaceNil() bool {
	return nqsh == nil
}
