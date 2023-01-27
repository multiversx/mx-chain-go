package mock

import "github.com/multiversx/mx-chain-core-go/core"

// QuotaStatusHandlerStub -
type QuotaStatusHandlerStub struct {
	ResetStatisticsCalled func()
	AddQuotaCalled        func(pid core.PeerID, numReceivedMessages uint32, sizeReceivedMessages uint64,
		numProcessedMessages uint32, sizeProcessedMessages uint64)
}

// ResetStatistics -
func (qshs *QuotaStatusHandlerStub) ResetStatistics() {
	if qshs.ResetStatisticsCalled != nil {
		qshs.ResetStatisticsCalled()
	}
}

// AddQuota -
func (qshs *QuotaStatusHandlerStub) AddQuota(
	pid core.PeerID,
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	if qshs.AddQuotaCalled != nil {
		qshs.AddQuotaCalled(pid, numReceived, sizeReceived, numProcessed, sizeProcessed)
	}
}

// IsInterfaceNil -
func (qshs *QuotaStatusHandlerStub) IsInterfaceNil() bool {
	return qshs == nil
}
