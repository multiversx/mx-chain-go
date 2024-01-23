package headerCheck

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

type extraHeaderSigVerifierHolder struct {
	mutExtraVerifiers sync.RWMutex
	extraVerifiers    map[string]process.ExtraHeaderSigVerifierHandler
}

// NewExtraHeaderSigVerifierHolder creates a holder for extra header sig verifiers
func NewExtraHeaderSigVerifierHolder() *extraHeaderSigVerifierHolder {
	return &extraHeaderSigVerifierHolder{
		mutExtraVerifiers: sync.RWMutex{},
		extraVerifiers:    make(map[string]process.ExtraHeaderSigVerifierHandler),
	}
}

// VerifyAggregatedSignature calls VerifyAggregatedSignature for all registered verifiers
func (holder *extraHeaderSigVerifierHolder) VerifyAggregatedSignature(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error {
	holder.mutExtraVerifiers.RLock()
	defer holder.mutExtraVerifiers.RUnlock()

	for id, extraSigner := range holder.extraVerifiers {
		err := extraSigner.VerifyAggregatedSignature(header, multiSigVerifier, pubKeysSigners)
		if err != nil {
			log.Debug("holder.VerifyAggregatedSignature",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

// VerifyLeaderSignature calls VerifyLeaderSignature for all registered verifiers
func (holder *extraHeaderSigVerifierHolder) VerifyLeaderSignature(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error {
	holder.mutExtraVerifiers.RLock()
	defer holder.mutExtraVerifiers.RUnlock()

	for id, extraSigner := range holder.extraVerifiers {
		err := extraSigner.VerifyLeaderSignature(header, leaderPubKey)
		if err != nil {
			log.Debug("holder.VerifyLeaderSignature",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

// RemoveLeaderSignature calls RemoveLeaderSignature for all registered verifiers
func (holder *extraHeaderSigVerifierHolder) RemoveLeaderSignature(header data.HeaderHandler) error {
	holder.mutExtraVerifiers.RLock()
	defer holder.mutExtraVerifiers.RUnlock()

	for id, extraSigner := range holder.extraVerifiers {
		err := extraSigner.RemoveLeaderSignature(header)
		if err != nil {
			log.Debug("holder.RemoveLeaderSignature",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

// RemoveAllSignatures calls RemoveAllSignatures for all registered verifiers
func (holder *extraHeaderSigVerifierHolder) RemoveAllSignatures(header data.HeaderHandler) error {
	holder.mutExtraVerifiers.RLock()
	defer holder.mutExtraVerifiers.RUnlock()

	for id, extraSigner := range holder.extraVerifiers {
		err := extraSigner.RemoveAllSignatures(header)
		if err != nil {
			log.Debug("holder.RemoveAllSignatures",
				"error", err.Error(),
				"id", id,
			)
			return err
		}
	}

	return nil
}

// RegisterExtraHeaderSigVerifier will register a new extra header sig verifier
func (holder *extraHeaderSigVerifierHolder) RegisterExtraHeaderSigVerifier(extraVerifier process.ExtraHeaderSigVerifierHandler) error {
	if check.IfNil(extraVerifier) {
		return errors.ErrNilExtraSubRoundSigner
	}

	id := extraVerifier.Identifier()
	log.Debug("holder.RegisterExtraHeaderSigVerifier", "identifier", id)

	holder.mutExtraVerifiers.Lock()
	defer holder.mutExtraVerifiers.Unlock()

	if _, exists := holder.extraVerifiers[id]; exists {
		return errors.ErrExtraSignerIdAlreadyExists
	}

	holder.extraVerifiers[id] = extraVerifier
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (holder *extraHeaderSigVerifierHolder) IsInterfaceNil() bool {
	return holder == nil
}
