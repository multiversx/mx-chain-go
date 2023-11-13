package bls

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
)

type subRoundStartExtraSignersHolder struct {
	mutExtraSigners sync.RWMutex
	extraSigners    map[string]consensus.SubRoundStartExtraSignatureHandler
}

func NewSubRoundStartExtraSignersHolder() *subRoundStartExtraSignersHolder {
	return &subRoundStartExtraSignersHolder{
		mutExtraSigners: sync.RWMutex{},
		extraSigners:    make(map[string]consensus.SubRoundStartExtraSignatureHandler),
	}
}

func (holder *subRoundStartExtraSignersHolder) Reset(pubKeys []string) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		err := extraSigner.Reset(pubKeys)
		if err != nil {
			log.Debug("holder.extraSigner.subRoundStartExtraSignersHolder",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

func (holder *subRoundStartExtraSignersHolder) RegisterExtraSigningHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error {
	if check.IfNil(extraSigner) {
		return errors.ErrNilExtraSubRoundSigner
	}

	id := extraSigner.Identifier()
	log.Debug("holder.registerExtraSingingHandler", "identifier", id)

	holder.mutExtraSigners.Lock()
	defer holder.mutExtraSigners.Unlock()

	if _, exists := holder.extraSigners[id]; exists {
		return errors.ErrExtraSignerIdAlreadyExists
	}

	holder.extraSigners[id] = extraSigner
	return nil
}

func (holder *subRoundStartExtraSignersHolder) IsInterfaceNil() bool {
	return holder == nil
}
