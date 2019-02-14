package multisig_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/multisig"
	"github.com/stretchr/testify/assert"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

func genMultiSigParams(cnGrSize int) (
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	pubKeys []string,
	kg crypto.KeyGenerator,
	ownIndex uint16,
) {
	suite := kv2.NewBlakeSHA256Ed25519()
	kg = signing.NewKeyGenerator(suite)

	var pubKeyBytes []byte
	pubKeys = make([]string, 0)
	for i := 0; i < cnGrSize; i++ {
		privKey, pubKey = kg.GeneratePair()
		pubKeyBytes, _ = pubKey.ToByteArray()
		pubKeys = append(pubKeys, string(pubKeyBytes))
	}

	return privKey, pubKey, pubKeys, kg, uint16(cnGrSize - 1)
}

func setComms(multiSig crypto.MultiSigner, grSize uint16) (bitmap []byte) {
	_, comm := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(0, comm)

	commSecret, comm := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(grSize-1, comm)
	_ = multiSig.SetCommitmentSecret(commSecret)

	bitmap = make([]byte, grSize/8+1)
	bitmap[0] = 1
	bitmap[grSize/8] |= 1 << ((grSize - 1) % 8)

	return bitmap
}

func createSignerAndSigShare(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	kg crypto.KeyGenerator,
	grSize uint16,
	ownIndex uint16) (sigShare []byte, multiSig crypto.MultiSigner, bitmap []byte) {

	multiSig, _ = multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.SetMessage([]byte("message"))
	bitmap = setComms(multiSig, grSize)
	_, _ = multiSig.AggregateCommitments(bitmap)
	sigShare, _ = multiSig.CreateSignatureShare(bitmap)

	return sigShare, multiSig, bitmap

}

func TestNewBelNevMultisig_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	multiSig, err := multisig.NewBelNevMultisig(nil, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilHasher, err)
}

func TestNewBelNevMultisig_NilPrivKeyShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	_, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, nil, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestNewBelNevMultisig_NilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, _, kg, ownIndex := genMultiSigParams(4)
	multiSig, err := multisig.NewBelNevMultisig(hasher, nil, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestNewBelNevMultisig_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, _, ownIndex := genMultiSigParams(4)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, nil, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewBelNevMultisig_InvalidOwnIndexShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, _, ownIndex := genMultiSigParams(4)
	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, nil, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewBelNevMultisig_InvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	pubKeys[1] = "invalid"

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
}

func TestNewBelNevMultisig_EmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	pubKeys[1] = ""

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
}

func TestNewBelNevMultisig_OK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, err)
	assert.NotNil(t, multiSig)
}

func TestBelNevSigner_ResetNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.Reset(nil, ownIndex)

	assert.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestBelNevSigner_ResetInvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = "invalid"
	err := multiSig.Reset(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
}

func TestBelNevSigner_ResetEmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = ""
	err := multiSig.Reset(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
}

func TestBelNevSigner_ResetOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commSecret, comm := multiSig.CreateCommitment()

	multiSig.SetCommitmentSecret(commSecret)
	multiSig.AddCommitment(0, comm)

	err := multiSig.Reset(pubKeys, ownIndex)
	assert.Nil(t, err)

	_, err = multiSig.Commitment(ownIndex)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_SetNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetMessage(nil)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBelNevSigner_SetEmptyMessageShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetMessage([]byte(""))

	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestBelNevSigner_SetMessageShouldOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetMessage([]byte("message"))

	assert.Nil(t, err)
}

func TestBelNevSigner_AddCommitmentHashInvalidIndexShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	// invalid index > len(pubKeys)
	err := multiSig.AddCommitmentHash(100, commHash)

	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_AddCommitmentHashNilCommitmentHashShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.AddCommitmentHash(0, nil)

	assert.Equal(t, crypto.ErrNilCommitmentHash, err)
}

func TestBelNevSigner_AddCommitmentHashOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	err := multiSig.AddCommitmentHash(0, commHash)
	assert.Nil(t, err)

	commHashRead, _ := multiSig.CommitmentHash(0)
	assert.Equal(t, commHash, commHashRead)
}

