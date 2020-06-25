package mock

import "github.com/ElrondNetwork/elrond-go/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
	GetPeerInfoCalled func(pid core.PeerID) core.P2PPeerInfo
}

// GetPeerInfo -
func (psms *PeerShardMapperStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	if psms.GetPeerInfoCalled != nil {
		return psms.GetPeerInfoCalled(pid)
	}

	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (psms *PeerShardMapperStub) IsInterfaceNil() bool {
	return psms == nil
}
