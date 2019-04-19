package multisig

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
)

/*
This implementation follows the modified BLS scheme presented here (curve notation changed in this file as compared to
the link, so curves G0, G1 in link are refered to as G1, G2 in this file and in kyber library):
https://crypto.stanford.edu/~dabo/pubs/papers/BLSmultisig.html

In addition to the common BLS single signature, for aggregation of multiple signatures it requires another hashing
function H1, that translates from public keys (points on G2) to scalars H1: G2^n -> R^n

This extra hashing function is used only for the aggregation of standard single BLS signatures and to verify the
aggregated signature.

Even though standard BLS allows aggregation as well, it is susceptible to rogue key attacks.
This is where the modified BLS scheme comes into play and prevents this attacks by using this extra hashing function.
*/

type blsMultiSigData struct {
	message []byte
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
	suite      crypto.Suite
	hasher     hashing.Hasher
	keyGen     crypto.KeyGenerator
}

// NewBLSMultisig creates a new BLS multi-signer
func NewBLSMultisig(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	keyGen crypto.KeyGenerator,
	ownIndex uint16) (*blsMultiSigner, error) {

	if hasher == nil {
		return nil, crypto.ErrNilHasher
	}

	if privKey == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	if pubKeys == nil {
		return nil, crypto.ErrNilPublicKeys
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

	sizeConsensus := len(pubKeys)
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
		suite:      keyGen.Suite(),
	}, nil
}

// Reset resets the multiSigData inside the multiSigner
func (bms *blsMultiSigner) Reset(pubKeys []string, index uint16) error {
	if index >= uint16(len(pubKeys)) {
		return crypto.ErrIndexOutOfBounds
	}

	sizeConsensus := len(pubKeys)
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

	return NewBelNevMultisig(bms.hasher, pubKeys, privKey, bms.keyGen, index)
}

// SetMessage sets the message to be multi-signed upon
func (bms *blsMultiSigner) SetMessage(msg []byte) error {
	if msg == nil {
		return crypto.ErrNilMessage
	}

	if len(msg) == 0 {
		return crypto.ErrInvalidParam
	}

	bms.mutSigData.Lock()
	bms.data.message = msg
	bms.mutSigData.Unlock()

	return nil
}

// CreateSignatureShare returns a BLS single signature over the message previously configured with a previous call of
// SetMessage
func (bms *blsMultiSigner) CreateSignatureShare(bitmap []byte) ([]byte, error) {
	if bitmap == nil {
		return nil, crypto.ErrNilBitmap
	}

	bms.mutSigData.Lock()
	defer bms.mutSigData.Unlock()

	maxFlags := len(bitmap) * 8
	data := bms.data
	flagsMismatch := maxFlags < len(data.pubKeys)
	if flagsMismatch {
		return nil, crypto.ErrBitmapMismatch
	}

	blsSingleSigner := &singlesig.BlsSingleSigner{}
	sigShareBytes, err := blsSingleSigner.Sign(data.privKey, data.message)
	if err != nil {
		return nil, err
	}

	data.sigShares[data.ownIndex] = sigShareBytes

	return sigShareBytes, nil
}

