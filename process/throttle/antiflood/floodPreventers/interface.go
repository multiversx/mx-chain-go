package floodPreventers

import "github.com/multiversx/mx-chain-core-go/core"

// QuotaStatusHandler defines the behavior of a quota handler able to process periodic updates of peers quota measured
// by the system
type QuotaStatusHandler interface {
	ResetStatistics()
	AddQuota(pid core.PeerID, numReceived uint32, sizeReceived uint64, numProcessed uint32, sizeProcessed uint64)
	IsInterfaceNil() bool
}
