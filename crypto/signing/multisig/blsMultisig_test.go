package multisig_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/multisig"
	llsig "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/stretchr/testify/assert"
)

func genMultiSigParamsBLS(cnGrSize int, ownIndex uint16) (
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	pubKeys []string,
	kg crypto.KeyGenerator,
	llSigner crypto.LowLevelSignerBLS,
) {
	suite := kyber.NewSuitePairingBn256()
	kg = signing.NewKeyGenerator(suite)
	var pubKeyBytes []byte
	pubKeys = make([]string, 0)
	llSigner = &llsig.KyberMultiSignerBLS{}

	for i := 0; i < cnGrSize; i++ {
		sk, pk := kg.GeneratePair()
		if uint16(i) == ownIndex {
			privKey = sk
			pubKey = pk
		}

		pubKeyBytes, _ = pk.ToByteArray()
		pubKeys = append(pubKeys, string(pubKeyBytes))
	}

	return privKey, pubKey, pubKeys, kg, llSigner
}

func createSignerAndSigShareBLS(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	kg crypto.KeyGenerator,
	ownIndex uint16,
) (sigShare []byte, multiSig crypto.MultiSignerBLS) {
	llSigner := &llsig.KyberMultiSignerBLS{}
	multiSig, _ = multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.SetMessage([]byte("message"))
	sigShare, _ = multiSig.CreateSignatureShare()

	return sigShare, multiSig
}

func createSigSharesBLS(
	nbSigs uint16,
	grSize uint16,
	message []byte,
	ownIndex uint16,
) (sigShares [][]byte, multiSigner crypto.MultiSignerBLS) {

	hasher := &mock.HasherSpongeMock{}
	suite := kyber.NewSuitePairingBn256()
	kg := signing.NewKeyGenerator(suite)

	var pubKeyBytes []byte

	privKeys := make([]crypto.PrivateKey, grSize)
	pubKeys := make([]crypto.PublicKey, grSize)
	pubKeysStr := make([]string, grSize)

	for i := uint16(0); i < grSize; i++ {
		sk, pk := kg.GeneratePair()
		privKeys[i] = sk
		pubKeys[i] = pk

		pubKeyBytes, _ = pk.ToByteArray()
		pubKeysStr[i] = string(pubKeyBytes)
	}

	sigShares = make([][]byte, nbSigs)
	multiSigners := make([]crypto.MultiSignerBLS, nbSigs)
	llSigner := &llsig.KyberMultiSignerBLS{}

	for i := uint16(0); i < nbSigs; i++ {
		multiSigners[i], _ = multisig.NewBLSMultisig(llSigner, hasher, pubKeysStr, privKeys[i], kg, i)
	}

	for i := uint16(0); i < nbSigs; i++ {
		_ = multiSigners[i].SetMessage(message)
		sigShares[i], _ = multiSigners[i].CreateSignatureShare()
	}

	return sigShares, multiSigners[ownIndex]
}

func createAndAddSignatureSharesBLS() (multiSigner crypto.MultiSignerBLS, bitmap []byte) {
	grSize := uint16(15)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap = make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigSharesBLS(nbSigners, grSize, message, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	return multiSigner, bitmap
}

func createAggregatedSigBLS(t *testing.T) (multiSigner crypto.MultiSignerBLS, aggSig []byte, bitmap []byte) {
	multiSigner, bitmap = createAndAddSignatureSharesBLS()
	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, err)

	return multiSigner, aggSig, bitmap
}

func TestNewBLSMultisig_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, err := multisig.NewBLSMultisig(llSigner, nil, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilHasher, err)
}

func TestNewBLSMultisig_NilPrivKeyShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	_, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, nil, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilPrivateKey, err)
}

func TestNewBLSMultisig_NilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, _, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, nil, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNoPublicKeySet, err)
}

func TestNewBLSMultisig_NoPubKeysSetShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, _, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	pubKeys := make([]string, 0)

	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNoPublicKeySet, err)
}

func TestNewBLSMultisig_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, _, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, nil, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewBLSMultisig_InvalidOwnIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, _, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, nil, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrNilKeyGenerator, err)
}

func TestNewBLSMultisig_OutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, 10)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestNewBLSMultisig_InvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	pubKeys[1] = "invalid"

	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
}

func TestNewBLSMultisig_EmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	pubKeys[1] = ""

	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, multiSig)
	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
}

func TestNewBLSMultisig_OK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, err := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	assert.Nil(t, err)
	assert.NotNil(t, multiSig)
}

func TestBLSMultiSigner_CreateNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	multiSigCreated, err := multiSig.Create(nil, ownIndex)

	assert.Equal(t, crypto.ErrNoPublicKeySet, err)
	assert.Nil(t, multiSigCreated)
}

