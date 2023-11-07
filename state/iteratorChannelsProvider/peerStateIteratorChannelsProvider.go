package iteratorChannelsProvider

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
)

type peerStateIteratorChannelsProvider struct {
}

// NewPeerStateIteratorChannelsProvider creates a new instance of peer state iterator channels provider
func NewPeerStateIteratorChannelsProvider() *peerStateIteratorChannelsProvider {
	return &peerStateIteratorChannelsProvider{}
}

// GetIteratorChannels returns trie iterator channels wit a nil leaves channel
func (psicp *peerStateIteratorChannelsProvider) GetIteratorChannels() *common.TrieIteratorChannels {
	return &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    errChan.NewErrChanWrapper(),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (psicp *peerStateIteratorChannelsProvider) IsInterfaceNil() bool {
	return psicp == nil
}
