package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type NetworkShardingCollectorStub struct {
	UpdatePeerIdPublicKeyCalled  func(pid p2p.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled func(pk []byte, shardId uint32)
}

func (nscs *NetworkShardingCollectorStub) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	nscs.UpdatePeerIdPublicKeyCalled(pid, pk)
}

func (nscs *NetworkShardingCollectorStub) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nscs.UpdatePublicKeyShardIdCalled(pk, shardId)
}

func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
