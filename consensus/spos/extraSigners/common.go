package extraSigners

import (
	"github.com/multiversx/mx-chain-go/consensus"
)

func initExtraSignatureEntry(cnsMsg *consensus.Message, key string) {
	if cnsMsg.ExtraSignatures == nil {
		cnsMsg.ExtraSignatures = make(map[string]*consensus.ExtraSignatureData)
	}

	if _, found := cnsMsg.ExtraSignatures[key]; !found {
		cnsMsg.ExtraSignatures[key] = &consensus.ExtraSignatureData{}
	}
}
