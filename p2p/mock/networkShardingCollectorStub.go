package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type NetworkShardingCollectorStub struct {
	UpdatePeerIdPublicKeyCalled func(pid p2p.PeerID, pk []byte)
}

func (nscs *NetworkShardingCollectorStub) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	nscs.UpdatePeerIdPublicKeyCalled(pid, pk)
}

func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
