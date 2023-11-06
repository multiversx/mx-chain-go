package bls

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
)

type subRoundStartExtraSignersHolder struct {
	mutExtraSigners sync.RWMutex
	extraSigners    map[string]SubRoundStartExtraSignatureHandler
}

func newSubRoundStartExtraSignersHolder() *subRoundStartExtraSignersHolder {
	return &subRoundStartExtraSignersHolder{
		mutExtraSigners: sync.RWMutex{},
		extraSigners:    make(map[string]SubRoundStartExtraSignatureHandler),
	}
}

func (holder *subRoundStartExtraSignersHolder) reset(pubKeys []string) error {
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

func (holder *subRoundStartExtraSignersHolder) registerExtraSingingHandler(extraSigner SubRoundStartExtraSignatureHandler) error {
	if check.IfNil(extraSigner) {
		return errors.ErrNilExtraSubRoundSigner
	}

	id := extraSigner.Identifier()
	log.Debug("holder.registerExtraSingingHandler", "identifier", id)

	holder.mutExtraSigners.Lock()
	holder.extraSigners[id] = extraSigner
	holder.mutExtraSigners.Unlock()

	return nil
}
