package testscommon

// StateSyncNotifierSubscriberStub -
type StateSyncNotifierSubscriberStub struct {
	MissingDataTrieNodeFoundCalled func(hash []byte)
}

// MissingDataTrieNodeFound -
func (ssns *StateSyncNotifierSubscriberStub) MissingDataTrieNodeFound(hash []byte) {
	if ssns.MissingDataTrieNodeFoundCalled != nil {
		ssns.MissingDataTrieNodeFoundCalled(hash)
	}
}

// IsInterfaceNil -
func (ssns *StateSyncNotifierSubscriberStub) IsInterfaceNil() bool {
	return ssns == nil
}
