package multisig

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

/*
belnev.go implements the multi-signature algorithm (BN-multiSig) presented in paper
"Multi-Signatures in the Plain Public-Key Model and a General Forking Lemma"
by Mihir Bellare and Gregory Neven. See https://cseweb.ucsd.edu/~mihir/papers/multisignatures-ccs.pdf.
This package provides the functionality for the cryptographic operations.
The message transfer functionality required for the algorithm are assumed to be
handled elsewhere. An overview of the protocol will be provided below.

The BN-multiSig protocol has 4 phases executed between a list of participants (public keys)
having a protocol leader (index = 0) and validators (index > 0). Each participant has it's
own private/public key pair (x_i, X_i), where x_i is the private key of participant
i and X_i is it's associated public key X_i = x_i * G. G is the base point on the used
curve, and * is the scalar multiplication. The protocol assumes that each participant has
the same view on the ordering of participants in the list L, so when communicating it is
not needed to pass on the list as well, but only a bitmap for the participating validators.

The phases of the protocol are as follows:

1. All participants/signers in L (including the leader) choose a random scalar 1 < r_i < n-1,
called a commitment secret, calculates the commitment R_i = r_i *G, and the commitment hash
t_i = H0(R_i), where H0 is a hashing function, different than H1 which will be used in next
rounds. Each of the members then broadcast these commitment hashes to all participants
including leader. (In a protocol where the communication is done only through the leader it needs
to be taken into account that the leader might not be honest)

2. When each signer i receives the commitment hash together with the public key of the sender
(t_j, X_j) it will send back the full R_i along with its public key, (R_i, X_i)

3. When signer i receives the full commitment from a signer j, it computes t_j = H0(R_j)
and verifies it with the previously received t_j. Locally, each participant keeps track of
the sender, using a bitmap initially set to 0, by setting the corresponding bit to 1 and
storing the received commitment. If the commitment fails to validate the hash, the protocol
is aborted. If commitment is not received in a bounded time delta and less than 2/3 of signers
have provided the commitment then protocol is aborted. If there are enough commitments >2/3
the leader broadcasts the bitmap and calculates the aggregated commitment R = Sum(R_i * B[i])

4. When signer i receives the bitmap, it calculates the signature share and broadcasts it to all
participating signers. R = Sum(R_i) is the aggregated commitment each signer needs to calculate
before signing. s_i = r_i + H1(<L'>||X_i||R||m) * x_i is the signature share for signer i
When signer i receives all signature shares, it can calculate the aggregated signature
s=Sum(s_i) for all s_i of the participants.

5. Verification is done by checking the equality:
	s * G = R + Sum(H1(<L'>||X_i||R||m) * X_i * B[i])
*/

type multiSigData struct {
	message       []byte
	pubKeys       []crypto.PublicKey
	privKey       crypto.PrivateKey
	commHashes    [][]byte
	commSecret    crypto.Scalar
	commitments   []crypto.Point
	aggCommitment crypto.Point
	sigShares     []crypto.Scalar
	aggSig        crypto.Scalar
	ownIndex      uint16
}

type belNevSigner struct {
	data       *multiSigData
	mutSigData sync.RWMutex
	hasher     hashing.Hasher
	keyGen     crypto.KeyGenerator
}

// NewBelNevMultisig creates a new Bellare Neven multi-signer
func NewBelNevMultisig(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	keyGen crypto.KeyGenerator,
	ownIndex uint16) (*belNevSigner, error) {

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
	commHashes := make([][]byte, sizeConsensus)
	commitments := make([]crypto.Point, sizeConsensus)
	sigShares := make([]crypto.Scalar, sizeConsensus)
	pk, err := convertStringsToPubKeys(pubKeys, keyGen)

	if err != nil {
		return nil, err
	}

	data := &multiSigData{
		pubKeys:     pk,
		privKey:     privKey,
		ownIndex:    ownIndex,
		commHashes:  commHashes,
		commitments: commitments,
		sigShares:   sigShares,
	}

	// own index is used only for signing
	return &belNevSigner{
		data:       data,
		mutSigData: sync.RWMutex{},
		hasher:     hasher,
		keyGen:     keyGen,
	}, nil
}

