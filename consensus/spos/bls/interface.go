package bls

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

type SubRoundStartExtraSignersHolder interface {
	Reset(pubKeys []string) error
	RegisterExtraSingingHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error
	IsInterfaceNil() bool
}

type SubRoundSignatureExtraSignatureHandler interface {
	CreateSignatureShare(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error)
	AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error
	Identifier() string
	IsInterfaceNil() bool
}
