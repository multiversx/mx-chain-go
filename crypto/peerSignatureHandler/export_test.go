package peerSignatureHandler

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
)

func (psh *peerSignatureHandler) GetPIDAndSig(entry interface{}) (core.PeerID, []byte, error) {
	pidSig, ok := entry.(*pidSignature)
	if !ok {
		return "", nil, crypto.ErrWrongTypeAssertion
	}

	return pidSig.pid, pidSig.signature, nil
}

func (psh *peerSignatureHandler) GetCacheEntry(pid core.PeerID, sig []byte) *pidSignature {
	return &pidSignature{
		pid:       pid,
		signature: sig,
	}
}