func convertStringsToPubKeys(pubKeys []string, kg crypto.KeyGenerator) ([]crypto.PublicKey, error) {
	var pk []crypto.PublicKey

	//convert pubKeys
	for _, pubKeyStr := range pubKeys {
		if pubKeyStr == "" {
			return nil, crypto.ErrEmptyPubKeyString
		}

		pubKey, err := kg.PublicKeyFromByteArray([]byte(pubKeyStr))
		if err != nil {
			return nil, crypto.ErrInvalidPublicKeyString
		}

		pk = append(pk, pubKey)
	}
	return pk, nil
}

// Reset resets the multiSigData inside the multiSigner
func (bn *belNevSigner) Reset(pubKeys []string, index uint16) error {
	if index >= uint16(len(pubKeys)) {
		return crypto.ErrIndexOutOfBounds
	}

	sizeConsensus := len(pubKeys)
	commHashes := make([][]byte, sizeConsensus)
	commitments := make([]crypto.Point, sizeConsensus)
	sigShares := make([]crypto.Scalar, sizeConsensus)
	pk, err := convertStringsToPubKeys(pubKeys, bn.keyGen)

	if err != nil {
		return err
	}

	bn.mutSigData.Lock()
	defer bn.mutSigData.Unlock()

	privKey := bn.data.privKey

	data := &multiSigData{
		pubKeys:     pk,
		privKey:     privKey,
		ownIndex:    index,
		commHashes:  commHashes,
		commitments: commitments,
		sigShares:   sigShares,
	}

	bn.data = data

	return nil
}

// Create generates a multiSigner and initializes corresponding fields with the given params
func (bn *belNevSigner) Create(pubKeys []string, index uint16) (crypto.MultiSigner, error) {
	bn.mutSigData.RLock()
	privKey := bn.data.privKey
	bn.mutSigData.RUnlock()

	return NewBelNevMultisig(bn.hasher, pubKeys, privKey, bn.keyGen, index)
}

// SetMessage sets the message to be multi-signed upon
func (bn *belNevSigner) SetMessage(msg []byte) error {
	if msg == nil {
		return crypto.ErrNilMessage
	}

	if len(msg) == 0 {
		return crypto.ErrInvalidParam
	}

	bn.mutSigData.Lock()
	bn.data.message = msg
	bn.mutSigData.Unlock()

	return nil
}

// StoreCommitmentHash sets a commitment Hash
func (bn *belNevSigner) StoreCommitmentHash(index uint16, commHash []byte) error {
	if commHash == nil {
		return crypto.ErrNilCommitmentHash
	}

	bn.mutSigData.Lock()
	if int(index) >= len(bn.data.commHashes) {
		bn.mutSigData.Unlock()
		return crypto.ErrIndexOutOfBounds
	}

	bn.data.commHashes[index] = commHash
	bn.mutSigData.Unlock()

	return nil
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bn *belNevSigner) CommitmentHash(index uint16) ([]byte, error) {
	bn.mutSigData.RLock()
	defer bn.mutSigData.RUnlock()

	if int(index) >= len(bn.data.commHashes) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	if bn.data.commHashes[index] == nil {
		return nil, crypto.ErrNilElement
	}

	return bn.data.commHashes[index], nil
}

// CreateCommitment creates a secret commitment and the corresponding public commitment point
func (bn *belNevSigner) CreateCommitment() (commSecret []byte, commitment []byte) {
	suite := bn.keyGen.Suite()
	rand := suite.RandomStream()
	sk, _ := suite.CreateScalar().Pick(rand)
	pk := suite.CreatePoint().Base()
	pk, _ = pk.Mul(sk)

	bn.mutSigData.Lock()
	bn.data.commSecret = sk
	bn.data.commitments[bn.data.ownIndex] = pk
	bn.mutSigData.Unlock()

	commSecret, _ = sk.MarshalBinary()
	commitment, _ = pk.MarshalBinary()

	return commSecret, commitment
}