func TestBLSMultiSigner_CreateInvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = "invalid"
	multiSigCreated, err := multiSig.Create(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
	assert.Nil(t, multiSigCreated)
}

func TestBLSMultiSigner_CreateEmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = ""
	multiSigCreated, err := multiSig.Create(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
	assert.Nil(t, multiSigCreated)
}

func TestBLSMultiSigner_CreateOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	multiSigCreated, err := multiSig.Create(pubKeys, ownIndex)
	assert.Nil(t, err)
	assert.NotNil(t, multiSigCreated)
}

func TestBLSMultiSigner_ResetOutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.Reset(pubKeys, 10)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBLSMultiSigner_ResetNilPubKeysShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.Reset(nil, ownIndex)

	assert.Equal(t, crypto.ErrNilPublicKeys, err)
}

func TestBLSMultiSigner_ResetInvalidPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = "invalid"
	err := multiSig.Reset(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrInvalidPublicKeyString, err)
}

func TestBLSMultiSigner_ResetEmptyPubKeyInListShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	pubKeys[1] = ""
	err := multiSig.Reset(pubKeys, ownIndex)

	assert.Equal(t, crypto.ErrEmptyPubKeyString, err)
}

func TestBLSMultiSigner_ResetOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.Reset(pubKeys, ownIndex)
	assert.Nil(t, err)
}

func TestBLSMultiSigner_SetNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetMessage(nil)

	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBLSMultiSigner_SetEmptyMessageShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetMessage([]byte(""))

	assert.Equal(t, crypto.ErrInvalidParam, err)
}

func TestBLSMultiSigner_SetMessageShouldOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.SetMessage([]byte("message"))

	assert.Nil(t, err)
}

func TestBLSMultiSigner_CreateSignatureShareNotSetMessageShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	sigShare, err := multiSig.CreateSignatureShare()

	assert.Nil(t, sigShare)
	assert.Equal(t, crypto.ErrNilMessage, err)
}

func TestBLSMultiSigner_CreateSignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner:= genMultiSigParamsBLS(4, ownIndex)

	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.SetMessage([]byte("message"))
	sigShare, err := multiSig.CreateSignatureShare()

	verifErr := multiSig.VerifySignatureShare(ownIndex, sigShare)

	assert.Nil(t, err)
	assert.NotNil(t, sigShare)
	assert.Nil(t, verifErr)
}

func TestBLSMultiSigner_VerifySignatureShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	_, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)

	verifErr := multiSig.VerifySignatureShare(ownIndex, nil)

	assert.Equal(t, crypto.ErrNilSignature, verifErr)
}

func TestBLSMultiSigner_VerifySignatureShareInvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	sigShare, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)

	verifErr := multiSig.VerifySignatureShare(0, sigShare)

	assert.Contains(t, verifErr.Error(), "invalid signature")
}

func TestBLSMultiSigner_VerifySignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	sigShare, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)

	verifErr := multiSig.VerifySignatureShare(ownIndex, sigShare)

	assert.Nil(t, verifErr)
}

func TestBLSMultiSigner_AddSignatureShareNilSigShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	err := multiSig.StoreSignatureShare(ownIndex, nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBLSMultiSigner_AddSignatureShareIndexOutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, llSigner := genMultiSigParamsBLS(4, ownIndex)
	multiSig, _ := multisig.NewBLSMultisig(llSigner, hasher, pubKeys, privKey, kg, ownIndex)

	_ = multiSig.SetMessage([]byte("message"))
	sigShare, _ := multiSig.CreateSignatureShare()

	err := multiSig.StoreSignatureShare(15, sigShare)

	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBLSMultiSigner_AddSignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	sigShare, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)
	err := multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, _ := multiSig.SignatureShare(ownIndex)

	assert.Nil(t, err)
	assert.Equal(t, sigShare, sigShareRead)
}

func TestBLSMultiSigner_SignatureShareOutOfBoundsIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	sigShare, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, err := multiSig.SignatureShare(15)

	assert.Nil(t, sigShareRead)
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestBLSMultiSigner_SignatureShareNotSetIndexShouldErr(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	sigShare, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, err := multiSig.SignatureShare(2)

	assert.Nil(t, sigShareRead)
	assert.Equal(t, crypto.ErrNilElement, err)
}

func TestBLSMultiSigner_SignatureShareOK(t *testing.T) {
	t.Parallel()

	ownIndex := uint16(3)
	hasher := &mock.HasherSpongeMock{}
	privKey, _, pubKeys, kg, _ := genMultiSigParamsBLS(4, ownIndex)
	sigShare, multiSig := createSignerAndSigShareBLS(hasher, pubKeys, privKey, kg, ownIndex)
	_ = multiSig.StoreSignatureShare(ownIndex, sigShare)
	sigShareRead, err := multiSig.SignatureShare(ownIndex)

	assert.Nil(t, err)
	assert.Equal(t, sigShare, sigShareRead)
}

