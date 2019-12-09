package mock

type QuotaStatusHandlerStub struct {
	ResetStatisticsCalled func()
	AddQuotaCalled        func(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
		numProcessedMessages uint32, sizeProcessedMessages uint64)
}

func (qshs *QuotaStatusHandlerStub) ResetStatistics() {
	qshs.ResetStatisticsCalled()
}

func (qshs *QuotaStatusHandlerStub) AddQuota(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64,
	numProcessedMessages uint32, sizeProcessedMessages uint64) {

	qshs.AddQuotaCalled(identifier, numReceivedMessages, sizeReceivedMessages, numProcessedMessages, sizeProcessedMessages)
}

func (qshs *QuotaStatusHandlerStub) IsInterfaceNil() bool {
	return qshs == nil
}
