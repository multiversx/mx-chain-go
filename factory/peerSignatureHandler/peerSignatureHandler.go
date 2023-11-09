package peerSignatureHandler

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/storage"
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
		return nil, errors.ErrNilCacher
	}
	if check.IfNil(singleSigner) {
		return nil, errors.ErrNilSingleSigner
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
	if len(pk) == 0 {
		return crypto.ErrInvalidPublicKey
	}
	if len(pid) == 0 {
		return errors.ErrInvalidPID
	}
	if len(signature) == 0 {
		return errors.ErrInvalidSignature
	}

	senderPubKey, err := psh.keygen.PublicKeyFromByteArray(pk)
	if err != nil {
		return err
	}

	retrievedPID, retrievedSig := psh.getBufferedPIDSignature(pk)
	newPidAndSig := pid != retrievedPID && !bytes.Equal(retrievedSig, signature)
	shouldVerify := len(retrievedPID) == 0 || len(retrievedSig) == 0 || newPidAndSig
	if shouldVerify {
		err = psh.singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
		if err != nil {
			return err
		}

		psh.bufferPIDSignature(pk, pid, signature)
		return nil
	}

	if retrievedPID != pid {
		return errors.ErrPIDMismatch
	}

	if !bytes.Equal(retrievedSig, signature) {
		return errors.ErrSignatureMismatch
	}

	return nil
}

// GetPeerSignature returns the needed signature if it is already cached.
// Otherwise, the signature will be computed.
func (psh *peerSignatureHandler) GetPeerSignature(privateKey crypto.PrivateKey, pid []byte) ([]byte, error) {
	privateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	retrievedPID, retrievedSig := psh.getBufferedPIDSignature(privateKeyBytes)
	isValidPID := len(retrievedPID) > 0 && retrievedPID == core.PeerID(pid)
	if isValidPID {
		return retrievedSig, nil
	}

	signature, err := psh.singleSigner.Sign(privateKey, pid)
	if err != nil {
		return nil, err
	}

	psh.bufferPIDSignature(privateKeyBytes, core.PeerID(pid), signature)
	return signature, nil
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