// CommitmentSecret returns the set commitment secret
func (bn *belNevSigner) CommitmentSecret() ([]byte, error) {
	bn.mutSigData.RLock()
	if bn.data.commSecret == nil {
		bn.mutSigData.RUnlock()
		return nil, crypto.ErrNilCommitmentSecret
	}

	commSecret, err := bn.data.commSecret.MarshalBinary()
	bn.mutSigData.RUnlock()

	return commSecret, err
}

// StoreCommitment adds a commitment to the list on the specified position
func (bn *belNevSigner) StoreCommitment(index uint16, commitment []byte) error {
	if commitment == nil {
		return crypto.ErrNilCommitment
	}

	commPoint := bn.keyGen.Suite().CreatePoint()
	err := commPoint.UnmarshalBinary(commitment)

	if err != nil {
		return err
	}

	bn.mutSigData.Lock()
	if int(index) >= len(bn.data.commitments) {
		bn.mutSigData.Unlock()
		return crypto.ErrIndexOutOfBounds
	}

	bn.data.commitments[index] = commPoint
	bn.mutSigData.Unlock()
	return nil
}

// Commitment returns the commitment from the list with the specified position
func (bn *belNevSigner) Commitment(index uint16) ([]byte, error) {
	bn.mutSigData.RLock()
	defer bn.mutSigData.RUnlock()

	if int(index) >= len(bn.data.commitments) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	if bn.data.commitments[index] == nil {
		return nil, crypto.ErrNilElement
	}

	commArray, err := bn.data.commitments[index].MarshalBinary()

	if err != nil {
		return nil, err
	}

	return commArray, nil
}

// AggregateCommitments aggregates the list of commitments
func (bn *belNevSigner) AggregateCommitments(bitmap []byte) error {
	if bitmap == nil {
		return crypto.ErrNilBitmap
	}

	maxFlags := len(bitmap) * 8

	bn.mutSigData.Lock()
	defer bn.mutSigData.Unlock()

	flagsMismatch := maxFlags < len(bn.data.pubKeys)
	if flagsMismatch {
		return crypto.ErrBitmapMismatch
	}

	aggComm := bn.keyGen.Suite().CreatePoint().Null()

	for i := range bn.data.commitments {
		err := bn.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		aggComm, err = aggComm.Add(bn.data.commitments[i])
		if err != nil {
			return err
		}
	}

	bn.data.aggCommitment = aggComm

	return nil
}

// Creates the challenge for the specific index H1(<L'>||X_i||R||m)
// Not concurrent safe, should be used under RLock
func (bn *belNevSigner) computeChallenge(index uint16, bitmap []byte) (crypto.Scalar, error) {
	sizeConsensus := uint16(len(bn.data.commitments))

	if index >= sizeConsensus {
		return nil, crypto.ErrIndexOutOfBounds
	}

	if bn.data.message == nil {
		return nil, crypto.ErrNilMessage
	}

	if bitmap == nil {
		return nil, crypto.ErrNilBitmap
	}

	concatenated := make([]byte, 0)

	// Concatenate pubKeys to form <L'>
	for i := range bn.data.pubKeys {
		err := bn.isIndexInBitmap(uint16(i), bitmap)

		if err != nil {
			continue
		}

		pubKey, _ := bn.data.pubKeys[i].Point().MarshalBinary()
		concatenated = append(concatenated, pubKey...)
	}

	pubKey, err := bn.data.pubKeys[index].Point().MarshalBinary()
	if err != nil {
		return nil, err
	}

	if bn.data.aggCommitment == nil {
		return nil, crypto.ErrNilAggregatedCommitment
	}

	aggCommBytes, _ := bn.data.aggCommitment.MarshalBinary()

	// <L'> || X_i
	concatenated = append(concatenated, pubKey...)
	// <L'> || X_i || R
	concatenated = append(concatenated, aggCommBytes...)
	// <L'> || X_i || R || m
	concatenated = append(concatenated, bn.data.message...)
	// H(<L'> || X_i || R || m)
	challenge := bn.hasher.Compute(string(concatenated))

	challengeScalar := bn.keyGen.Suite().CreateScalar()
	challengeScalar, err = challengeScalar.SetBytes(challenge)

	if err != nil {
		return nil, err
	}

	return challengeScalar, nil
}

