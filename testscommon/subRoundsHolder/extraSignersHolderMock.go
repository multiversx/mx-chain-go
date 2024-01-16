package subRoundsHolder

import (
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
)

// ExtraSignersHolderMock -
type ExtraSignersHolderMock struct {
	GetSubRoundStartExtraSignersHolderCalled     func() bls.SubRoundStartExtraSignersHolder
	GetSubRoundSignatureExtraSignersHolderCalled func() bls.SubRoundSignatureExtraSignersHolder
	GetSubRoundEndExtraSignersHolderCalled       func() bls.SubRoundEndExtraSignersHolder
}

// GetSubRoundStartExtraSignersHolder -
func (mock *ExtraSignersHolderMock) GetSubRoundStartExtraSignersHolder() bls.SubRoundStartExtraSignersHolder {
	if mock.GetSubRoundStartExtraSignersHolderCalled != nil {
		return mock.GetSubRoundStartExtraSignersHolderCalled()
	}
	return &subRounds.SubRoundStartExtraSignersHolderMock{}
}

// GetSubRoundSignatureExtraSignersHolder -
func (mock *ExtraSignersHolderMock) GetSubRoundSignatureExtraSignersHolder() bls.SubRoundSignatureExtraSignersHolder {
	if mock.GetSubRoundSignatureExtraSignersHolderCalled != nil {
		return mock.GetSubRoundSignatureExtraSignersHolderCalled()
	}
	return &subRounds.SubRoundSignatureExtraSignersHolderMock{}
}

// GetSubRoundEndExtraSignersHolder -
func (mock *ExtraSignersHolderMock) GetSubRoundEndExtraSignersHolder() bls.SubRoundEndExtraSignersHolder {
	if mock.GetSubRoundEndExtraSignersHolderCalled != nil {
		return mock.GetSubRoundEndExtraSignersHolderCalled()
	}
	return &subRounds.SubRoundEndExtraSignersHolderMock{}
}

// IsInterfaceNil -
func (mock *ExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
