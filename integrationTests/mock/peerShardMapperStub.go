package mock

import "github.com/ElrondNetwork/elrond-go/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
}

// GetPeerInfo -
func (psms *PeerShardMapperStub) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (psms *PeerShardMapperStub) IsInterfaceNil() bool {
	return psms == nil
}
