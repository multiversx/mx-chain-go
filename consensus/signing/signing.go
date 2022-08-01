package signing

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	cryptoCommon "github.com/ElrondNetwork/elrond-go/common/crypto"
)

// ArgsSignatureHolder defines the arguments needed to create a new signature holder component
type ArgsSignatureHolder struct {
	PubKeys              []string
	PrivKey              crypto.PrivateKey
	MultiSignerContainer cryptoCommon.MultiSignerContainer
	KeyGenerator         crypto.KeyGenerator
}

type signatureHolderData struct {
	pubKeys   [][]byte
	privKey   crypto.PrivateKey
	sigShares [][]byte
	aggSig    []byte
}

type signatureHolder struct {
	data                 *signatureHolderData
	mutSigningData       sync.RWMutex
	multiSignerContainer cryptoCommon.MultiSignerContainer
	multiSigner          crypto.MultiSigner
	mutMultiSigner       sync.RWMutex
	keyGen               crypto.KeyGenerator
}

// NewSignatureHolder will create a new signature holder component
func NewSignatureHolder(args ArgsSignatureHolder) (*signatureHolder, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	sigSharesSize := uint16(len(args.PubKeys))
	sigShares := make([][]byte, sigSharesSize)
	pk, err := convertStringsToPubKeysBytes(args.PubKeys)
	if err != nil {
		return nil, err
	}

	data := &signatureHolderData{
		pubKeys:   pk,
		privKey:   args.PrivKey,
		sigShares: sigShares,
	}

	multiSigner, err := args.MultiSignerContainer.GetMultiSigner(0)
	if err != nil {
		return nil, err
	}

	return &signatureHolder{
		data:                 data,
		mutSigningData:       sync.RWMutex{},
		multiSignerContainer: args.MultiSignerContainer,
		multiSigner:          multiSigner,
		keyGen:               args.KeyGenerator,
	}, nil
}

func checkArgs(args ArgsSignatureHolder) error {
	if check.IfNil(args.MultiSignerContainer) {
		return ErrNilMultiSignerContainer
	}
	if check.IfNil(args.PrivKey) {
		return ErrNilPrivateKey
	}
	if check.IfNil(args.KeyGenerator) {
		return ErrNilKeyGenerator
	}
	if len(args.PubKeys) == 0 {
		return ErrNoPublicKeySet
	}

	return nil
}

// Create generated a signature holder component and initializes corresponding fields
func (sh *signatureHolder) Create(pubKeys []string, index uint16) (*signatureHolder, error) {
	sh.mutSigningData.RLock()
	privKey := sh.data.privKey
	sh.mutSigningData.RUnlock()

	args := ArgsSignatureHolder{
		PubKeys:              pubKeys,
		PrivKey:              privKey,
		MultiSignerContainer: sh.multiSignerContainer,
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
	pk, err := convertStringsToPubKeysBytes(pubKeys)
	if err != nil {
		return err
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	privKey := sh.data.privKey

	data := &signatureHolderData{
		pubKeys:   pk,
		privKey:   privKey,
		sigShares: sigShares,
	}

	sh.data = data

	return nil
}

func (sh *signatureHolder) SetMultiSignerByEpoch(epoch uint32) error {
	multiSigner, err := sh.multiSignerContainer.GetMultiSigner(epoch)
	if err != nil {
		return err
	}

	sh.mutMultiSigner.Lock()
	sh.multiSigner = multiSigner
	sh.mutMultiSigner.Unlock()

	return nil
}

// CreateSignatureShare returns a signature over a message
func (sh *signatureHolder) CreateSignatureShare(message []byte, selfIndex uint16) ([]byte, error) {
	if message == nil {
		return nil, ErrNilMessage
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	privKeyBytes, err := sh.data.privKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	sh.mutMultiSigner.RLock()
	sigShareBytes, err := sh.multiSigner.CreateSignatureShare(privKeyBytes, message)
	if err != nil {
		return nil, err
	}
	sh.mutMultiSigner.RUnlock()

	sh.data.sigShares[selfIndex] = sigShareBytes

	return sigShareBytes, nil
}

// VerifySignatureShare will verify the signature share based on the specified index
func (sh *signatureHolder) VerifySignatureShare(index uint16, sig []byte, message []byte) error {
	if sig == nil {
		return ErrNilSignature
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	indexOutOfBounds := index >= uint16(len(sh.data.pubKeys))
	if indexOutOfBounds {
		return ErrIndexOutOfBounds
	}

	pubKey := sh.data.pubKeys[index]

	sh.mutMultiSigner.RLock()
	defer sh.mutMultiSigner.RUnlock()

	return sh.multiSigner.VerifySignatureShare(pubKey, message, sig)
}

// StoreSignatureShare stores the partial signature of the signer with specified position
func (sh *signatureHolder) StoreSignatureShare(index uint16, sig []byte) error {
	// TODO: evaluate verifying if sig bytes is a valid BLS signature
	if sig == nil {
		return ErrNilSignature
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
	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	if int(index) >= len(sh.data.sigShares) {
		return nil, ErrIndexOutOfBounds
	}

	if sh.data.sigShares[index] == nil {
		return nil, ErrNilElement
	}

	return sh.data.sigShares[index], nil
}

// not concurrent safe, should be used under RLock mutex
func (sh *signatureHolder) isIndexInBitmap(index uint16, bitmap []byte) error {
	indexOutOfBounds := index >= uint16(len(sh.data.pubKeys))
	if indexOutOfBounds {
		return ErrIndexOutOfBounds
	}

	indexNotInBitmap := bitmap[index/8]&(1<<uint8(index%8)) == 0
	if indexNotInBitmap {
		return ErrIndexNotSelected
	}

	return nil
}

// AggregateSigs aggregates all collected partial signatures
func (sh *signatureHolder) AggregateSigs(bitmap []byte) ([]byte, error) {
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

	signatures := make([][]byte, 0, len(sh.data.sigShares))
	pubKeysSigners := make([][]byte, 0, len(sh.data.sigShares))

	for i := range sh.data.sigShares {
		err := sh.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		signatures = append(signatures, sh.data.sigShares[i])
		pubKeysSigners = append(pubKeysSigners, sh.data.pubKeys[i])
	}

	sh.mutMultiSigner.RLock()
	defer sh.mutMultiSigner.RUnlock()

	return sh.multiSigner.AggregateSigs(pubKeysSigners, signatures)
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
func (sh *signatureHolder) Verify(message []byte, bitmap []byte) error {
	if bitmap == nil {
		return ErrNilBitmap
	}

	sh.mutSigningData.Lock()
	defer sh.mutSigningData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(sh.data.pubKeys)
	if flagsMismatch {
		return ErrBitmapMismatch
	}

	pubKeys := make([][]byte, 0)
	for i := range sh.data.pubKeys {
		err := sh.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		pubKeys = append(pubKeys, sh.data.pubKeys[i])
	}

	sh.mutMultiSigner.RLock()
	defer sh.mutMultiSigner.RUnlock()

	return sh.multiSigner.VerifyAggregatedSig(pubKeys, message, sh.data.aggSig)
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
