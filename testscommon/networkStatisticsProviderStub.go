package testscommon

// NetworkStatisticsProviderStub -
type NetworkStatisticsProviderStub struct {
	BpsSentCalled                          func() uint64
	BpsRecvCalled                          func() uint64
	BpsSentPeakCalled                      func() uint64
	BpsRecvPeakCalled                      func() uint64
	PercentSentCalled                      func() uint64
	PercentRecvCalled                      func() uint64
	TotalBytesSentInCurrentEpochCalled     func() uint64
	TotalBytesReceivedInCurrentEpochCalled func() uint64
	TotalSentInCurrentEpochCalled          func() string
	TotalReceivedInCurrentEpochCalled      func() string
	EpochConfirmedCalled                   func(epoch uint32, timestamp uint64)
	CloseCalled                            func() error
}

// BpsSent -
func (stub *NetworkStatisticsProviderStub) BpsSent() uint64 {
	if stub.BpsSentCalled != nil {
		return stub.BpsSentCalled()
	}
	return 0
}

// BpsRecv -
func (stub *NetworkStatisticsProviderStub) BpsRecv() uint64 {
	if stub.BpsRecvCalled != nil {
		return stub.BpsRecvCalled()
	}
	return 0
}

// BpsSentPeak -
func (stub *NetworkStatisticsProviderStub) BpsSentPeak() uint64 {
	if stub.BpsSentPeakCalled != nil {
		return stub.BpsSentPeakCalled()
	}
	return 0
}

// BpsRecvPeak -
func (stub *NetworkStatisticsProviderStub) BpsRecvPeak() uint64 {
	if stub.BpsRecvPeakCalled != nil {
		return stub.BpsRecvPeakCalled()
	}
	return 0
}

// PercentSent -
func (stub *NetworkStatisticsProviderStub) PercentSent() uint64 {
	if stub.PercentSentCalled != nil {
		return stub.PercentSentCalled()
	}
	return 0
}

// PercentRecv -
func (stub *NetworkStatisticsProviderStub) PercentRecv() uint64 {
	if stub.PercentRecvCalled != nil {
		return stub.PercentRecvCalled()
	}
	return 0
}

// TotalBytesSentInCurrentEpoch -
func (stub *NetworkStatisticsProviderStub) TotalBytesSentInCurrentEpoch() uint64 {
	if stub.TotalBytesSentInCurrentEpochCalled != nil {
		return stub.TotalBytesSentInCurrentEpochCalled()
	}
	return 0
}

// TotalBytesReceivedInCurrentEpoch -
func (stub *NetworkStatisticsProviderStub) TotalBytesReceivedInCurrentEpoch() uint64 {
	if stub.TotalBytesReceivedInCurrentEpochCalled != nil {
		return stub.TotalBytesReceivedInCurrentEpochCalled()
	}
	return 0
}

// TotalSentInCurrentEpoch -
func (stub *NetworkStatisticsProviderStub) TotalSentInCurrentEpoch() string {
	if stub.TotalSentInCurrentEpochCalled != nil {
		return stub.TotalSentInCurrentEpochCalled()
	}
	return ""
}

// TotalReceivedInCurrentEpoch -
func (stub *NetworkStatisticsProviderStub) TotalReceivedInCurrentEpoch() string {
	if stub.TotalReceivedInCurrentEpochCalled != nil {
		return stub.TotalReceivedInCurrentEpochCalled()
	}
	return ""
}

// EpochConfirmed -
func (stub *NetworkStatisticsProviderStub) EpochConfirmed(epoch uint32, timestamp uint64) {
	if stub.EpochConfirmedCalled != nil {
		stub.EpochConfirmedCalled(epoch, timestamp)
	}
}

// Close -
func (stub *NetworkStatisticsProviderStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *NetworkStatisticsProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
