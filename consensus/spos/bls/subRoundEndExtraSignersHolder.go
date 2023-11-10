package bls

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
)

type subRoundEndExtraSignersHolder struct {
	mutExtraSigners sync.RWMutex
	extraSigners    map[string]consensus.SubRoundEndExtraSignatureAggregatorHandler
}

func NewSubRoundEndExtraSignersHolder() *subRoundEndExtraSignersHolder {
	return &subRoundEndExtraSignersHolder{
		mutExtraSigners: sync.RWMutex{},
		extraSigners:    make(map[string]consensus.SubRoundEndExtraSignatureAggregatorHandler),
	}
}

func (holder *subRoundEndExtraSignersHolder) AggregateSignatures(bitmap []byte, epoch uint32) (map[string][]byte, error) {
	aggregatedSigs := make(map[string][]byte)

	holder.mutExtraSigners.RLock()
	for id, extraSigner := range holder.extraSigners {
		aggregatedSig, err := extraSigner.AggregateSignatures(bitmap, epoch)
		if err != nil {
			log.Debug("holder.extraSigner.AddLeaderAndAggregatedSignatures",
				"error", err.Error(),
				"id", id,
			)
			return nil, err
		}

		aggregatedSigs[id] = aggregatedSig
	}
	holder.mutExtraSigners.RUnlock()

	return aggregatedSigs, nil
}

func (holder *subRoundEndExtraSignersHolder) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		err := extraSigner.AddLeaderAndAggregatedSignatures(header, cnsMsg)
		if err != nil {
			log.Debug("holder.extraSigner.AddLeaderAndAggregatedSignatures",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

func (holder *subRoundEndExtraSignersHolder) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		err := extraSigner.SignAndSetLeaderSignature(header, leaderPubKey)
		if err != nil {
			log.Debug("holder.extraSigner.SignAndSetLeaderSignature",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

func (holder *subRoundEndExtraSignersHolder) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSigs map[string][]byte) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		aggregatedSig, found := aggregatedSigs[id]
		if !found {
			return fmt.Errorf("aggregated sig not found for signer id=%s", id)
		}

		err := extraSigner.SetAggregatedSignatureInHeader(header, aggregatedSig)
		if err != nil {
			log.Debug("holder.extraSigner.SetAggregatedSignatureInHeader",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

func (holder *subRoundEndExtraSignersHolder) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		err := extraSigner.VerifyAggregatedSignatures(bitmap, header)
		if err != nil {
			log.Debug("holder.extraSigner.VerifyAggregatedSignatures",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

func (holder *subRoundEndExtraSignersHolder) HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	holder.mutExtraSigners.RLock()
	defer holder.mutExtraSigners.RUnlock()

	for id, extraSigner := range holder.extraSigners {
		err := extraSigner.HaveConsensusHeaderWithFullInfo(header, cnsMsg)
		if err != nil {
			log.Debug("holder.extraSigner.HaveConsensusHeaderWithFullInfo",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

func (holder *subRoundEndExtraSignersHolder) RegisterExtraEndRoundSigAggregatorHandler(extraSignatureAggregator consensus.SubRoundEndExtraSignatureAggregatorHandler) error {
	if check.IfNil(extraSignatureAggregator) {
		return errors.ErrNilExtraSubRoundSigner
	}

	id := extraSignatureAggregator.Identifier()
	log.Debug("holder.RegisterExtraEndRoundSigAggregatorHandler", "identifier", id)

	holder.mutExtraSigners.Lock()
	holder.extraSigners[id] = extraSignatureAggregator
	holder.mutExtraSigners.Unlock()

	return nil
}

func (holder *subRoundEndExtraSignersHolder) IsInterfaceNil() bool {
	return holder == nil
}
