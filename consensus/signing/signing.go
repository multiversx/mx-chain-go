package signing

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// ArgsSinger defines the arguments needed to create a new signer
type ArgsSinger struct {
	PubKeys      []string
	OwnIndex     uint16
	PrivKey      crypto.PrivateKey
	SingleSigner crypto.SingleSigner
	MultiSigner  crypto.MultiSigner
	KeyGenerator crypto.KeyGenerator
}

type signingData struct {
	pubKeys   [][]byte
	privKey   crypto.PrivateKey
	sigShares [][]byte
	aggSig    []byte
	ownIndex  uint16
}

type signer struct {
	data           *signingData
	mutSigningData sync.RWMutex
	singleSigner   crypto.SingleSigner
	multiSigner    crypto.MultiSigner
	keyGen         crypto.KeyGenerator
}

// NewSigner will create a new signer component
func NewSigner(args ArgsSinger) (*signer, error) {
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

	data := &signingData{
		pubKeys:   pk,
		privKey:   args.PrivKey,
		sigShares: sigShares,
		ownIndex:  args.OwnIndex,
	}

	return &signer{
		data:           data,
		mutSigningData: sync.RWMutex{},
		singleSigner:   args.SingleSigner,
		multiSigner:    args.MultiSigner,
		keyGen:         args.KeyGenerator,
	}, nil
}

func checkArgs(args ArgsSinger) error {
	if check.IfNil(args.SingleSigner) {
		return ErrNilSingleSigner
	}
	if check.IfNil(args.MultiSigner) {
		return ErrNilMultiSigner
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
	if args.OwnIndex >= uint16(len(args.PubKeys)) {
		return ErrIndexOutOfBounds
	}

	return nil
}

// TODO: modify to use interface here
func (sg *signer) Create(pubKeys []string, index uint16) (*signer, error) {
	sg.mutSigningData.RLock()
	privKey := sg.data.privKey
	sg.mutSigningData.RUnlock()

	args := ArgsSinger{
		PubKeys:      pubKeys,
		PrivKey:      privKey,
		SingleSigner: sg.singleSigner,
		MultiSigner:  sg.multiSigner,
		KeyGenerator: sg.keyGen,
	}
	return NewSigner(args)
}

func (sg *signer) Reset(pubKeys []string, index uint16) error {
	if pubKeys == nil {
		return ErrNilPublicKeys
	}

	if index >= uint16(len(pubKeys)) {
		return ErrIndexOutOfBounds
	}

	sigSharesSize := uint16(len(pubKeys))
	sigShares := make([][]byte, sigSharesSize)
	pk, err := convertStringsToPubKeysBytes(pubKeys)
	if err != nil {
		return err
	}

	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	privKey := sg.data.privKey

	data := &signingData{
		pubKeys:   pk,
		privKey:   privKey,
		sigShares: sigShares,
		ownIndex:  index,
	}

	sg.data = data

	return nil
}

func (sg *signer) CreateSignatureShare(message []byte, _ []byte) ([]byte, error) {
	if message == nil {
		return nil, ErrNilMessage
	}

	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	privKeyBytes, err := sg.data.privKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	sigShareBytes, err := sg.multiSigner.CreateSignatureShare(privKeyBytes, message)
	if err != nil {
		return nil, err
	}

	sg.data.sigShares[sg.data.ownIndex] = sigShareBytes

	return sigShareBytes, nil
}

func (sg *signer) VerifySignatureShare(index uint16, sig []byte, message []byte, _ []byte) error {
	if sig == nil {
		return ErrNilSignature
	}

	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	indexOutOfBounds := index >= uint16(len(sg.data.pubKeys))
	if indexOutOfBounds {
		return ErrIndexOutOfBounds
	}

	pubKey := sg.data.pubKeys[index]

	return sg.multiSigner.VerifySignatureShare(pubKey, message, sig)
}

func (sg *signer) StoreSignatureShare(index uint16, sig []byte) error {
	// TODO: verify sig bytes

	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	if int(index) >= len(sg.data.sigShares) {
		return ErrIndexOutOfBounds
	}

	sg.data.sigShares[index] = sig

	return nil
}

func (sg *signer) SignatureShare(index uint16) ([]byte, error) {
	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	if int(index) >= len(sg.data.sigShares) {
		return nil, ErrIndexOutOfBounds
	}

	if sg.data.sigShares[index] == nil {
		return nil, ErrNilElement
	}

	return sg.data.sigShares[index], nil
}

// not concurrent safe, should be used under RLock mutex
func (sg *signer) isIndexInBitmap(index uint16, bitmap []byte) error {
	indexOutOfBounds := index >= uint16(len(sg.data.pubKeys))
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
func (sg *signer) AggregateSigs(bitmap []byte) ([]byte, error) {
	if bitmap == nil {
		return nil, ErrNilBitmap
	}

	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(sg.data.pubKeys)
	if flagsMismatch {
		return nil, ErrBitmapMismatch
	}

	signatures := make([][]byte, 0, len(sg.data.sigShares))
	pubKeysSigners := make([][]byte, 0, len(sg.data.sigShares))

	for i := range sg.data.sigShares {
		err := sg.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		signatures = append(signatures, sg.data.sigShares[i])
		pubKeysSigners = append(pubKeysSigners, sg.data.pubKeys[i])
	}

	return sg.multiSigner.AggregateSigs(pubKeysSigners, signatures)
}

// SetAggregatedSig sets the aggregated signature
func (sg *signer) SetAggregatedSig(aggSig []byte) error {
	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	sg.data.aggSig = aggSig

	return nil
}

// Verify verifies the aggregated signature by checking that aggregated signature is valid with respect
// to aggregated public keys.
func (sg *signer) Verify(message []byte, bitmap []byte) error {
	if bitmap == nil {
		return ErrNilBitmap
	}

	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(sg.data.pubKeys)
	if flagsMismatch {
		return ErrBitmapMismatch
	}

	pubKeys := make([][]byte, 0)
	for i := range sg.data.pubKeys {
		err := sg.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		pubKeys = append(pubKeys, sg.data.pubKeys[i])
	}

	return sg.multiSigner.VerifyAggregatedSig(pubKeys, message, sg.data.aggSig)
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
func (sg *signer) IsInterfaceNil() bool {
	return sg == nil
}
