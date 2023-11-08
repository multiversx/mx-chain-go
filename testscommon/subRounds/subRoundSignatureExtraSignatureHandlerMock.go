package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundSignatureExtraSignatureHandlerStub -
type SubRoundSignatureExtraSignatureHandlerStub struct {
	CreateSignatureShareCalled          func(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error)
	AddSigShareToConsensusMessageCalled func(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShareCalled           func(index uint16, cnsMsg *consensus.Message) error
	IdentifierCalled                    func() string
}

// CreateSignatureShare -
func (stub *SubRoundSignatureExtraSignatureHandlerStub) CreateSignatureShare(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error) {
	if stub.CreateSignatureShareCalled != nil {
		return stub.CreateSignatureShareCalled(header, selfIndex, selfPubKey)
	}
	return nil, nil
}

// AddSigShareToConsensusMessage -
func (stub *SubRoundSignatureExtraSignatureHandlerStub) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) {
	if stub.AddSigShareToConsensusMessageCalled != nil {
		stub.AddSigShareToConsensusMessageCalled(sigShare, cnsMsg)
	}
}

// StoreSignatureShare -
func (stub *SubRoundSignatureExtraSignatureHandlerStub) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	if stub.StoreSignatureShareCalled != nil {
		return stub.StoreSignatureShareCalled(index, cnsMsg)
	}
	return nil
}

// Identifier -
func (stub *SubRoundSignatureExtraSignatureHandlerStub) Identifier() string {
	if stub.IdentifierCalled != nil {
		return stub.IdentifierCalled()
	}
	return ""
}

// IsInterfaceNil -
func (stub *SubRoundSignatureExtraSignatureHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
