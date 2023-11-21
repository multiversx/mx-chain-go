package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundEndExtraSignersHolderMock -
type SubRoundEndExtraSignersHolderMock struct {
	AggregateSignaturesCalled                       func(bitmap []byte, epoch uint32) (map[string][]byte, error)
	AddLeaderAndAggregatedSignaturesCalled          func(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignatureCalled                 func(header data.HeaderHandler, leaderPubKey []byte) error
	SetAggregatedSignatureInHeaderCalled            func(header data.HeaderHandler, aggregatedSigs map[string][]byte) error
	VerifyAggregatedSignaturesCalled                func(bitmap []byte, header data.HeaderHandler) error
	HaveConsensusHeaderWithFullInfoCalled           func(header data.HeaderHandler, cnsMsg *consensus.Message) error
	RegisterExtraEndRoundSigAggregatorHandlerCalled func(extraSigner consensus.SubRoundEndExtraSignatureHandler) error
}

// AggregateSignatures -
func (mock *SubRoundEndExtraSignersHolderMock) AggregateSignatures(bitmap []byte, epoch uint32) (map[string][]byte, error) {
	if mock.AggregateSignaturesCalled != nil {
		return mock.AggregateSignaturesCalled(bitmap, epoch)
	}
	return nil, nil
}

// AddLeaderAndAggregatedSignatures -
func (mock *SubRoundEndExtraSignersHolderMock) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	if mock.AddLeaderAndAggregatedSignaturesCalled != nil {
		return mock.AddLeaderAndAggregatedSignaturesCalled(header, cnsMsg)
	}
	return nil
}

// SignAndSetLeaderSignature -
func (mock *SubRoundEndExtraSignersHolderMock) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	if mock.SignAndSetLeaderSignatureCalled != nil {
		return mock.SignAndSetLeaderSignatureCalled(header, leaderPubKey)
	}
	return nil
}

// SetAggregatedSignatureInHeader -
func (mock *SubRoundEndExtraSignersHolderMock) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSigs map[string][]byte) error {
	if mock.SetAggregatedSignatureInHeaderCalled != nil {
		return mock.SetAggregatedSignatureInHeaderCalled(header, aggregatedSigs)
	}
	return nil
}

// VerifyAggregatedSignatures -
func (mock *SubRoundEndExtraSignersHolderMock) VerifyAggregatedSignatures(header data.HeaderHandler, bitmap []byte) error {
	if mock.VerifyAggregatedSignaturesCalled != nil {
		return mock.VerifyAggregatedSignaturesCalled(bitmap, header)
	}
	return nil
}

// HaveConsensusHeaderWithFullInfo -
func (mock *SubRoundEndExtraSignersHolderMock) HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	if mock.HaveConsensusHeaderWithFullInfoCalled != nil {
		return mock.HaveConsensusHeaderWithFullInfoCalled(header, cnsMsg)
	}
	return nil
}

// RegisterExtraSigningHandler -
func (mock *SubRoundEndExtraSignersHolderMock) RegisterExtraSigningHandler(extraSigner consensus.SubRoundEndExtraSignatureHandler) error {
	if mock.RegisterExtraEndRoundSigAggregatorHandlerCalled != nil {
		return mock.RegisterExtraEndRoundSigAggregatorHandlerCalled(extraSigner)
	}
	return nil
}

// IsInterfaceNil -
func (mock *SubRoundEndExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
