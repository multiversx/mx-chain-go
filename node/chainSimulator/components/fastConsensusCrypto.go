package components

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"sync"

	crypto "github.com/multiversx/mx-chain-crypto-go"

	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	mxChainErrors "github.com/multiversx/mx-chain-go/errors"
)

const fastConsensusSignatureSize = sha512.Size384

var _ crypto.SingleSigner = (*fastConsensusSigner)(nil)
var _ crypto.MultiSignerV2 = (*fastConsensusSigner)(nil)
var _ cryptoCommon.MultiSignerContainer = (*fastConsensusMultiSignerContainer)(nil)

var (
	errInvalidFastConsensusSignature  = errors.New("invalid fast consensus signature")
	errUnknownFastConsensusPrivateKey = errors.New("unknown fast consensus private key")
)

// fastConsensusSigner is a deterministic, non-cryptographic signer used only by the chain
// simulator. Signatures remain bound to the validator public key and message, while avoiding the
// expensive BLS/CGo operations. It deliberately preserves the production SingleSigner and
// MultiSignerV2 interfaces so consensus still creates, collects, aggregates and verifies shares.
//
// This signer is not secure: anyone with the public key can reproduce a signature. It must never
// be installed outside an explicitly opted-in simulator.
type fastConsensusSigner struct {
	keyGenerator crypto.KeyGenerator

	mutPublicKeys      sync.RWMutex
	publicKeyByPrivate map[string][]byte
}

func newFastConsensusSigner(
	privateKey crypto.PrivateKey,
	publicKey crypto.PublicKey,
	keyGenerator crypto.KeyGenerator,
) (*fastConsensusSigner, error) {
	signer := &fastConsensusSigner{
		keyGenerator:       keyGenerator,
		publicKeyByPrivate: make(map[string][]byte),
	}

	err := signer.rememberKeyPair(privateKey, publicKey)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func (signer *fastConsensusSigner) rememberKeyPair(
	privateKey crypto.PrivateKey,
	publicKey crypto.PublicKey,
) error {
	if privateKey == nil || privateKey.IsInterfaceNil() {
		return crypto.ErrNilPrivateKey
	}
	if publicKey == nil || publicKey.IsInterfaceNil() {
		return crypto.ErrEmptyPubKey
	}

	privateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return err
	}
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return err
	}

	signer.mutPublicKeys.Lock()
	signer.publicKeyByPrivate[string(privateKeyBytes)] = append([]byte(nil), publicKeyBytes...)
	signer.mutPublicKeys.Unlock()

	return nil
}

func (signer *fastConsensusSigner) publicKeyBytes(privateKey crypto.PrivateKey) ([]byte, error) {
	if privateKey == nil || privateKey.IsInterfaceNil() {
		return nil, crypto.ErrNilPrivateKey
	}

	privateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	signer.mutPublicKeys.RLock()
	publicKeyBytes, ok := signer.publicKeyByPrivate[string(privateKeyBytes)]
	signer.mutPublicKeys.RUnlock()
	if ok {
		return append([]byte(nil), publicKeyBytes...), nil
	}

	publicKey := privateKey.GeneratePublic()
	if publicKey == nil || publicKey.IsInterfaceNil() {
		return nil, errUnknownFastConsensusPrivateKey
	}
	err = signer.rememberKeyPair(privateKey, publicKey)
	if err != nil {
		return nil, err
	}

	return publicKey.ToByteArray()
}

func (signer *fastConsensusSigner) publicKeyBytesFromPrivateBytes(privateKeyBytes []byte) ([]byte, error) {
	signer.mutPublicKeys.RLock()
	publicKeyBytes, ok := signer.publicKeyByPrivate[string(privateKeyBytes)]
	signer.mutPublicKeys.RUnlock()
	if ok {
		return append([]byte(nil), publicKeyBytes...), nil
	}

	if signer.keyGenerator == nil || signer.keyGenerator.IsInterfaceNil() {
		return nil, errUnknownFastConsensusPrivateKey
	}
	privateKey, err := signer.keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return signer.publicKeyBytes(privateKey)
}

func (signer *fastConsensusSigner) Sign(privateKey crypto.PrivateKey, message []byte) ([]byte, error) {
	publicKeyBytes, err := signer.publicKeyBytes(privateKey)
	if err != nil {
		return nil, err
	}

	return computeFastConsensusSignature(publicKeyBytes, message), nil
}

func (signer *fastConsensusSigner) Verify(
	publicKey crypto.PublicKey,
	message []byte,
	signature []byte,
) error {
	if publicKey == nil || publicKey.IsInterfaceNil() {
		return crypto.ErrEmptyPubKey
	}
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return err
	}

	return verifyFastConsensusSignature(publicKeyBytes, message, signature)
}

