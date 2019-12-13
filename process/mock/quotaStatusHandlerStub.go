package mock

type QuotaStatusHandlerStub struct {
	ResetStatisticsCalled func()
	AddQuotaCalled        func(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
		numProcessedMessages uint32, sizeProcessedMessages uint64)
}

func (qshs *QuotaStatusHandlerStub) ResetStatistics() {
	qshs.ResetStatisticsCalled()
}

func (qshs *QuotaStatusHandlerStub) AddQuota(
	identifier string,
	numReceived uint32,
	sizeReceived uint64,
	numProcessed uint32,
	sizeProcessed uint64,
) {
	qshs.AddQuotaCalled(identifier, numReceived, sizeReceived, numProcessed, sizeProcessed)
}

func (qshs *QuotaStatusHandlerStub) IsInterfaceNil() bool {
	return qshs == nil
}
