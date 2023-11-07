package subRounds

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

type SubRoundSignatureExtraSignersHolderMock struct {
}

func (mock *SubRoundSignatureExtraSignersHolderMock) CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (mock *SubRoundSignatureExtraSignersHolderMock) AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error {
	return nil
}

func (mock *SubRoundSignatureExtraSignersHolderMock) StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	return nil
}

func (mock *SubRoundSignatureExtraSignersHolderMock) RegisterExtraSingingHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error {
	return nil
}

func (mock *SubRoundSignatureExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