func TestBLSMultiSigner_AggregateSigsNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigSharesBLS(nbSigners, grSize, message, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	aggSig, err := multiSigner.AggregateSigs(nil)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBLSMultiSigner_AggregateSigsInvalidBitmapShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(21)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 3)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigSharesBLS(nbSigners, grSize, message, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	bitmap = make([]byte, 1)
	bitmap[0] = 0x07

	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBLSMultiSigner_AggregateSigsMissingSigShareShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigSharesBLS(nbSigners, grSize, message, ownIndex)

	for i := 0; i < len(sigs)-1; i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrNilParam, err)
}

func TestBLSMultiSigner_AggregateSigsZeroSelectionBitmapShouldErr(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigSharesBLS(nbSigners, grSize, message, ownIndex)

	for i := 0; i < len(sigs)-1; i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}
	bitmap[0] = 0
	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, aggSig)
	assert.Equal(t, crypto.ErrNilSignaturesList, err)
}

func TestBLSMultiSigner_AggregateSigsOK(t *testing.T) {
	t.Parallel()

	grSize := uint16(6)
	ownIndex := uint16(0)
	nbSigners := uint16(3)
	message := []byte("message")
	bitmap := make([]byte, 2)
	bitmap[0] = 0x07

	sigs, multiSigner := createSigSharesBLS(nbSigners, grSize, message, ownIndex)

	for i := 0; i < len(sigs); i++ {
		_ = multiSigner.StoreSignatureShare(uint16(i), sigs[i])
	}

	aggSig, err := multiSigner.AggregateSigs(bitmap)

	assert.Nil(t, err)
	assert.NotNil(t, aggSig)
}

func TestBLSMultiSigner_SetAggregatedSigNilSigShouldErr(t *testing.T) {
	t.Parallel()

	multiSigner, _, _ := createAggregatedSigBLS(t)
	err := multiSigner.SetAggregatedSig(nil)

	assert.Equal(t, crypto.ErrNilSignature, err)
}

func TestBLSMultiSigner_SetAggregatedSigInvalidScalarShouldErr(t *testing.T) {
	t.Parallel()

	multiSigner, _, _ := createAggregatedSigBLS(t)
	aggSig := []byte("invalid agg signature xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	err := multiSigner.SetAggregatedSig(aggSig)

	assert.Contains(t, err.Error(), "malformed point")
}

func TestBLSMultiSigner_SetAggregatedSigOK(t *testing.T) {
	t.Parallel()

	multiSigner, aggSig, _ := createAggregatedSigBLS(t)
	err := multiSigner.SetAggregatedSig(aggSig)

	assert.Nil(t, err)
}

func TestBLSMultiSigner_VerifyNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	multiSigner, aggSig, _ := createAggregatedSigBLS(t)
	_ = multiSigner.SetAggregatedSig(aggSig)
	err := multiSigner.Verify(nil)

	assert.Equal(t, crypto.ErrNilBitmap, err)
}

func TestBLSMultiSigner_VerifyBitmapMismatchShouldErr(t *testing.T) {
	t.Parallel()

	multiSigner, aggSig, bitmap := createAggregatedSigBLS(t)
	_ = multiSigner.SetAggregatedSig(aggSig)
	// set a smaller bitmap
	bitmap = make([]byte, 1)

	err := multiSigner.Verify(bitmap)
	assert.Equal(t, crypto.ErrBitmapMismatch, err)
}

func TestBLSMultiSigner_VerifyAggSigNotSetShouldErr(t *testing.T) {
	t.Parallel()

	multiSigner, bitmap := createAndAddSignatureSharesBLS()
	err := multiSigner.Verify(bitmap)

	assert.Contains(t, err.Error(), "not enough data")
}

func TestBLSMultiSigner_VerifySigValid(t *testing.T) {
	t.Parallel()

	multiSigner, aggSig, bitmap := createAggregatedSigBLS(t)
	_ = multiSigner.SetAggregatedSig(aggSig)

	err := multiSigner.Verify(bitmap)
	assert.Nil(t, err)
}

func TestBLSMultiSigner_VerifySigInvalid(t *testing.T) {
	t.Parallel()

	multiSigner, aggSig, bitmap := createAggregatedSigBLS(t)
	// make sig invalid
	aggSig[len(aggSig)-1] = aggSig[len(aggSig)-1] ^ 255
	_ = multiSigner.SetAggregatedSig(aggSig)

	err := multiSigner.Verify(bitmap)
	assert.Contains(t, err.Error(), "malformed point")
}
