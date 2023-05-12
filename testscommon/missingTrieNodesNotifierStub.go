package testscommon

import "github.com/multiversx/mx-chain-go/common"

// MissingTrieNodesNotifierStub -
type MissingTrieNodesNotifierStub struct {
	RegisterHandlerCalled            func(handler common.StateSyncNotifierSubscriber) error
	AsyncNotifyMissingTrieNodeCalled func(hash []byte)
}

// RegisterHandler -
func (mtnns *MissingTrieNodesNotifierStub) RegisterHandler(handler common.StateSyncNotifierSubscriber) error {
	if mtnns.RegisterHandlerCalled != nil {
		return mtnns.RegisterHandlerCalled(handler)
	}

	return nil
}

// AsyncNotifyMissingTrieNode -
func (mtnns *MissingTrieNodesNotifierStub) AsyncNotifyMissingTrieNode(hash []byte) {
	if mtnns.AsyncNotifyMissingTrieNodeCalled != nil {
		mtnns.AsyncNotifyMissingTrieNodeCalled(hash)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mtnns *MissingTrieNodesNotifierStub) IsInterfaceNil() bool {
	return mtnns == nil
}
