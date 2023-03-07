package peerSignatureHandler

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/errors"
)

func (psh *peerSignatureHandler) GetPIDAndSig(entry interface{}) (core.PeerID, []byte, error) {
	pidSig, ok := entry.(*pidSignature)
	if !ok {
		return "", nil, errors.ErrWrongTypeAssertion
	}

	return pidSig.pid, pidSig.signature, nil
}

func (psh *peerSignatureHandler) GetCacheEntry(pid core.PeerID, sig []byte) *pidSignature {
	return &pidSignature{
		pid:       pid,
		signature: sig,
	}
}
