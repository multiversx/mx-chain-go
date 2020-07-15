package peerSignatureHandler

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// pidSignature holds the peer id and the associated signature
type pidSignature struct {
	pid       core.PeerID
	signature []byte
}

// peerSignatureHandler is used as a wrapper for the single signer, that buffers signatures.
type peerSignatureHandler struct {
	pkPIDSignature storage.Cacher
	singleSigner   crypto.SingleSigner
	keygen         crypto.KeyGenerator
}

// NewPeerSignatureHandler creates a new instance of peerSignatureHandler.
func NewPeerSignatureHandler(
	pkPIDSignature storage.Cacher,
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
) (*peerSignatureHandler, error) {
	if check.IfNil(pkPIDSignature) {
		return nil, crypto.ErrNilCacher
	}
	if check.IfNil(singleSigner) {
		return nil, crypto.ErrNilSingleSigner
	}
	if check.IfNil(keygen) {
		return nil, crypto.ErrNilKeyGenerator
	}

	return &peerSignatureHandler{
		pkPIDSignature: pkPIDSignature,
		singleSigner:   singleSigner,
		keygen:         keygen,
	}, nil
}

// VerifyPeerSignature verifies the signature associated with the public key. It first checks the cache for the public key,
// and if it is not present, it will recompute the signature.
func (psh *peerSignatureHandler) VerifyPeerSignature(pk []byte, pid core.PeerID, signature []byte) error {
	senderPubKey, err := psh.keygen.PublicKeyFromByteArray(pk)
	if err != nil {
		return err
	}

	retrievedPid, retrievedSig := psh.getBufferedPIDSignature(pk)
	if retrievedPid == "" || retrievedSig == nil {
		err = psh.singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
		if err != nil {
			return err
		}

		psh.bufferPIDSignature(pk, pid, signature)
		return nil
	}

	if retrievedPid != pid {
		return crypto.ErrPIDMissmatch
	}

	if !bytes.Equal(retrievedSig, signature) {
		return crypto.ErrSignatureMissmatch
	}

	return nil
}

// GetPeerSignature checks if the needed signature is not present in the buffer, and returns it.
// If it is not present, the signature will be computed
func (psh *peerSignatureHandler) GetPeerSignature(key crypto.PrivateKey, pid []byte) ([]byte, error) {
	return psh.singleSigner.Sign(key, pid)
}

func (psh *peerSignatureHandler) bufferPIDSignature(pk []byte, pid core.PeerID, signature []byte) {
	pidSig := &pidSignature{
		pid:       pid,
		signature: signature,
	}

	entryLen := len(pid) + len(signature)
	psh.pkPIDSignature.Put(pk, pidSig, entryLen)
}

func (psh *peerSignatureHandler) getBufferedPIDSignature(pk []byte) (pid core.PeerID, signature []byte) {
	entry, ok := psh.pkPIDSignature.Get(pk)
	if !ok {
		return "", nil
	}

	pidSig, ok := entry.(*pidSignature)
	if !ok {
		return "", nil
	}

	return pidSig.pid, pidSig.signature
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *peerSignatureHandler) IsInterfaceNil() bool {
	return psh == nil
}
