package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundSignatureExtraSignersHolderMock -
type SubRoundSignatureExtraSignersHolderMock struct {
	CreateExtraSignatureSharesCalled          func(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error)
	AddExtraSigSharesToConsensusMessageCalled func(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error
	StoreExtraSignatureShareCalled            func(index uint16, cnsMsg *consensus.Message) error
	RegisterExtraSingingHandlerCalled         func(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error
}

// CreateExtraSignatureShares -
func (mock *SubRoundSignatureExtraSignersHolderMock) CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error) {
	if mock.CreateExtraSignatureSharesCalled != nil {
		return mock.CreateExtraSignatureSharesCalled(header, selfIndex, selfPubKey)
	}
	return nil, nil
}

// AddExtraSigSharesToConsensusMessage -
func (mock *SubRoundSignatureExtraSignersHolderMock) AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error {
	if mock.AddExtraSigSharesToConsensusMessageCalled != nil {
		return mock.AddExtraSigSharesToConsensusMessageCalled(extraSigShares, cnsMsg)
	}
	return nil
}

// StoreExtraSignatureShare -
func (mock *SubRoundSignatureExtraSignersHolderMock) StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	if mock.StoreExtraSignatureShareCalled != nil {
		return mock.StoreExtraSignatureShareCalled(index, cnsMsg)
	}
	return nil
}

// RegisterExtraSingingHandler -
func (mock *SubRoundSignatureExtraSignersHolderMock) RegisterExtraSingingHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error {
	if mock.RegisterExtraSingingHandlerCalled != nil {
		return mock.RegisterExtraSingingHandlerCalled(extraSigner)
	}
	return nil
}

// IsInterfaceNil -
func (mock *SubRoundSignatureExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