func TestBelNevSigner_CommitmentHashIndexOutOfBoundsShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	_ = multiSig.AddCommitmentHash(0, commHash)
	// index > len(pubKeys)
	commHashRead, err := multiSig.CommitmentHash(100)

	assert.Nil(t, commHashRead)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_CommitmentHashNotSetIndexShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHashRead, err := multiSig.CommitmentHash(0)

	assert.Nil(t, commHashRead)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_CommitmentHashOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commHash := []byte("commHash")
	_ = multiSig.AddCommitmentHash(0, commHash)
	commHashRead, err := multiSig.CommitmentHash(0)

	assert.Nil(t, err)
	assert.Equal(t, commHash, commHashRead)
}

func TestBelNevSigner_CreateCommitmentOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	secret, comm := multiSig.CreateCommitment()

	sk, _ := kg.PrivateKeyFromByteArray(secret)
	pk := sk.GeneratePublic()
	pkBytes, err := pk.Point().MarshalBinary()

	assert.Nil(t, err)
	assert.NotNil(t, secret)
	assert.NotNil(t, comm)
	assert.Equal(t, comm, pkBytes)
}

func TestBelNevSigner_SetCommitmentSecretNilCommSecretShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetCommitmentSecret(nil)

	assert.Equal(t, crypto.ErrNilCommitmentSecret, err)
}

func TestBelNevSigner_SetCommitmentSecretOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commSecret, _ := multiSig.CreateCommitment()
	err := multiSig.SetCommitmentSecret(commSecret)
	commSecretRead, _ := multiSig.CommitmentSecret()

	assert.Nil(t, err)
	assert.Equal(t, commSecret, commSecretRead)
}

func TestBelNevSigner_AddCommitmentNilCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.AddCommitment(0, nil)

	assert.Equal(t, crypto.ErrNilCommitment, err)
}

func TestBelNevSigner_AddCommitmentIndexOutOfBoundsShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	err := multiSig.AddCommitment(100, commitment)

	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_AddCommitmentOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	err := multiSig.AddCommitment(0, commitment)

	assert.Nil(t, err)
}

func TestBelNevSigner_CommitmentIndexOutOfBoundsShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(0, commitment)

	commRead, err := multiSig.Commitment(100)

	assert.Nil(t, commRead)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBelNevSigner_CommitmentNilCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	commRead, err := multiSig.Commitment(0)

	assert.Nil(t, commRead)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBelNevSigner_CommitmentOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(0, commitment)
	commRead, err := multiSig.Commitment(0)

	assert.Nil(t, err)
	assert.Equal(t, commRead, commitment)
}

func TestBelNevSigner_AggregateCommitmentsNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_ = setComms(multiSig, 4)
	aggComm, err := multiSig.AggregateCommitments(nil)

	assert.Nil(t, aggComm)
	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBelNevSigner_AggregateCommitmentsWrongBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(21)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 21)

	aggComm, err := multiSig.AggregateCommitments(bitmap[:1])

	assert.Nil(t, aggComm)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBelNevSigner_AggregateCommitmentsNotSetCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, commitment := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(0, commitment)
	_, commitment2 := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(2, commitment2)

	bitmap := make([]byte, 1)
	bitmap[0] |= 7 // 0b00000111

	aggComm, err := multiSig.AggregateCommitments(bitmap)

	assert.Nil(t, aggComm)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestBelNevSigner_AggregateCommitmentsOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	aggComm, err := multiSig.AggregateCommitments(bitmap)

	assert.Nil(t, err)
	assert.NotNil(t, aggComm)
}

func TestBelNevSigner_SetAggCommitmentNilAggCommShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetAggCommitment(nil)

	assert.Equal(t, crypto.ErrNilAggregatedCommitment, err)
}

func TestBelNevSigner_SetAggCommitmentOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, comm := multiSig.CreateCommitment()

	err := multiSig.SetAggCommitment(comm)
	aggCommRead, _ := multiSig.AggCommitment()

	assert.Nil(t, err)
	assert.Equal(t, comm, aggCommRead)
}

func TestBelNevSigner_AggCommitmentNilShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)

	aggCommRead, err := multiSig.AggCommitment()

	assert.Nil(t, aggCommRead)
	assert.Equal(t, crypto.ErrNilAggregatedCommitment, err)
}