// CreateSignatureShare creates a partial signature s_i = r_i + H(<L'> || X_i || R || m)*x_i
func (bn *belNevSigner) CreateSignatureShare(bitmap []byte) ([]byte, error) {
	if bitmap == nil {
		return nil, crypto.ErrNilBitmap
	}

	bn.mutSigData.Lock()
	defer bn.mutSigData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(bn.data.pubKeys)
	if flagsMismatch {
		return nil, crypto.ErrBitmapMismatch
	}
	ownIndex := bn.data.ownIndex

	challengeScalar, err := bn.computeChallenge(ownIndex, bitmap)
	if err != nil {
		return nil, err
	}

	privKeyScalar := bn.data.privKey.Scalar()
	// H(<L'> || X_i || R || m)*x_i
	sigShareScalar, err := challengeScalar.Mul(privKeyScalar)

	if err != nil {
		return nil, err
	}

	if bn.data.commSecret == nil {
		return nil, crypto.ErrNilCommitmentSecret
	}

	// s_i = r_i + H(<L'> || X_i || R || m)*x_i
	sigShareScalar, _ = sigShareScalar.Add(bn.data.commSecret)
	sigShare, _ := sigShareScalar.MarshalBinary()
	bn.data.sigShares[ownIndex] = sigShareScalar

	return sigShare, nil
}

// not concurrent safe, should be used under RLock mutex
func (bn *belNevSigner) isIndexInBitmap(index uint16, bitmap []byte) error {
	indexOutOfBounds := index >= uint16(len(bn.data.pubKeys))
	if indexOutOfBounds {
		return crypto.ErrIndexOutOfBounds
	}

	indexNotInBitmap := bitmap[index/8]&(1<<uint8(index%8)) == 0
	if indexNotInBitmap {
		return crypto.ErrIndexNotSelected
	}

	return nil
}

// VerifySignatureShare verifies the partial signature of the signer with specified position
// s_i * G = R_i + H1(<L'>||X_i||R||m)*X_i
func (bn *belNevSigner) VerifySignatureShare(index uint16, sig []byte, bitmap []byte) error {
	if sig == nil {
		return crypto.ErrNilSignature
	}

	if bitmap == nil {
		return crypto.ErrNilBitmap
	}

	suite := bn.keyGen.Suite()
	sigScalar := suite.CreateScalar()
	_ = sigScalar.UnmarshalBinary(sig)
	// s_i * G
	basePoint := suite.CreatePoint().Base()
	left, _ := basePoint.Mul(sigScalar)

	bn.mutSigData.RLock()
	defer bn.mutSigData.RUnlock()

	err := bn.isIndexInBitmap(index, bitmap)
	if err != nil {
		return err
	}

	challengeScalar, err := bn.computeChallenge(index, bitmap)
	if err != nil {
		return err
	}

	pubKey := bn.data.pubKeys[index].Point()
	// H1(<L'>||X_i||R||m)*X_i
	right, _ := pubKey.Mul(challengeScalar)
	// R_i + H1(<L'>||X_i||R||m)*X_i
	right, err = right.Add(bn.data.commitments[index])
	if err != nil {
		return err
	}

	eq, err := right.Equal(left)
	if err != nil {
		return err
	}

	if !eq {
		return crypto.ErrSigNotValid
	}

	return nil
}

// StoreSignatureShare adds the partial signature of the signer with specified position
func (bn *belNevSigner) StoreSignatureShare(index uint16, sig []byte) error {
	if sig == nil {
		return crypto.ErrNilSignature
	}

	sigScalar := bn.keyGen.Suite().CreateScalar()
	err := sigScalar.UnmarshalBinary(sig)
	if err != nil {
		return err
	}

	bn.mutSigData.Lock()
	defer bn.mutSigData.Unlock()

	if int(index) >= len(bn.data.sigShares) {
		return crypto.ErrIndexOutOfBounds
	}

	bn.data.sigShares[index] = sigScalar

	return nil
}

// SignatureShare returns the partial signature set for given index
func (bn *belNevSigner) SignatureShare(index uint16) ([]byte, error) {
	bn.mutSigData.RLock()
	defer bn.mutSigData.RUnlock()

	if int(index) >= len(bn.data.sigShares) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	if bn.data.sigShares[index] == nil {
		return nil, crypto.ErrNilElement
	}

	sigShareBytes, err := bn.data.sigShares[index].MarshalBinary()

	if err != nil {
		return nil, err
	}

	return sigShareBytes, nil
}

