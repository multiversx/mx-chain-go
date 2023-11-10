package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundEndExtraSignatureAggregatorHandlerMock -
type SubRoundEndExtraSignatureAggregatorHandlerMock struct {
	AggregateSignaturesCalled              func(bitmap []byte, epoch uint32) ([]byte, error)
	AddLeaderAndAggregatedSignaturesCalled func(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignatureCalled        func(header data.HeaderHandler, leaderPubKey []byte) error
	SeAggregatedSignatureInHeaderCalled    func(header data.HeaderHandler, aggregatedSig []byte) error
	HaveConsensusHeaderWithFullInfoCalled  func(header data.HeaderHandler, cnsMsg *consensus.Message) error
	VerifyAggregatedSignaturesCalled       func(bitmap []byte, header data.HeaderHandler) error
	IdentifierCalled                       func() string
}

// AggregateSignatures -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) AggregateSignatures(bitmap []byte, epoch uint32) ([]byte, error) {
	if mock.AggregateSignaturesCalled != nil {
		return mock.AggregateSignaturesCalled(bitmap, epoch)
	}
	return nil, nil
}

// AddLeaderAndAggregatedSignatures -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	if mock.AddLeaderAndAggregatedSignaturesCalled != nil {
		return mock.AddLeaderAndAggregatedSignaturesCalled(header, cnsMsg)
	}
	return nil
}

// SignAndSetLeaderSignature -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	if mock.SignAndSetLeaderSignatureCalled != nil {
		return mock.SignAndSetLeaderSignatureCalled(header, leaderPubKey)
	}
	return nil
}

// SetAggregatedSignatureInHeader -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	if mock.SeAggregatedSignatureInHeaderCalled != nil {
		return mock.SeAggregatedSignatureInHeaderCalled(header, aggregatedSig)
	}
	return nil
}

// HaveConsensusHeaderWithFullInfo -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	if mock.HaveConsensusHeaderWithFullInfoCalled != nil {
		return mock.HaveConsensusHeaderWithFullInfoCalled(header, cnsMsg)
	}
	return nil
}

// VerifyAggregatedSignatures -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	if mock.VerifyAggregatedSignaturesCalled != nil {
		return mock.VerifyAggregatedSignaturesCalled(bitmap, header)
	}
	return nil
}

// Identifier -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) Identifier() string {
	if mock.IdentifierCalled != nil {
		return mock.IdentifierCalled()
	}
	return ""
}

// IsInterfaceNil -
func (mock *SubRoundEndExtraSignatureAggregatorHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
