package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundSignatureExtraSignatureHandlerMock -
type SubRoundSignatureExtraSignatureHandlerMock struct {
	CreateSignatureShareCalled          func(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error)
	AddSigShareToConsensusMessageCalled func(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShareCalled           func(index uint16, cnsMsg *consensus.Message) error
	IdentifierCalled                    func() string
}

// CreateSignatureShare -
func (mock *SubRoundSignatureExtraSignatureHandlerMock) CreateSignatureShare(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error) {
	if mock.CreateSignatureShareCalled != nil {
		return mock.CreateSignatureShareCalled(header, selfIndex, selfPubKey)
	}
	return nil, nil
}

// AddSigShareToConsensusMessage -
func (mock *SubRoundSignatureExtraSignatureHandlerMock) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) {
	if mock.AddSigShareToConsensusMessageCalled != nil {
		mock.AddSigShareToConsensusMessageCalled(sigShare, cnsMsg)
	}
}

// StoreSignatureShare -
func (mock *SubRoundSignatureExtraSignatureHandlerMock) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	if mock.StoreSignatureShareCalled != nil {
		return mock.StoreSignatureShareCalled(index, cnsMsg)
	}
	return nil
}

// Identifier -
func (mock *SubRoundSignatureExtraSignatureHandlerMock) Identifier() string {
	if mock.IdentifierCalled != nil {
		return mock.IdentifierCalled()
	}
	return ""
}

// IsInterfaceNil -
func (mock *SubRoundSignatureExtraSignatureHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
