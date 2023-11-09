package antiflood

import "github.com/multiversx/mx-chain-core-go/core"

type nilQuotaStatusHandler struct {
}

// ResetStatistics is not implemented
func (nqsh *nilQuotaStatusHandler) ResetStatistics() {
}

// AddQuota is not implemented
func (nqsh *nilQuotaStatusHandler) AddQuota(_ core.PeerID, _ uint32, _ uint64, _ uint32, _ uint64) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (nqsh *nilQuotaStatusHandler) IsInterfaceNil() bool {
	return nqsh == nil
}
