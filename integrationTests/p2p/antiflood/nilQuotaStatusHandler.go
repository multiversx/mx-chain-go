package antiflood

type nilQuotaStatusHandler struct {
}

// ResetStatistics is not implemented
func (nqsh *nilQuotaStatusHandler) ResetStatistics() {
}

// AddQuota is not implemented
func (nqsh *nilQuotaStatusHandler) AddQuota(_ string, _ uint32, _ uint64, _ uint32, _ uint64) {
}

// SetGlobalQuota is not implemented
func (nqsh *nilQuotaStatusHandler) SetGlobalQuota(_ uint32, _ uint64, _ uint32, _ uint64) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (nqsh *nilQuotaStatusHandler) IsInterfaceNil() bool {
	return nqsh == nil
}