func TestBelNevSigner_AggCommitmentOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)
	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, comm := multiSig.CreateCommitment()
	_ = multiSig.SetAggCommitment(comm)

	aggCommRead, err := multiSig.AggCommitment()

	assert.NotNil(t, aggCommRead)
	assert.Nil(t, err)
	assert.Equal(t, comm, aggCommRead)
}

func TestBelNevSigner_CreateSignatureShareNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(3)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.SetMessage([]byte("message"))
	bitmap := setComms(multiSig, 3)
	_, _ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare(nil)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBelNevSigner_CreateSignatureShareInvalidBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(21)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.SetMessage([]byte("message"))
	bitmap := setComms(multiSig, 21)
	_, _ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare(bitmap[:1])

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBelNevSigner_CreateSignatureShareNotSetCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(15)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, comm := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(0, comm)

	bitmap := make([]byte, 2)
	bitmap[0] = 0x01

	_, _ = multiSig.AggregateCommitments(bitmap)
	_ = multiSig.SetMessage([]byte("message"))

	// corrupt bitmap by selecting a not set commitment
	bitmap[0] = 0x10
	sigShare, err := multiSig.CreateSignatureShare(bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilCommitment, err)
}

func TestBelNevSigner_CreateSignatureShareNotSetMessageShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	_, _ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare(bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBelNevSigner_CreateSignatureShareNotSetAggCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)

	_ = multiSig.SetMessage([]byte("message"))
	sigShare, err := multiSig.CreateSignatureShare(bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilAggregatedCommitment, err)
}

func TestBelNevSigner_CreateSignatureShareNotSetCommSecretShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, err := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	_, comm := multiSig.CreateCommitment()
	_ = multiSig.AddCommitment(3, comm)

	bitmap := make([]byte, 1)
	bitmap[0] = 1 << 3

	_ = multiSig.SetMessage([]byte("message"))
	_, _ = multiSig.AggregateCommitments(bitmap)
	sigShare, err := multiSig.CreateSignatureShare(bitmap)

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilCommitmentSecret, err)
}

func TestBelNevSigner_CreateSignatureShareOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)

	multiSig, _ := multisig.NewBelNevMultisig(hasher, pubKeys, privKey, kg, ownIndex)
	bitmap := setComms(multiSig, 4)
	_, _ = multiSig.AggregateCommitments(bitmap)
	_ = multiSig.SetMessage([]byte("message"))
	sigShare, err := multiSig.CreateSignatureShare(bitmap)

	verifErr := multiSig.VerifySignatureShare(3, sigShare, bitmap)

	assert.Nil(t, err)
	assert.NotNil(t, sigShare)
	assert.Nil(t, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	_, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)

	verifErr := multiSig.VerifySignatureShare(3, nil, bitmap)

	assert.Equal(t, crypto.ErrNilSignature, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	sigShare, multiSig, _ := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)

	verifErr := multiSig.VerifySignatureShare(3, sigShare, nil)

	assert.Equal(t, crypto.ErrNilBitmap, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNotSelectedIndexShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)

	bitmap[0] = 0
	verifErr := multiSig.VerifySignatureShare(3, sigShare, bitmap)

	assert.Equal(t, crypto.ErrIndexNotSelected, verifErr)
}

func TestBelNevSigner_VerifySignatureShareNotSetCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)

	bitmap[0] |= 1 << 2
	verifErr := multiSig.VerifySignatureShare(2, sigShare, bitmap)

	assert.Equal(t, crypto.ErrNilCommitment, verifErr)
}

func TestBelNevSigner_VerifySignatureShareInvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)

	verifErr := multiSig.VerifySignatureShare(0, sigShare, bitmap)

	assert.Equal(t, crypto.ErrSigNotValid, verifErr)
}

func TestBelNevSigner_VerifySignatureShareOK(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherMock{}
	privKey, _, pubKeys, kg, ownIndex := genMultiSigParams(4)
	sigShare, multiSig, bitmap := createSignerAndSigShare(hasher, pubKeys, privKey, kg, 4, ownIndex)

	verifErr := multiSig.VerifySignatureShare(3, sigShare, bitmap)

	assert.Nil(t, verifErr)
}

