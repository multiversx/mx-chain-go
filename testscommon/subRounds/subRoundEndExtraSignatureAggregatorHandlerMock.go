package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundEndExtraSignatureMock -
type SubRoundEndExtraSignatureMock struct {
	AggregateSignaturesCalled              func(bitmap []byte, epoch uint32) ([]byte, error)
	AddLeaderAndAggregatedSignaturesCalled func(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignatureCalled        func(header data.HeaderHandler, leaderPubKey []byte) error
	SetAggregatedSignatureInHeaderCalled   func(header data.HeaderHandler, aggregatedSig []byte) error
	HaveConsensusHeaderWithFullInfoCalled  func(header data.HeaderHandler, cnsMsg *consensus.Message) error
	VerifyAggregatedSignaturesCalled       func(bitmap []byte, header data.HeaderHandler) error
	IdentifierCalled                       func() string
}

// AggregateAndSetSignatures -
func (mock *SubRoundEndExtraSignatureMock) AggregateAndSetSignatures(bitmap []byte, epoch uint32) ([]byte, error) {
	if mock.AggregateSignaturesCalled != nil {
		return mock.AggregateSignaturesCalled(bitmap, epoch)
	}
	return nil, nil
}

// AddLeaderAndAggregatedSignatures -
func (mock *SubRoundEndExtraSignatureMock) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	if mock.AddLeaderAndAggregatedSignaturesCalled != nil {
		return mock.AddLeaderAndAggregatedSignaturesCalled(header, cnsMsg)
	}
	return nil
}

// SignAndSetLeaderSignature -
func (mock *SubRoundEndExtraSignatureMock) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	if mock.SignAndSetLeaderSignatureCalled != nil {
		return mock.SignAndSetLeaderSignatureCalled(header, leaderPubKey)
	}
	return nil
}

// SetAggregatedSignatureInHeader -
func (mock *SubRoundEndExtraSignatureMock) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	if mock.SetAggregatedSignatureInHeaderCalled != nil {
		return mock.SetAggregatedSignatureInHeaderCalled(header, aggregatedSig)
	}
	return nil
}

// SetConsensusDataInHeader -
func (mock *SubRoundEndExtraSignatureMock) SetConsensusDataInHeader(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	if mock.HaveConsensusHeaderWithFullInfoCalled != nil {
		return mock.HaveConsensusHeaderWithFullInfoCalled(header, cnsMsg)
	}
	return nil
}

// VerifyAggregatedSignatures -
func (mock *SubRoundEndExtraSignatureMock) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	if mock.VerifyAggregatedSignaturesCalled != nil {
		return mock.VerifyAggregatedSignaturesCalled(bitmap, header)
	}
	return nil
}

// Identifier -
func (mock *SubRoundEndExtraSignatureMock) Identifier() string {
	if mock.IdentifierCalled != nil {
		return mock.IdentifierCalled()
	}
	return ""
}

// IsInterfaceNil -
func (mock *SubRoundEndExtraSignatureMock) IsInterfaceNil() bool {
	return mock == nil
}
