package subRounds

import (
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
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
	return &SubRoundStartExtraSignersHolderMock{}
}

// GetSubRoundSignatureExtraSignersHolder -
func (mock *ExtraSignersHolderMock) GetSubRoundSignatureExtraSignersHolder() bls.SubRoundSignatureExtraSignersHolder {
	if mock.GetSubRoundSignatureExtraSignersHolderCalled != nil {
		return mock.GetSubRoundSignatureExtraSignersHolderCalled()
	}
	return &SubRoundSignatureExtraSignersHolderMock{}
}

// GetSubRoundEndExtraSignersHolder -
func (mock *ExtraSignersHolderMock) GetSubRoundEndExtraSignersHolder() bls.SubRoundEndExtraSignersHolder {
	if mock.GetSubRoundEndExtraSignersHolderCalled != nil {
		return mock.GetSubRoundEndExtraSignersHolderCalled()
	}
	return &SubRoundEndExtraSignersHolderMock{}
}

// IsInterfaceNil -
func (mock *ExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
