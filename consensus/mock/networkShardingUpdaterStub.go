package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type NetworkShardingUpdaterStub struct {
	UpdatePeerIdPublicKeyCalled  func(pid p2p.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled func(pk []byte, shardId uint32)
}

func (nsus *NetworkShardingUpdaterStub) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	nsus.UpdatePeerIdPublicKeyCalled(pid, pk)
}

func (nsus *NetworkShardingUpdaterStub) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nsus.UpdatePublicKeyShardIdCalled(pk, shardId)
}

func (nsus *NetworkShardingUpdaterStub) IsInterfaceNil() bool {
	return nsus == nil
}
