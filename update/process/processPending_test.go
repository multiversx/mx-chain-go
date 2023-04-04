package process

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

func createMockArgsPendingTransactionProcessor() ArgsPendingTransactionProcessor {
	return ArgsPendingTransactionProcessor{
		Accounts:         &stateMock.AccountsStub{},
		TxProcessor:      &testscommon.TxProcessorMock{},
		RwdTxProcessor:   &testscommon.RewardTxProcessorMock{},
		ScrTxProcessor:   &testscommon.SCProcessorMock{},
		PubKeyConv:       &testscommon.PubkeyConverterStub{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
	}
}

func TestNewPendingTransactionProcessor(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingTransactionProcessor()

	pendingTxProcessor, err := NewPendingTransactionProcessor(args)
	assert.NoError(t, err)
	assert.False(t, check.IfNil(pendingTxProcessor))
}

func TestPendingTransactionProcessor_ProcessTransactionsDstMe(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")
	addr4 := []byte("addr4")
	addr5 := []byte("addr5")
	args := createMockArgsPendingTransactionProcessor()
	args.PubKeyConv = &testscommon.PubkeyConverterStub{}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if address != nil {
			return 0
		}
		return 1
	}
	_ = shardCoordinator.SetSelfId(1)
	args.ShardCoordinator = shardCoordinator

	args.TxProcessor = &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			if bytes.Equal(transaction.SndAddr, addr4) {
				return 0, errors.New("localErr")
			}
			return 0, nil
		},
	}

	called := false
	args.Accounts = &stateMock.AccountsStub{
		CommitCalled: func() ([]byte, error) {
			called = true
			return nil, nil
		},
	}

	pendingTxProcessor, _ := NewPendingTransactionProcessor(args)

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hash3 := []byte("hash3")
	hash4 := []byte("hash4")
	hash5 := []byte("hash5")

	tx1 := &transaction.Transaction{RcvAddr: addr1}
	tx2 := &transaction.Transaction{SndAddr: addr2}
	tx3 := &transaction.Transaction{SndAddr: addr3, RcvAddr: addr3}
	tx4 := &transaction.Transaction{SndAddr: addr4}
	tx5 := &transaction.Transaction{SndAddr: addr5}

	mbInfo := &update.MbInfo{
		TxsInfo: []*update.TxInfo{
			{TxHash: hash1, Tx: tx1},
			{TxHash: hash2, Tx: tx2},
			{TxHash: hash3, Tx: tx3},
			{TxHash: hash4, Tx: tx4},
			{TxHash: hash5, Tx: tx5},
		}}

	mb, err := pendingTxProcessor.ProcessTransactionsDstMe(mbInfo)
	_, _ = pendingTxProcessor.Commit()
	assert.True(t, called)
	assert.NotNil(t, mb)
	assert.NoError(t, err)
	assert.Equal(t, hash2, mb.TxHashes[1])
}

func TestRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	args := createMockArgsPendingTransactionProcessor()
	args.Accounts = &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	pendingTxProcessor, _ := NewPendingTransactionProcessor(args)

	rHash, _ := pendingTxProcessor.RootHash()
	assert.Equal(t, rootHash, rHash)
}
