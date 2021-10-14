package notifier_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	mock3 "github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	mock4 "github.com/ElrondNetwork/elrond-go/process/slash/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash/notifier"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	state2 "github.com/ElrondNetwork/elrond-go/testscommon/state"
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
				args.Marshaller = nil
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
		{
			args: func() *notifier.SlashingNotifierArgs {
				args := generateSlashingNotifierArgs()
				args.Hasher = nil
				return args
			},
			expectedErr: process.ErrNilHasher,
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

	tx, err := sn.CreateShardSlashingTransaction(&mock4.SlashingProofStub{})
	require.Nil(t, tx)
	require.Equal(t, process.ErrInvalidProof, err)
}

func TestSlashingNotifier_CreateShardSlashingTransaction_MultipleProposalProof(t *testing.T) {
	args := generateSlashingNotifierArgs()
	args.Hasher = &testscommon.HasherStub{
		ComputeCalled: func(string) []byte {
			return []byte{byte('a'), byte('b'), byte('c'), byte('d')}
		},
	}

	sn, _ := notifier.NewSlashingNotifier(args)

	h1 := testscommon.CreateInterceptedHeaderData(&block.Header{
		Round:        4,
		PrevRandSeed: []byte("seed1"),
	})
	h2 := testscommon.CreateInterceptedHeaderData(&block.Header{
		Round:        4,
		PrevRandSeed: []byte("seed2"),
	})
	proof := &mock4.MultipleHeaderProposalProofStub{
		GetHeadersCalled: func() []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{h1, h2}
		},
	}

	expectedData := []byte{slash.MultipleProposalProofID, byte('@'), byte('c'), byte('d'), byte('@')}
	expectedData = append(expectedData, []byte("signature")...)

	tx, _ := sn.CreateShardSlashingTransaction(proof)
	require.NotNil(t, tx)
	require.Equal(t, expectedData, tx.GetData())
	require.Equal(t, tx.GetValue(), big.NewInt(notifier.CommitmentProofValue))
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

	h1 := testscommon.CreateInterceptedHeaderData(&block.Header{
		Round:        4,
		PrevRandSeed: []byte("seed1"),
	})
	h2 := testscommon.CreateInterceptedHeaderData(&block.Header{
		Round:        4,
		PrevRandSeed: []byte("seed2"),
	})
	proof := &mock4.MultipleHeaderSigningProofStub{
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

	expectedData := []byte{slash.MultipleSigningProofID, byte('@'), byte('c'), byte('d'), byte('@')}
	expectedData = append(expectedData, []byte("signature")...)

	tx, _ := sn.CreateShardSlashingTransaction(proof)
	require.NotNil(t, tx)
	require.Equal(t, expectedData, tx.GetData())
	require.Equal(t, tx.GetValue(), big.NewInt(notifier.CommitmentProofValue))
}

func generateSlashingNotifierArgs() *notifier.SlashingNotifierArgs {
	accountHandler := &mock3.BaseAccountMock{}
	accountsAdapter := &state2.AccountsStub{
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
		Marshaller:      &testscommon.MarshalizerMock{},
		Hasher:          &testscommon.HasherMock{},
	}
}
