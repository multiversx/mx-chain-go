package mock

import "github.com/ElrondNetwork/elrond-go/core"

// FloodPreventerStub -
type FloodPreventerStub struct {
	IncreaseLoadCalled       func(fromConnectedPid core.PeerID, originator core.PeerID, size uint64) error
	ApplyConsensusSizeCalled func(size int)
	ResetCalled              func()
}

// IncreaseLoad -
func (fps *FloodPreventerStub) IncreaseLoad(fromConnectedPid core.PeerID, originator core.PeerID, size uint64) error {
	return fps.IncreaseLoadCalled(fromConnectedPid, originator, size)
}

// ApplyConsensusSize -
func (fps *FloodPreventerStub) ApplyConsensusSize(size int) {
	if fps.ApplyConsensusSizeCalled != nil {
		fps.ApplyConsensusSizeCalled(size)
	}
}

// Reset -
func (fps *FloodPreventerStub) Reset() {
	fps.ResetCalled()
}

// IsInterfaceNil -
func (fps *FloodPreventerStub) IsInterfaceNil() bool {
	return fps == nil
}
