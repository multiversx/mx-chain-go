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

type SubRoundSignatureExtraSignersHolder interface {
	CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error)
	AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error
	StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error
	RegisterExtraSingingHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error
	IsInterfaceNil() bool
}
