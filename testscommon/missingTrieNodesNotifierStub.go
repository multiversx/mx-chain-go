package testscommon

import "github.com/multiversx/mx-chain-go/common"

// MissingTrieNodesNotifierStub -
type MissingTrieNodesNotifierStub struct {
	RegisterHandlerCalled       func(handler common.StateSyncNotifierSubscriber)
	NotifyMissingTrieNodeCalled func(hash []byte)
}

func (mtnns *MissingTrieNodesNotifierStub) RegisterHandler(handler common.StateSyncNotifierSubscriber) {
	if mtnns.RegisterHandlerCalled != nil {
		mtnns.RegisterHandlerCalled(handler)
	}
}

func (mtnns *MissingTrieNodesNotifierStub) NotifyMissingTrieNode(hash []byte) {
	if mtnns.NotifyMissingTrieNodeCalled != nil {
		mtnns.NotifyMissingTrieNodeCalled(hash)
	}
}

func (mtnns *MissingTrieNodesNotifierStub) IsInterfaceNil() bool {
	return mtnns == nil
}