func (signer *fastConsensusSigner) CreateSignatureShare(
	privateKeyBytes []byte,
	message []byte,
) ([]byte, error) {
	publicKeyBytes, err := signer.publicKeyBytesFromPrivateBytes(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return computeFastConsensusSignature(publicKeyBytes, message), nil
}

func (signer *fastConsensusSigner) CreateSignatureShareV2(
	privateKey crypto.PrivateKey,
	message []byte,
) ([]byte, error) {
	return signer.Sign(privateKey, message)
}

func (signer *fastConsensusSigner) VerifySignatureShare(
	publicKey []byte,
	message []byte,
	signature []byte,
) error {
	return verifyFastConsensusSignature(publicKey, message, signature)
}

func (signer *fastConsensusSigner) VerifySignatureShareV2(
	publicKey crypto.PublicKey,
	message []byte,
	signature []byte,
) error {
	return signer.Verify(publicKey, message, signature)
}

func (signer *fastConsensusSigner) AggregateSigs(
	publicKeys [][]byte,
	signatures [][]byte,
) ([]byte, error) {
	return computeFastConsensusAggregate(publicKeys, signatures)
}

func (signer *fastConsensusSigner) AggregateSigsV2(
	publicKeys []crypto.PublicKey,
	signatures [][]byte,
) ([]byte, error) {
	publicKeyBytes, err := publicKeysToBytes(publicKeys)
	if err != nil {
		return nil, err
	}

	return computeFastConsensusAggregate(publicKeyBytes, signatures)
}

func (signer *fastConsensusSigner) VerifyAggregatedSig(
	publicKeys [][]byte,
	message []byte,
	aggregatedSignature []byte,
) error {
	return verifyFastConsensusAggregate(publicKeys, message, aggregatedSignature)
}

func (signer *fastConsensusSigner) VerifyAggregatedSigV2(
	publicKeys []crypto.PublicKey,
	message []byte,
	aggregatedSignature []byte,
) error {
	publicKeyBytes, err := publicKeysToBytes(publicKeys)
	if err != nil {
		return err
	}

	return verifyFastConsensusAggregate(publicKeyBytes, message, aggregatedSignature)
}

func (signer *fastConsensusSigner) IsInterfaceNil() bool {
	return signer == nil
}

type fastConsensusMultiSignerContainer struct {
	signer crypto.MultiSignerV2
}

func (container *fastConsensusMultiSignerContainer) GetMultiSigner(_ uint32) (crypto.MultiSignerV2, error) {
	if container == nil || container.signer == nil || container.signer.IsInterfaceNil() {
		return nil, mxChainErrors.ErrNilMultiSigner
	}

	return container.signer, nil
}

func (container *fastConsensusMultiSignerContainer) IsInterfaceNil() bool {
	return container == nil
}

func computeFastConsensusSignature(publicKey []byte, message []byte) []byte {
	hasher := sha512.New384()
	writeFastConsensusField(hasher, []byte("mx-chain-simulator/fast-consensus/signature/v1"))
	writeFastConsensusField(hasher, publicKey)
	writeFastConsensusField(hasher, message)

	return hasher.Sum(nil)
}

func verifyFastConsensusSignature(publicKey []byte, message []byte, signature []byte) error {
	if len(publicKey) == 0 {
		return crypto.ErrInvalidPublicKey
	}
	if len(signature) != fastConsensusSignatureSize {
		return errInvalidFastConsensusSignature
	}

	expected := computeFastConsensusSignature(publicKey, message)
	if !bytes.Equal(expected, signature) {
		return errInvalidFastConsensusSignature
	}

	return nil
}

func computeFastConsensusAggregate(publicKeys [][]byte, signatures [][]byte) ([]byte, error) {
	if len(publicKeys) == 0 {
		return nil, crypto.ErrNilPublicKeys
	}
	if len(publicKeys) != len(signatures) {
		return nil, crypto.ErrInvalidParam
	}

	hasher := sha512.New384()
	writeFastConsensusField(hasher, []byte("mx-chain-simulator/fast-consensus/aggregate/v1"))
	for index := range publicKeys {
		if len(publicKeys[index]) == 0 {
			return nil, crypto.ErrInvalidPublicKey
		}
		if len(signatures[index]) != fastConsensusSignatureSize {
			return nil, errInvalidFastConsensusSignature
		}
		writeFastConsensusField(hasher, publicKeys[index])
		writeFastConsensusField(hasher, signatures[index])
	}

	return hasher.Sum(nil), nil
}

func verifyFastConsensusAggregate(
	publicKeys [][]byte,
	message []byte,
	aggregatedSignature []byte,
) error {
	signatures := make([][]byte, len(publicKeys))
	for index, publicKey := range publicKeys {
		signatures[index] = computeFastConsensusSignature(publicKey, message)
	}

	expected, err := computeFastConsensusAggregate(publicKeys, signatures)
	if err != nil {
		return err
	}
	if len(aggregatedSignature) != fastConsensusSignatureSize ||
		!bytes.Equal(expected, aggregatedSignature) {
		return errInvalidFastConsensusSignature
	}

	return nil
}

func publicKeysToBytes(publicKeys []crypto.PublicKey) ([][]byte, error) {
	if len(publicKeys) == 0 {
		return nil, crypto.ErrNilPublicKeys
	}

	result := make([][]byte, len(publicKeys))
	for index, publicKey := range publicKeys {
		if publicKey == nil || publicKey.IsInterfaceNil() {
			return nil, crypto.ErrEmptyPubKey
		}
		publicKeyBytes, err := publicKey.ToByteArray()
		if err != nil {
			return nil, err
		}
		result[index] = publicKeyBytes
	}

	return result, nil
}

func writeFastConsensusField(hasher interface{ Write([]byte) (int, error) }, value []byte) {
	var length [binary.MaxVarintLen64]byte
	numBytes := binary.PutUvarint(length[:], uint64(len(value)))
	_, _ = hasher.Write(length[:numBytes])
	_, _ = hasher.Write(value)
}