// not concurrent safe, should be used under RLock mutex
func (bms *blsMultiSigner) isValidIndex(index uint16, bitmap []byte) error {
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

// VerifySignatureShare verifies the single signature share of the signer with specified position in bitmap
// Signature is verified over a message configured with a previous call of SetMessage
func (bms *blsMultiSigner) VerifySignatureShare(index uint16, sig []byte, bitmap []byte) error {
	if sig == nil {
		return crypto.ErrNilSignature
	}

	if bitmap == nil {
		return crypto.ErrNilBitmap
	}

	bms.mutSigData.RLock()
	defer bms.mutSigData.RUnlock()

	err := bms.isValidIndex(index, bitmap)
	if err != nil {
		return err
	}

	pubKey := bms.data.pubKeys[index]
	blsSingleSigner := &singlesig.BlsSingleSigner{}

	return blsSingleSigner.Verify(pubKey, bms.data.message, sig)
}

// StoreSignatureShare stores the partial signature of the signer with specified position
func (bms *blsMultiSigner) StoreSignatureShare(index uint16, sig []byte) error {
	if sig == nil {
		return crypto.ErrNilSignature
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
		err := bms.isValidIndex(uint16(i), bitmap)
		if err != nil {
			continue
		}

		hScalar, err := hashPublicKeyPointToScalar(bms.suite, bms.hasher, bms.data.pubKeys[i].Point())
		if err != nil {
			return nil, err
		}

		// H1(pubKey_i)*sig_i
		s, err := scalarMulSig(bms.suite, hScalar, bms.data.sigShares[i])
		if err != nil {
			return nil, err
		}

		prepSigs = append(prepSigs, s)
	}

	aggSigs, err := aggregateSignatures(bms.suite, prepSigs...)
	if err != nil {
		return nil, err
	}

	bms.data.aggSig = aggSigs

	return aggSigs, nil
}

// SetAggregatedSig sets the aggregated signature
func (bms *blsMultiSigner) SetAggregatedSig(aggSig []byte) error {
	if aggSig == nil {
		return crypto.ErrNilSignature
	}

	bms.mutSigData.Lock()
	bms.data.aggSig = aggSig
	bms.mutSigData.Unlock()

	return nil
}

// Verify verifies the aggregated signature by checking that aggregated signature is valid with respect
// to aggregated public keys.
func (bms *blsMultiSigner) Verify(bitmap []byte) error {
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

	prepPubKeysPoints := make([]crypto.Point, 0)

	for i := range bms.data.pubKeys {
		err := bms.isValidIndex(uint16(i), bitmap)
		if err != nil {
			continue
		}

		pubKeyPoint := bms.data.pubKeys[i].Point()

		// t_i = H(pubKey_i)
		hScalar, err := hashPublicKeyPointToScalar(bms.suite, bms.hasher, pubKeyPoint)
		if err != nil {
			return err
		}

		// t_i*pubKey_i
		prepPoint, err := pubKeyPoint.Mul(hScalar)
		if err != nil {
			return err
		}

		prepPubKeysPoints = append(prepPubKeysPoints, prepPoint)
	}

	aggPointsBytes, err := aggregatePublicKeys(bms.suite, prepPubKeysPoints...)
	if err != nil {
		return err
	}

	return verifyAggregatedSig(bms.suite, aggPointsBytes, bms.data.aggSig, bms.data.message)
}

// aggregateSignatures produces an aggregation of single BLS signatures
func aggregateSignatures(suite crypto.Suite, sigs ...[]byte) ([]byte, error) {
	if suite == nil {
		return nil, crypto.ErrNilSuite
	}

	if sigs == nil {
		return nil, crypto.ErrNilSignaturesList
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	return bls.AggregateSignatures(kSuite, sigs...)
}

func verifyAggregatedSig(suite crypto.Suite, aggPointsBytes []byte, aggSigBytes []byte, msg []byte) error {
	if suite == nil {
		return crypto.ErrNilSuite
	}

	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return crypto.ErrInvalidSuite
	}

	aggKPoint := kSuite.G2().Point()
	err := aggKPoint.UnmarshalBinary(aggPointsBytes)
	if err != nil {
		return err
	}

	return bls.Verify(kSuite, aggKPoint, msg, aggSigBytes)
}

// aggregatePublicKeys produces an aggregation of BLS public keys
func aggregatePublicKeys(suite crypto.Suite, pubKeys ...crypto.Point) ([]byte, error) {
	if pubKeys == nil {
		return nil, crypto.ErrNilPublicKeys
	}

	kSuite, ok := suite.(pairing.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	kyberPoints := make([]kyber.Point, len(pubKeys))
	for i, pubKey := range pubKeys {
		if pubKey == nil {
			return nil, crypto.ErrNilPublicKeyPoint
		}

		kyberPoints[i], ok = pubKey.GetUnderlyingObj().(kyber.Point)
		if !ok {
			return nil, crypto.ErrInvalidPublicKey
		}
	}

	kyberAggPubKey := bls.AggregatePublicKeys(kSuite, kyberPoints...)

	return kyberAggPubKey.MarshalBinary()
}

// scalarMulSig returns the result of multiplying a BLS scalar with a BLS single signature
func scalarMulSig(suite crypto.Suite, scalar crypto.Scalar, sig []byte) ([]byte, error) {
	kSuite, ok := suite.GetUnderlyingSuite().(pairing.Suite)
	if !ok {
		return nil, crypto.ErrInvalidSuite
	}

	kscalar, ok := scalar.GetUnderlyingObj().(kyber.Scalar)
	if !ok {
		return nil, crypto.ErrInvalidScalar
	}

	sigKPoint := kSuite.G1().Point()
	err := sigKPoint.UnmarshalBinary(sig)
	if err != nil {
		return nil, err
	}

	resPoint := sigKPoint.Mul(kscalar, nil)
	resBytes, err := resPoint.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return resBytes, nil
}

// hashPublicKeyPointToScalar
func hashPublicKeyPointToScalar(
	suite crypto.Suite,
	hasher hashing.Hasher,
	pubKeyPoint crypto.Point,
) (crypto.Scalar, error) {
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
	hScalar := suite.CreateScalar()
	err = hScalar.UnmarshalBinary(h)
	if err != nil {
		return nil, err
	}

	return hScalar, nil
}
