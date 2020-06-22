package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// PeerDenialEvaluatorStub -
type PeerDenialEvaluatorStub struct {
	IsDeniedCalled func(pid core.PeerID) bool
}

// IsDenied -
func (pdes *PeerDenialEvaluatorStub) IsDenied(pid core.PeerID) bool {
	if pdes.IsDeniedCalled != nil {
		return pdes.IsDeniedCalled(pid)
	}

	return false
}

// IsInterfaceNil -
func (pdes *PeerDenialEvaluatorStub) IsInterfaceNil() bool {
	return pdes == nil
}
