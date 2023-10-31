package bls

import "github.com/multiversx/mx-chain-go/consensus"

type SubRoundExtraDataSignatureHandler interface {
	CreateSignatureShare(selfIndex uint16) ([]byte, error)
	AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error
	Identifier() string
	IsInterfaceNil() bool
}
