package consensus

// NtpSyncControllerMock -
type NtpSyncControllerMock struct {
	AddOutOfRangeNonceCalled         func(round uint64, hash string)
	AddLeaderNonceAsOutOfRangeCalled func(round uint64, hash string)
}

// AddOutOfRangeNonce -
func (mock *NtpSyncControllerMock) AddOutOfRangeNonce(nonce uint64, hash string) {
	if mock.AddOutOfRangeNonceCalled != nil {
		mock.AddOutOfRangeNonceCalled(nonce, hash)
	}
}

// AddLeaderNonceAsOutOfRange -
func (mock *NtpSyncControllerMock) AddLeaderNonceAsOutOfRange(nonce uint64, hash string) {
	if mock.AddLeaderNonceAsOutOfRangeCalled != nil {
		mock.AddLeaderNonceAsOutOfRangeCalled(nonce, hash)
	}
}

// IsInterfaceNil -
func (mock *NtpSyncControllerMock) IsInterfaceNil() bool {
	return mock == nil
}
