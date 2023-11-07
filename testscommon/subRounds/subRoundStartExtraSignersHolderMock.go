package subRounds

import (
	"github.com/multiversx/mx-chain-go/consensus"
)

type SubRoundStartExtraSignersHolderMock struct {
}

func (mock *SubRoundStartExtraSignersHolderMock) Reset(pubKeys []string) error {
	return nil
}

func (mock *SubRoundStartExtraSignersHolderMock) RegisterExtraSingingHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error {
	return nil
}

func (mock *SubRoundStartExtraSignersHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
