package multisig

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

/*
This implementation follows the modified BLS scheme presented here (curve notation changed in this file as compared to
the link, so curves G0, G1 in link are refered to as G1, G2 in this file):
https://crypto.stanford.edu/~dabo/pubs/papers/BLSmultisig.html

In addition to the common BLS single signature, for aggregation of multiple signatures it requires another hashing
function H1, that translates from public keys (points on G2) to scalars H1: G2^n -> R^n

This extra hashing function is used only for the aggregation of standard single BLS signatures and to verify the
aggregated signature.

Even though standard BLS allows aggregation as well, it is susceptible to rogue key attacks.
This is where the modified BLS scheme comes into play and prevents this attacks by using this extra hashing function.
*/

const hasherOutputSize = 16

type blsMultiSigData struct {
	pubKeys []crypto.PublicKey
	privKey crypto.PrivateKey
	// signatures in BLS are points on curve G1
	sigShares [][]byte
	aggSig    []byte
	ownIndex  uint16
}

type blsMultiSigner struct {
	data       *blsMultiSigData
	mutSigData sync.RWMutex
	hasher     hashing.Hasher // 16bytes output hasher!
	keyGen     crypto.KeyGenerator
	llSigner   crypto.LowLevelSignerBLS
}

// NewBLSMultisig creates a new BLS multi-signer
func NewBLSMultisig(
	llSigner crypto.LowLevelSignerBLS,
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	keyGen crypto.KeyGenerator,
	ownIndex uint16) (*blsMultiSigner, error) {

	if hasher == nil {
		return nil, crypto.ErrNilHasher
	}

	if hasher.Size() != hasherOutputSize {
		return nil, crypto.ErrWrongSizeHasher
	}

	if privKey == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	if len(pubKeys) == 0 {
		return nil, crypto.ErrNoPublicKeySet
	}

	if keyGen == nil {
		return nil, crypto.ErrNilKeyGenerator
	}

	if ownIndex >= uint16(len(pubKeys)) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	sizeConsensus := uint16(len(pubKeys))
	sigShares := make([][]byte, sizeConsensus)
	pk, err := convertStringsToPubKeys(pubKeys, keyGen)

	if err != nil {
		return nil, err
	}

	data := &blsMultiSigData{
		pubKeys:   pk,
		privKey:   privKey,
		ownIndex:  ownIndex,
		sigShares: sigShares,
	}

	// own index is used only for signing
	return &blsMultiSigner{
		data:       data,
		mutSigData: sync.RWMutex{},
		hasher:     hasher,
		keyGen:     keyGen,
		llSigner:   llSigner,
	}, nil
}

// Reset resets the multiSigData inside the multiSigner
func (bms *blsMultiSigner) Reset(pubKeys []string, index uint16) error {
	if pubKeys == nil {
		return crypto.ErrNilPublicKeys
	}

	if index >= uint16(len(pubKeys)) {
		return crypto.ErrIndexOutOfBounds
	}

	sizeConsensus := uint16(len(pubKeys))
	sigShares := make([][]byte, sizeConsensus)
	pk, err := convertStringsToPubKeys(pubKeys, bms.keyGen)

	if err != nil {
		return err
	}

	bms.mutSigData.Lock()
	defer bms.mutSigData.Unlock()

	privKey := bms.data.privKey

	data := &blsMultiSigData{
		pubKeys:   pk,
		privKey:   privKey,
		ownIndex:  index,
		sigShares: sigShares,
	}

	bms.data = data

	return nil
}

// Create generates a multiSigner and initializes corresponding fields with the given params
func (bms *blsMultiSigner) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	bms.mutSigData.RLock()
	privKey := bms.data.privKey
	bms.mutSigData.RUnlock()

	return NewBLSMultisig(bms.llSigner, bms.hasher, pubKeys, privKey, bms.keyGen, index)
}

// CreateSignatureShare returns a BLS single signature over the message
func (bms *blsMultiSigner) CreateSignatureShare(message []byte, _ []byte) ([]byte, error) {
	bms.mutSigData.Lock()
	defer bms.mutSigData.Unlock()

	data := bms.data
	sigShareBytes, err := bms.llSigner.SignShare(data.privKey, message)
	if err != nil {
		return nil, err
	}

	data.sigShares[data.ownIndex] = sigShareBytes

	return sigShareBytes, nil
}

