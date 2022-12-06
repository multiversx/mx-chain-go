package signing

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	cryptoCommon "github.com/ElrondNetwork/elrond-go/common/crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
)

// ArgsSignatureHolder defines the arguments needed to create a new signature holder component
type ArgsSignatureHolder struct {
	PubKeys              []string
	MultiSignerContainer cryptoCommon.MultiSignerContainer
	SingleSigner         crypto.SingleSigner
	KeyGenerator         crypto.KeyGenerator
	KeysHandler          consensus.KeysHandler
}

type signatureHolderData struct {
	pubKeys   [][]byte
	sigShares [][]byte
	aggSig    []byte
}

type signatureHolder struct {
	data                 *signatureHolderData
	mutSigningData       sync.RWMutex
	multiSignerContainer cryptoCommon.MultiSignerContainer
	singleSigner         crypto.SingleSigner
	keyGen               crypto.KeyGenerator
	keysHandler          consensus.KeysHandler
}

// NewSignatureHolder will create a new signature holder component
func NewSignatureHolder(args ArgsSignatureHolder) (*signatureHolder, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	sigSharesSize := uint16(len(args.PubKeys))
	sigShares := make([][]byte, sigSharesSize)

	pubKeysBytes, err := convertStringsToPubKeysBytes(args.PubKeys)
	if err != nil {
		return nil, err
	}

	data := &signatureHolderData{
		pubKeys:   pubKeysBytes,
		sigShares: sigShares,
	}

	return &signatureHolder{
		data:                 data,
		mutSigningData:       sync.RWMutex{},
		multiSignerContainer: args.MultiSignerContainer,
		singleSigner:         args.SingleSigner,
		keyGen:               args.KeyGenerator,
		keysHandler:          args.KeysHandler,
	}, nil
}

func checkArgs(args ArgsSignatureHolder) error {
	if check.IfNil(args.MultiSignerContainer) {
		return ErrNilMultiSignerContainer
	}
	if check.IfNil(args.SingleSigner) {
		return ErrNilSingleSigner
	}
	if check.IfNil(args.KeysHandler) {
		return ErrNilKeysHandler
	}
	if check.IfNil(args.KeyGenerator) {
		return ErrNilKeyGenerator
	}
	if len(args.PubKeys) == 0 {
		return ErrNoPublicKeySet
	}

	return nil
}

// Create generates a signature holder component and initializes corresponding fields
func (sh *signatureHolder) Create(pubKeys []string) (*signatureHolder, error) {
	args := ArgsSignatureHolder{
		PubKeys:              pubKeys,
		KeysHandler:          sh.keysHandler,
		MultiSignerContainer: sh.multiSignerContainer,
		SingleSigner:         sh.singleSigner,
		KeyGenerator:         sh.keyGen,
	}
	return NewSignatureHolder(args)
}

