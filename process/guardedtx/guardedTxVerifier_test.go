package guardedtx

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
	txStruct "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/guardianMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/interceptedTxMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/vmcommonMocks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNewGuardedTxSigVerifier(t *testing.T) {
	args := GuardedTxSigVerifierArgs{
		SigVerifier:     &cryptoMocks.SingleSignerStub{},
		GuardianChecker: &guardianMocks.GuardianCheckerStub{},
		PubKeyConverter: &testscommon.PubkeyConverterMock{},
		Marshaller:      &testscommon.MarshalizerMock{},
		KeyGen:          &cryptoMocks.KeyGenStub{},
	}

	t.Run("nil guardian sig verifier ", func(t *testing.T) {
		changedArgs := *&args
		changedArgs.SigVerifier = nil
		gtxSigVerifier, err := NewGuardedTxSigVerifier(changedArgs)
		require.Equal(t, process.ErrNilSingleSigner, err)
		require.Nil(t, gtxSigVerifier)
	})
	t.Run("nil guardian checker", func(t *testing.T) {
		changedArgs := *&args
		changedArgs.GuardianChecker = nil
		gtxSigVerifier, err := NewGuardedTxSigVerifier(changedArgs)
		require.Equal(t, process.ErrNilGuardianChecker, err)
		require.Nil(t, gtxSigVerifier)
	})
	t.Run("nil public key converter", func(t *testing.T) {
		changedArgs := *&args
		changedArgs.PubKeyConverter = nil
		gtxSigVerifier, err := NewGuardedTxSigVerifier(changedArgs)
		require.Equal(t, process.ErrNilPubkeyConverter, err)
		require.Nil(t, gtxSigVerifier)
	})
	t.Run("ok params", func(t *testing.T) {
		gtxSigVerifier, err := NewGuardedTxSigVerifier(args)
		require.Nil(t, err)
		require.NotNil(t, gtxSigVerifier)
	})
}

func TestGuardedTxSigVerifier_IsInterfaceNil(t *testing.T) {
	var gsv GuardianSigVerifier
	require.True(t, check.IfNil(gsv))

	var gsvNilPtr *guardedTxSigVerifier
	require.True(t, check.IfNil(gsvNilPtr))

	gsv = &guardedTxSigVerifier{}
	require.False(t, check.IfNil(gsv))
}

func TestGuardedTxSigVerifier_VerifyGuardianSignature(t *testing.T) {
	suite := ed25519.NewEd25519()
	keyGenerator := signing.NewKeyGenerator(suite)
	privateKeyGuardian, publicKeyGuardian := keyGenerator.GeneratePair()
	privateKeyOwner, publicKeyOwner := keyGenerator.GeneratePair()
	pubKeyGuardianBytes, _ := publicKeyGuardian.ToByteArray()
	publicKeyOwnerBytes, _ := publicKeyOwner.ToByteArray()

	signer := &singlesig.Ed25519Signer{}

	acc := &vmcommonMocks.UserAccountStub{}
	guardianChecker := &guardianMocks.GuardianCheckerStub{
		GetActiveGuardianCalled: func(handler vmcommon.UserAccountHandler) ([]byte, error) {
			return pubKeyGuardianBytes, nil
		},
	}

	value, _ := big.NewInt(0).SetString("100000000000000000000", 10)

	tx := &txStruct.Transaction{
		Nonce:             11,
		Value:             value,
		RcvAddr:           publicKeyOwnerBytes,
		RcvUserName:       nil,
		SndAddr:           publicKeyOwnerBytes,
		SndUserName:       nil,
		GasPrice:          1000000000,
		GasLimit:          50000,
		Data:              nil,
		ChainID:           []byte{31},
		Version:           2,
		Signature:         nil,
		Options:           0,
		GuardianSignature: nil,
	}

	marshaller := &marshal.JsonMarshalizer{}
	converter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, &mock.LoggerMock{})
	inTx := createSignedInterceptedTx(tx, signer, privateKeyOwner, privateKeyGuardian, converter, marshaller)
	args := GuardedTxSigVerifierArgs{
		SigVerifier:     signer,
		GuardianChecker: guardianChecker,
		PubKeyConverter: converter,
		Marshaller:      marshaller,
		KeyGen:          keyGenerator,
	}

	t.Run("verify OK", func(t *testing.T) {
		gtxSigVerifier, err := NewGuardedTxSigVerifier(args)
		require.Nil(t, err)

		err = gtxSigVerifier.VerifyGuardianSignature(acc, inTx)
		require.Nil(t, err)
	})
	t.Run("invalid guardian signature", func(t *testing.T) {
		gtxSigVerifier, err := NewGuardedTxSigVerifier(args)
		require.Nil(t, err)

		// owner signs for guardian
		inTx2 := createSignedInterceptedTx(tx, signer, privateKeyOwner, privateKeyOwner, converter, marshaller)
		err = gtxSigVerifier.VerifyGuardianSignature(acc, inTx2)
		require.NotNil(t, err)
	})
	t.Run("nil guardian signature", func(t *testing.T) {
		gtxSigVerifier, err := NewGuardedTxSigVerifier(args)
		require.Nil(t, err)

		// owner signs for guardian
		inTx2 := createSignedInterceptedTx(tx, signer, privateKeyOwner, privateKeyOwner, converter, marshaller)
		txHandler := inTx2.Transaction()
		inTx2.TransactionCalled = func() data.TransactionHandler {
			tx2, ok := txHandler.(*txStruct.Transaction)
			require.True(t, ok)

			tx2.GuardianSignature = nil
			return tx2
		}
		err = gtxSigVerifier.VerifyGuardianSignature(acc, inTx2)
		require.NotNil(t, err)
	})
}

func createSignedInterceptedTx(
	tx *txStruct.Transaction,
	signer crypto.SingleSigner,
	senderSk crypto.PrivateKey,
	guardianSk crypto.PrivateKey,
	converter core.PubkeyConverter,
	marshaller data.Marshaller,
) *interceptedTxMocks.InterceptedTxHandlerStub {
	ownerSig, guardianSig := signAndGuardTx(tx, signer, senderSk, guardianSk, converter, marshaller)

	txCopy := *tx
	txCopy.Signature = ownerSig
	txCopy.GuardianSignature = guardianSig

	return &interceptedTxMocks.InterceptedTxHandlerStub{
		TransactionCalled: func() data.TransactionHandler {
			return &txCopy
		},
	}
}

func signAndGuardTx(
	tx *txStruct.Transaction,
	signer crypto.SingleSigner,
	sk crypto.PrivateKey,
	skGuardian crypto.PrivateKey,
	converter core.PubkeyConverter,
	marshaller data.Marshaller,
) ([]byte, []byte) {
	ftx := &txStruct.FrontendTransaction{
		Nonce:             tx.Nonce,
		Value:             tx.Value.String(),
		Receiver:          converter.Encode(tx.RcvAddr),
		Sender:            converter.Encode(tx.RcvAddr),
		SenderUsername:    nil,
		ReceiverUsername:  nil,
		GasPrice:          tx.GasPrice,
		GasLimit:          tx.GasLimit,
		Data:              tx.Data,
		Signature:         "",
		ChainID:           string(tx.ChainID),
		Version:           tx.Version,
		GuardianSignature: "",
	}

	buff, _ := marshaller.Marshal(ftx)
	signature, _ := signer.Sign(sk, buff)
	guardianSignature, _ := signer.Sign(skGuardian, buff)

	return signature, guardianSignature
}