// not concurrent safe, should be used under RLock mutex
func (bms *blsMultiSigner) isIndexInBitmap(index uint16, bitmap []byte) error {
	indexOutOfBounds := index >= uint16(len(bms.data.pubKeys))
	if indexOutOfBounds {
		return crypto.ErrIndexOutOfBounds
	}

	indexNotInBitmap := bitmap[index/8]&(1<<uint8(index%8)) == 0
	if indexNotInBitmap {
		return crypto.ErrIndexNotSelected
	}

	return nil
}

// VerifySignatureShare verifies the single signature share of the signer with specified position
// Signature is verified over a message configured with a previous call of SetMessage
func (bms *blsMultiSigner) VerifySignatureShare(index uint16, sig []byte, message []byte, _ []byte) error {
	if sig == nil {
		return crypto.ErrNilSignature
	}

	bms.mutSigData.RLock()
	defer bms.mutSigData.RUnlock()

	indexOutOfBounds := index >= uint16(len(bms.data.pubKeys))
	if indexOutOfBounds {
		return crypto.ErrIndexOutOfBounds
	}

	pubKey := bms.data.pubKeys[index]

	return bms.llSigner.VerifySigShare(pubKey, message, sig)
}

// StoreSignatureShare stores the partial signature of the signer with specified position
// Function does not validate the signature, as it expects caller to have already called VerifySignatureShare
func (bms *blsMultiSigner) StoreSignatureShare(index uint16, sig []byte) error {
	err := bms.llSigner.VerifySigBytes(bms.keyGen.Suite(), sig)
	if err != nil {
		return err
	}

	bms.mutSigData.Lock()
	defer bms.mutSigData.Unlock()

	if int(index) >= len(bms.data.sigShares) {
		return crypto.ErrIndexOutOfBounds
	}

	bms.data.sigShares[index] = sig

	return nil
}

// SignatureShare returns the partial signature set for given index
func (bms *blsMultiSigner) SignatureShare(index uint16) ([]byte, error) {
	bms.mutSigData.RLock()
	defer bms.mutSigData.RUnlock()

	if int(index) >= len(bms.data.sigShares) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	if bms.data.sigShares[index] == nil {
		return nil, crypto.ErrNilElement
	}

	return bms.data.sigShares[index], nil
}

// AggregateSigs aggregates all collected partial signatures
func (bms *blsMultiSigner) AggregateSigs(bitmap []byte) ([]byte, error) {
	if bitmap == nil {
		return nil, crypto.ErrNilBitmap
	}

	bms.mutSigData.Lock()
	defer bms.mutSigData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(bms.data.pubKeys)
	if flagsMismatch {
		return nil, crypto.ErrBitmapMismatch
	}

	prepSigs := make([][]byte, 0)
	// for the modified BLS scheme, aggregation is done not between sigs but between H1(pubKey_i)*sig_i
	for i := range bms.data.sigShares {
		err := bms.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		hPk, err := hashPublicKeyPoint(bms.hasher, bms.data.pubKeys[i].Point())
		if err != nil {
			return nil, err
		}

		// H1(pubKey_i)*sig_i
		s, err := scalarMulSig(bms.llSigner, bms.keyGen.Suite(), hPk, bms.data.sigShares[i])
		if err != nil {
			return nil, err
		}

		prepSigs = append(prepSigs, s)
	}

	if len(prepSigs) == 0 {
		return nil, crypto.ErrNilSignaturesList
	}

	aggSigs, err := bms.llSigner.AggregateSignatures(bms.keyGen.Suite(), prepSigs...)
	if err != nil {
		return nil, err
	}

	bms.data.aggSig = aggSigs

	return aggSigs, nil
}

// SetAggregatedSig sets the aggregated signature
func (bms *blsMultiSigner) SetAggregatedSig(aggSig []byte) error {
	err := bms.llSigner.VerifySigBytes(bms.keyGen.Suite(), aggSig)
	if err != nil {
		return err
	}

	bms.mutSigData.Lock()
	bms.data.aggSig = aggSig
	bms.mutSigData.Unlock()

	return nil
}

