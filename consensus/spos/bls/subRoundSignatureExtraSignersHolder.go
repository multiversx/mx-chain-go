package bls

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
)

type subRoundSignatureExtraSignersHolder struct {
	mutExtraSigners sync.RWMutex
	extraSigners    map[string]consensus.SubRoundSignatureExtraSignatureHandler
}

func NewSubRoundSignatureExtraSignersHolder() *subRoundSignatureExtraSignersHolder {
	return &subRoundSignatureExtraSignersHolder{
		mutExtraSigners: sync.RWMutex{},
		extraSigners:    make(map[string]consensus.SubRoundSignatureExtraSignatureHandler),
	}
}

func (holder *subRoundSignatureExtraSignersHolder) CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error) {
	ret := make(map[string][]byte)

	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		extraSigShare, err := extraSigner.CreateSignatureShare(header, selfIndex, selfPubKey)
		if err != nil {
			log.Debug("holder.subRoundSignatureExtraSignersHolder.createExtraSignatureShares",
				"error", err.Error(), "id", id)
			return nil, err
		}

		ret[id] = extraSigShare
	}

	return ret, nil
}

func (holder *subRoundSignatureExtraSignersHolder) AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigShare := range extraSigShares {
		// this should never happen, but keep this sanity check anyway
		extraSigner, found := holder.extraSigners[id]
		if !found {
			return fmt.Errorf("extra signed not found for id=%s when trying to add extra sig share to consensus msg", id)
		}

		extraSigner.AddSigShareToConsensusMessage(extraSigShare, cnsMsg)
	}

	return nil
}

func (holder *subRoundSignatureExtraSignersHolder) StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		err := extraSigner.StoreSignatureShare(index, cnsMsg)
		if err != nil {
			log.Debug("holder.subRoundSignatureExtraSignersHolder.storeExtraSignatureShare",
				"error", err.Error(), "id", id)
			return err
		}
	}

	return nil
}

func (holder *subRoundSignatureExtraSignersHolder) RegisterExtraSigningHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error {
	if check.IfNil(extraSigner) {
		return errors.ErrNilExtraSubRoundSigner
	}

	id := extraSigner.Identifier()
	log.Debug("holder.subRoundStartExtraSignersHolder.registerExtraSingingHandler", "identifier", id)

	holder.mutExtraSigners.Lock()
	defer holder.mutExtraSigners.Unlock()

	if _, exists := holder.extraSigners[id]; exists {
		return errors.ErrExtraSignerIdAlreadyExists
	}

	holder.extraSigners[id] = extraSigner
	return nil
}

func (holder *subRoundSignatureExtraSignersHolder) IsInterfaceNil() bool {
	return holder == nil
}
