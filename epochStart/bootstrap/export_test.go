package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (smbi *simpleMetaBlockInterceptor) GetReceivedMetablocks() map[string]*block.MetaBlock {
	smbi.mutReceivedMetaBlocks.RLock()
	defer smbi.mutReceivedMetaBlocks.RUnlock()

	return smbi.mapReceivedMetaBlocks
}

func (smbi *simpleMetaBlockInterceptor) GetPeersSliceForMetablocks() map[string][]p2p.PeerID {
	smbi.mutReceivedMetaBlocks.RLock()
	defer smbi.mutReceivedMetaBlocks.RUnlock()

	return smbi.mapMetaBlocksFromPeers
}
