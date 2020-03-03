package mock

// QuotaStatusHandlerStub -
type QuotaStatusHandlerStub struct {
	ResetStatisticsCalled func()
	AddQuotaCalled        func(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
		numProcessedMessages uint32, sizeProcessedMessages uint64)
	SetGlobalQuotaCalled func(numReceived uint32, sizeReceived uint64, numProcessed uint32, sizeProcessed uint64)
}

// ResetStatistics -
func (qshs *QuotaStatusHandlerStub) ResetStatistics() {
	qshs.ResetStatisticsCalled()
}

// AddQuota -
func (qshs *QuotaStatusHandlerStub) AddQuota(
	identifier string,
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	qshs.AddQuotaCalled(identifier, numReceived, sizeReceived, numProcessed, sizeProcessed)
}

// SetGlobalQuota -
func (qshs *QuotaStatusHandlerStub) SetGlobalQuota(
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	qshs.SetGlobalQuotaCalled(numReceived, sizeReceived, numProcessed, sizeProcessed)
}

// IsInterfaceNil -
func (qshs *QuotaStatusHandlerStub) IsInterfaceNil() bool {
	return qshs == nil
}