// AggregateSigs aggregates all collected partial signatures
func (bn *belNevSigner) AggregateSigs(bitmap []byte) ([]byte, error) {
	if bitmap == nil {
		return nil, crypto.ErrNilBitmap
	}

	bn.mutSigData.Lock()
	defer bn.mutSigData.Unlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(bn.data.pubKeys)
	if flagsMismatch {
		return nil, crypto.ErrBitmapMismatch
	}

	suite := bn.keyGen.Suite()
	aggSig := suite.CreateScalar().Zero()

	for i := range bn.data.sigShares {
		err := bn.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		aggSig, err = aggSig.Add(bn.data.sigShares[i])
		if err != nil {
			return nil, err
		}
	}

	isZero, _ := aggSig.Equal(suite.CreateScalar().Zero())

	if isZero {
		return nil, crypto.ErrBitmapNotSet
	}

	aggSigBytes, err := aggSig.MarshalBinary()
	if err != nil {
		return nil, err
	}

	bn.data.aggSig = aggSig

	aggCommBytes, err := bn.data.aggCommitment.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// concatenate signature and aggregated commitment
	resultSig := make([]byte, 0)
	resultSig = append(resultSig, aggCommBytes...)
	resultSig = append(resultSig, aggSigBytes...)

	return resultSig, nil
}

// SetAggregatedSig sets the aggregated signature
func (bn *belNevSigner) SetAggregatedSig(aggSig []byte) error {
	if aggSig == nil {
		return crypto.ErrNilSignature
	}

	suite := bn.keyGen.Suite()
	lenComm := suite.PointLen()
	lenSig := suite.ScalarLen()
	if len(aggSig) != lenComm+lenSig {
		return crypto.ErrAggSigNotValid
	}

	// unpack the commitment and signature
	aggCommPoint := suite.CreatePoint()

	err := aggCommPoint.UnmarshalBinary(aggSig[:lenComm])
	if err != nil {
		return err
	}

	aggSigScalar := suite.CreateScalar()

	err = aggSigScalar.UnmarshalBinary(aggSig[lenComm:])
	if err != nil {
		return err
	}

	bn.mutSigData.Lock()
	bn.data.aggCommitment = aggCommPoint
	bn.data.aggSig = aggSigScalar
	bn.mutSigData.Unlock()

	return nil
}

// Verify verifies the aggregated signature by checking equality
// s * G = R + Sum(H1(<L'>||X_i||R||m)*X_i*B[i])
func (bn *belNevSigner) Verify(bitmap []byte) error {
	if bitmap == nil {
		return crypto.ErrNilBitmap
	}

	bn.mutSigData.RLock()
	defer bn.mutSigData.RUnlock()

	maxFlags := len(bitmap) * 8
	flagsMismatch := maxFlags < len(bn.data.pubKeys)
	if flagsMismatch {
		return crypto.ErrBitmapMismatch
	}

	suite := bn.keyGen.Suite()
	right := suite.CreatePoint().Null()

	for i := range bn.data.pubKeys {
		err := bn.isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}

		challengeScalar, err := bn.computeChallenge(uint16(i), bitmap)
		if err != nil {
			return err
		}

		pubKey := bn.data.pubKeys[i].Point()
		// H1(<L'>||X_i||R||m)*X_i
		part, _ := pubKey.Mul(challengeScalar)
		right, _ = right.Add(part)
	}

	// R + Sum(H1(<L'>||X_i||R||m)*X_i)
	right, _ = right.Add(bn.data.aggCommitment)
	// s * G
	left := suite.CreatePoint().Base()

	if bn.data.aggSig == nil {
		return crypto.ErrNilSignature
	}

	left, _ = left.Mul(bn.data.aggSig)
	// s * G = R + Sum(H1(<L'>||X_i||R||m)*X_i)
	eq, _ := right.Equal(left)

	if !eq {
		return crypto.ErrSigNotValid
	}

	return nil
}