// Reset resets the data inside the signature holder component
func (sh *signatureHolder) Reset(pubKeys []string) error {
	if pubKeys == nil {
		return ErrNilPublicKeys
	}

	sigSharesSize := uint16(len(pubKeys))
	sigShares := make([][]byte, sigSharesSize)
	pubKeysBytes, err := convertStringsToPubKeysBytes(pubKeys)
	if err != nil {
		return err
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	data := &signatureHolderData{
		pubKeys:   pubKeysBytes,
		sigShares: sigShares,
	}

	sh.data = data

	return nil
}

// CreateSignatureShareForPublicKey returns a signature over a message using the managed private key that was selected based on the provided
// publicKeyBytes argument
func (sh *signatureHolder) CreateSignatureShareForPublicKey(message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
	if message == nil {
		return nil, ErrNilMessage
	}

	privateKey := sh.keysHandler.GetHandledPrivateKey(publicKeyBytes)
	privateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	multiSigner, err := sh.multiSignerContainer.GetMultiSigner(epoch)
	if err != nil {
		return nil, err
	}

	sigShareBytes, err := multiSigner.CreateSignatureShare(privateKeyBytes, message)
	if err != nil {
		return nil, err
	}

	sh.data.sigShares[index] = sigShareBytes

	return sigShareBytes, nil
}

// CreateSignatureForPublicKey returns a signature over a message using the managed private key that was selected based on the provided
// publicKeyBytes argument
func (sh *signatureHolder) CreateSignatureForPublicKey(message []byte, publicKeyBytes []byte) ([]byte, error) {
	privateKey := sh.keysHandler.GetHandledPrivateKey(publicKeyBytes)

	return sh.singleSigner.Sign(privateKey, message)
}

// VerifySignatureShare will verify the signature share based on the specified index
func (sh *signatureHolder) VerifySignatureShare(index uint16, sig []byte, message []byte, epoch uint32) error {
	if len(sig) == 0 {
		return ErrInvalidSignature
	}

	sh.mutSigningData.RLock()
	defer sh.mutSigningData.RUnlock()

	indexOutOfBounds := index >= uint16(len(sh.data.pubKeys))
	if indexOutOfBounds {
		return ErrIndexOutOfBounds
	}

	pubKey := sh.data.pubKeys[index]

	multiSigner, err := sh.multiSignerContainer.GetMultiSigner(epoch)
	if err != nil {
		return err
	}

	return multiSigner.VerifySignatureShare(pubKey, message, sig)
}

// StoreSignatureShare stores the partial signature of the signer with specified position
func (sh *signatureHolder) StoreSignatureShare(index uint16, sig []byte) error {
	if len(sig) == 0 {
		return ErrInvalidSignature
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	if int(index) >= len(sh.data.sigShares) {
		return ErrIndexOutOfBounds
	}

	sh.data.sigShares[index] = sig

	return nil
}

// SignatureShare returns the partial signature set for given index
func (sh *signatureHolder) SignatureShare(index uint16) ([]byte, error) {
	sh.mutSigningData.RLock()
	defer sh.mutSigningData.RUnlock()

	if int(index) >= len(sh.data.sigShares) {
		return nil, ErrIndexOutOfBounds
	}

	if sh.data.sigShares[index] == nil {
		return nil, ErrNilElement
	}

	return sh.data.sigShares[index], nil
}

// not concurrent safe, should be used under RLock mutex
func (sh *signatureHolder) isIndexInBitmap(index uint16, bitmap []byte) bool {
	indexOutOfBounds := index >= uint16(len(sh.data.pubKeys))
	if indexOutOfBounds {
		return false
	}

	indexNotInBitmap := bitmap[index/8]&(1<<uint8(index%8)) == 0

	return !indexNotInBitmap
}

// AggregateSigs aggregates all collected partial signatures
func (sh *signatureHolder) AggregateSigs(bitmap []byte, epoch uint32) ([]byte, error) {
	if bitmap == nil {
		return nil, ErrNilBitmap
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(sh.data.pubKeys)
	if flagsMismatch {
		return nil, ErrBitmapMismatch
	}

	multiSigner, err := sh.multiSignerContainer.GetMultiSigner(epoch)
	if err != nil {
		return nil, err
	}

	signatures := make([][]byte, 0, len(sh.data.sigShares))
	pubKeysSigners := make([][]byte, 0, len(sh.data.sigShares))

	for i := range sh.data.sigShares {
		if !sh.isIndexInBitmap(uint16(i), bitmap) {
			continue
		}

		signatures = append(signatures, sh.data.sigShares[i])
		pubKeysSigners = append(pubKeysSigners, sh.data.pubKeys[i])
	}

	return multiSigner.AggregateSigs(pubKeysSigners, signatures)
}

// SetAggregatedSig sets the aggregated signature
func (sh *signatureHolder) SetAggregatedSig(aggSig []byte) error {
	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	sh.data.aggSig = aggSig

	return nil
}

// Verify verifies the aggregated signature by checking that aggregated signature is valid with respect
// to aggregated public keys.
func (sh *signatureHolder) Verify(message []byte, bitmap []byte, epoch uint32) error {
	if bitmap == nil {
		return ErrNilBitmap
	}

	sh.mutSigningData.RLock()
	defer sh.mutSigningData.RUnlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(sh.data.pubKeys)
	if flagsMismatch {
		return ErrBitmapMismatch
	}

	multiSigner, err := sh.multiSignerContainer.GetMultiSigner(epoch)
	if err != nil {
		return err
	}

	pubKeys := make([][]byte, 0, len(sh.data.pubKeys))
	for i, pk := range sh.data.pubKeys {
		if !sh.isIndexInBitmap(uint16(i), bitmap) {
			continue
		}

		pubKeys = append(pubKeys, pk)
	}

	return multiSigner.VerifyAggregatedSig(pubKeys, message, sh.data.aggSig)
}

func convertStringsToPubKeysBytes(pubKeys []string) ([][]byte, error) {
	pk := make([][]byte, 0, len(pubKeys))

	for _, pubKeyStr := range pubKeys {
		if pubKeyStr == "" {
			return nil, ErrEmptyPubKeyString
		}

		pubKey := []byte(pubKeyStr)
		pk = append(pk, pubKey)
	}

	return pk, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sh *signatureHolder) IsInterfaceNil() bool {
	return sh == nil
}
