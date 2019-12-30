package p2pQuota

type p2pQuotaBlacklistProcessor struct {
}

func (pqbp *p2pQuotaBlacklistProcessor) ResetStatistics() {
	panic("implement me")
}

func (pqbp *p2pQuotaBlacklistProcessor) AddQuota(identifier string, numReceived uint32, sizeReceived uint64, numProcessed uint32, sizeProcessed uint64) {
	panic("implement me")
}

// SetGlobalQuota does nothing (here to comply with QuotaStatusHandler interface)
func (pqbp *p2pQuotaBlacklistProcessor) SetGlobalQuota(_ uint32, _ uint64, _ uint32, _ uint64) {}

// IsInterfaceNil returns true if there is no value under the interface
func (pqbp *p2pQuotaBlacklistProcessor) IsInterfaceNil() bool {
	return pqbp == nil
}
