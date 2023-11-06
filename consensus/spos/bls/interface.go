package bls

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

type SubRoundExtraDataSignatureHandler interface {
	CreateSignatureShare(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error)
	AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error
	Identifier() string
	IsInterfaceNil() bool
}

type SubRoundStartExtraSignatureHandler interface {
	Reset(pubKeys []string) error
	Identifier() string
	IsInterfaceNil() bool
}