// Verify verifies the aggregated signature by checking that aggregated signature is valid with respect
// to aggregated public keys.
func (bms *blsMultiSigner) Verify(message []byte, bitmap []byte) error {
	if bitmap == nil {
		return crypto.ErrNilBitmap
	}

	bms.mutSigData.RLock()
	defer bms.mutSigData.RUnlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(bms.data.pubKeys)
	if flagsMismatch {
		return crypto.ErrBitmapMismatch
	}

	pubKeysPoints := make([]crypto.Point, 0)

	for i := range bms.data.pubKeys {
		err := bms.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		pubKeysPoints = append(pubKeysPoints, bms.data.pubKeys[i].Point())
	}

	aggPointsBytes, err := aggregatePublicKeys(bms.llSigner, bms.keyGen.Suite(), pubKeysPoints, bms.hasher)
	if err != nil {
		return err
	}

	return bms.llSigner.VerifyAggregatedSig(bms.keyGen.Suite(), aggPointsBytes, bms.data.aggSig, message)
}

func aggregatePublicKeys(
	lls crypto.LowLevelSignerBLS,
	suite crypto.Suite,
	pubKeys []crypto.Point,
	hasher hashing.Hasher,
) ([]byte, error) {
	prepPubKeysPoints := make([]crypto.Point, 0)

	for i := range pubKeys {
		// t_i = H(pubKey_i)
		hPk, err := hashPublicKeyPoint(hasher, pubKeys[i])
		if err != nil {
			return nil, err
		}

		// t_i*pubKey_i
		prepPoint, err := scalarMulPk(suite, hPk, pubKeys[i])
		if err != nil {
			return nil, err
		}

		prepPubKeysPoints = append(prepPubKeysPoints, prepPoint)
	}

	aggPointsBytes, err := lls.AggregatePublicKeys(suite, prepPubKeysPoints...)

	return aggPointsBytes, err
}

// scalarMulPk returns the result of multiplying a scalar given as a bytes array, with a BLS public key (point)
func scalarMulPk(suite crypto.Suite, scalarBytes []byte, pk crypto.Point) (crypto.Point, error) {
	if pk == nil {
		return nil, crypto.ErrNilParam
	}

	kScalar, err := createScalar(suite, scalarBytes)
	if err != nil {
		return nil, err
	}

	pkPoint, err := pk.Mul(kScalar)

	return pkPoint, nil
}

// scalarMulSig returns the result of multiplying a scalar given as a bytes array, with a BLS single signature
func scalarMulSig(lls crypto.LowLevelSignerBLS, suite crypto.Suite, scalarBytes []byte, sig []byte) ([]byte, error) {
	scalar, err := createScalar(suite, scalarBytes)
	if err != nil {
		return nil, err
	}

	return lls.ScalarMulSig(suite, scalar, sig)
}

// hashPublicKeyPoint hashes a BLS public key (point) into a byte array (32 bytes length)
func hashPublicKeyPoint(hasher hashing.Hasher, pubKeyPoint crypto.Point) ([]byte, error) {
	if hasher == nil {
		return nil, crypto.ErrNilHasher
	}

	if pubKeyPoint == nil {
		return nil, crypto.ErrNilPublicKeyPoint
	}

	pointBytes, err := pubKeyPoint.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// H1(pubkey_i)
	h := hasher.Compute(string(pointBytes))
	// accepted length 32, copy the hasherOutputSize bytes and have rest 0
	h32 := make([]byte, 32)
	copy(h32[hasherOutputSize:], h)

	return h32, nil
}

// createScalar creates crypto.Scalar from a byte array
func createScalar(suite crypto.Suite, scalarBytes []byte) (crypto.Scalar, error) {
	if suite == nil {
		return nil, crypto.ErrNilSuite
	}

	scalar := suite.CreateScalar()
	err := scalar.UnmarshalBinary(scalarBytes)
	if err != nil {
		return nil, err
	}

	return scalar, nil
}
