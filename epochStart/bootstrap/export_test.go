package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (s *simpleEpochStartMetaBlockInterceptor) GetReceivedMetablocks() map[string]*block.MetaBlock {
	s.mutReceivedMetaBlocks.RLock()
	defer s.mutReceivedMetaBlocks.RUnlock()

	return s.mapReceivedMetaBlocks
}

func (s *simpleEpochStartMetaBlockInterceptor) GetPeersSliceForMetablocks() map[string][]p2p.PeerID {
	s.mutReceivedMetaBlocks.RLock()
	defer s.mutReceivedMetaBlocks.RUnlock()

	return s.mapMetaBlocksFromPeers
}
