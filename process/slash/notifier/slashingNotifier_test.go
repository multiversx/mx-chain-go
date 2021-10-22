package notifier_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	mockGenesis "github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/notifier"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/update"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNewSlashingNotifier(t *testing.T) {
	tests := []struct {
		args        func() *notifier.SlashingNotifierArgs
		expectedErr error
	}{
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.PrivateKey = nil
				return args
			},
			expectedErr: crypto.ErrNilPrivateKey,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.PublicKey = nil
				return args
			},
			expectedErr: crypto.ErrNilPublicKey,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.PubKeyConverter = nil
				return args
			},
			expectedErr: update.ErrNilPubKeyConverter,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.Signer = nil
				return args
			},
			expectedErr: crypto.ErrNilSingleSigner,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.AccountAdapter = nil
				return args
			},
			expectedErr: state.ErrNilAccountsAdapter,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.Hasher = nil
				return args
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.Marshaller = nil
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
	}

	for _, currTest := range tests {
		_, err := notifier.NewSlashingNotifier(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidProof_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	sn, _ := notifier.NewSlashingNotifier(args)

	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.SlashingProofStub{})
	require.Nil(t, tx)
	require.Equal(t, process.ErrInvalidProof, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidPubKey_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	errPubKey := errors.New("pub key error")
	args.PublicKey = &cryptoMocks.PublicKeyStub{
		ToByteArrayStub: func() ([]byte, error) {
			return nil, errPubKey
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{})
	require.Nil(t, tx)
	require.Equal(t, errPubKey, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidAccount_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	errAcc := errors.New("accounts adapter error")
	args.AccountAdapter = &stateMock.AccountsStub{
		GetExistingAccountCalled: func([]byte) (vmcommon.AccountHandler, error) {
			return nil, errAcc
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{})
	require.Nil(t, tx)
	require.Equal(t, errAcc, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidMarshaller_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	errMarshaller := errors.New("marshaller error")
	args.Marshaller = &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, errMarshaller
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{})
	require.Nil(t, tx)
	require.Equal(t, errMarshaller, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidSlashType_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()

	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{
		GetTypeCalled: func() slash.SlashingType {
			return "invalid"
		},
	})
	require.Nil(t, tx)
	require.Equal(t, process.ErrInvalidProof, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_NilHeader_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{
		GetHeadersCalled: func(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{nil}
		},
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{[]byte("pubKey")}
		},
	})
	require.Nil(t, tx)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidProofSignature_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	errSign := errors.New("signature error")
	args.Signer = &cryptoMocks.SignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return nil, errSign
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{})
	require.Nil(t, tx)
	require.Equal(t, errSign, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_InvalidTxSignature_ExpectError(t *testing.T) {
	args := generateSlashingNotifierArgs()
	errSign := errors.New("signature error")
	flag := false
	args.Signer = &cryptoMocks.SignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			if flag {
				return nil, errSign
			}

			flag = true
			return []byte("signature"), nil
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)
	tx, err := sn.CreateShardSlashingTransaction(&slashMocks.MultipleHeaderSigningProofStub{})
	require.Nil(t, tx)
	require.Equal(t, errSign, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_MultipleProposalProof(t *testing.T) {
	args := generateSlashingNotifierArgs()
	args.Hasher = &testscommon.HasherStub{
		ComputeCalled: func(string) []byte {
			return []byte{byte('a'), byte('b'), byte('c'), byte('d')}
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)

	h1 := slashMocks.CreateInterceptedHeaderData(
		&block.HeaderV2{
			Header: &block.Header{
				Round:        4,
				PrevRandSeed: []byte("seed1"),
			},
		},
	)
	h2 := slashMocks.CreateInterceptedHeaderData(
		&block.HeaderV2{
			Header: &block.Header{
				Round:        4,
				PrevRandSeed: []byte("seed2"),
			},
		},
	)
	proof := &slashMocks.MultipleHeaderProposalProofStub{
		GetHeadersCalled: func() []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{h1, h2}
		},
	}

	expectedData := []byte(notifier.BuiltInFunctionSlashCommitmentProof)
	expectedData = append(expectedData, byte('@'), slash.MultipleProposalProofID, byte('@'), byte('c'), byte('d'), byte('@'))
	expectedData = append(expectedData, []byte("signature")...)
	expectedTx := &transaction.Transaction{
		Data:      expectedData,
		Nonce:     444,
		SndAddr:   []byte("address"),
		Value:     big.NewInt(notifier.CommitmentProofValue),
		GasPrice:  notifier.CommitmentProofGasPrice,
		GasLimit:  notifier.CommitmentProofGasLimit,
		Signature: []byte("signature"),
	}

	tx, _ := sn.CreateShardSlashingTransaction(proof)
	require.Equal(t, expectedTx, tx)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_MultipleSignProof(t *testing.T) {
	args := generateSlashingNotifierArgs()
	args.Hasher = &testscommon.HasherStub{
		ComputeCalled: func(string) []byte {
			return []byte{byte('a'), byte('b'), byte('c'), byte('d')}
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)

	pk1 := []byte("pubKey1")

	h1 := slashMocks.CreateInterceptedHeaderData(
		&block.HeaderV2{
			Header: &block.Header{
				Round:        4,
				PrevRandSeed: []byte("seed1"),
			},
		},
	)
	h2 := slashMocks.CreateInterceptedHeaderData(
		&block.HeaderV2{
			Header: &block.Header{
				Round:        4,
				PrevRandSeed: []byte("seed2"),
			},
		},
	)
	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetLevelCalled: func([]byte) slash.ThreatLevel {
			return slash.High
		},
		GetHeadersCalled: func([]byte) []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{h1, h2}
		},
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{pk1}
		},
	}

	expectedData := []byte(notifier.BuiltInFunctionSlashCommitmentProof)
	expectedData = append(expectedData, byte('@'), slash.MultipleSigningProofID, byte('@'), byte('c'), byte('d'), byte('@'))
	expectedData = append(expectedData, []byte("signature")...)
	expectedTx := &transaction.Transaction{
		Data:      expectedData,
		Nonce:     444,
		SndAddr:   []byte("address"),
		Value:     big.NewInt(notifier.CommitmentProofValue),
		GasPrice:  notifier.CommitmentProofGasPrice,
		GasLimit:  notifier.CommitmentProofGasLimit,
		Signature: []byte("signature"),
	}

	tx, _ := sn.CreateShardSlashingTransaction(proof)
	require.Equal(t, expectedTx, tx)
}

func generateSlashingNotifierArgs() *notifier.SlashingNotifierArgs {
	accountHandler := &mockGenesis.BaseAccountMock{
		Nonce:             444,
		AddressBytesField: []byte("address"),
	}
	accountsAdapter := &stateMock.AccountsStub{
		GetExistingAccountCalled: func([]byte) (vmcommon.AccountHandler, error) {
			return accountHandler, nil
		},
	}

	return &notifier.SlashingNotifierArgs{
		PrivateKey:      &mock.PrivateKeyMock{},
		PublicKey:       &mock.PublicKeyMock{},
		PubKeyConverter: &testscommon.PubkeyConverterMock{},
		Signer:          &mock.SignerMock{},
		AccountAdapter:  accountsAdapter,
		Hasher:          &hashingMocks.HasherMock{},
		Marshaller:      &testscommon.MarshalizerMock{},
	}
}
