package postprocess

import (
	"bytes"
	"math/big"
	"sort"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

const maxGasLimitPerBlock = uint64(1500000000)

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestNewIntermediateResultsProcessor_NilHashes(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		nil,
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateResultsProcessor_NilMarshalizer(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(5),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateResultsProcessor_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateResultsProcessor_NilPubkeyConverter(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		nil,
		&mock.ChainStorerMock{},
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewIntermediateResultsProcessor_NilStorer(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		createMockPubkeyConverter(),
		nil,
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateResultsProcessor_NilTxForCurrentBlockHandler(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.TxBlock,
		nil,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilTxForCurrentBlockHandler, err)
}

func TestNewIntermediateResultsProcessor_NilEconomicsFeeHandler(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		nil,
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewIntermediateResultsProcessor_Good(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.TxBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_getShardIdsFromAddressesGood(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	sndAddr := []byte("sndAddress")
	dstAddr := []byte("dstAddress")

	sndId, dstId := irp.getShardIdsFromAddresses(sndAddr, dstAddr)
	assert.Equal(t, uint32(0), sndId, dstId)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactions(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	err = irp.AddIntermediateTransactions(nil)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsWrongType(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &transaction.Transaction{})

	err = irp.AddIntermediateTransactions(txs)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsNilSender(t *testing.T) {
	t.Parallel()

	shardC := mock.NewMultiShardsCoordinatorMock(2)
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardC,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	scr := &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: nil, Value: big.NewInt(-100), PrevTxHash: []byte("hash")}
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)

	shardC.ComputeIdCalled = func(address []byte) uint32 {
		return shardC.SelfId()
	}
	err = irp.AddIntermediateTransactions(txs)
	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsNilReceiver(t *testing.T) {
	t.Parallel()

	shardC := mock.NewMultiShardsCoordinatorMock(2)
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardC,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	scr := &smartContractResult.SmartContractResult{RcvAddr: nil, SndAddr: []byte("snd"), Value: big.NewInt(-100), PrevTxHash: []byte("hash")}
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)

	shardC.ComputeIdCalled = func(address []byte) uint32 {
		return shardC.SelfId()
	}
	err = irp.AddIntermediateTransactions(txs)
	assert.Equal(t, process.ErrNilRcvAddr, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsShardIdMismatch(t *testing.T) {
	t.Parallel()

	shardC := &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 {
			return 0
		},
		ComputeIdCalled: func(address []byte) uint32 {
			return 1
		},
	}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardC,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	scr := &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(100), PrevTxHash: []byte("hash")}
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)

	err = irp.AddIntermediateTransactions(txs)
	assert.Equal(t, process.ErrShardIdMissmatch, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsNegativeValueIntraAndCrossShard(t *testing.T) {
	t.Parallel()

	shardC := mock.NewMultiShardsCoordinatorMock(2)
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardC,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	scr := &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(-100), PrevTxHash: []byte("hash")}
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)

	shardC.ComputeIdCalled = func(address []byte) uint32 {
		return shardC.SelfId()
	}
	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	shardC.ComputeIdCalled = func(address []byte) uint32 {
		return shardC.SelfId() + 1
	}

	err = irp.AddIntermediateTransactions(txs)
	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsAddrGood(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	scr := &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: []byte("hash")}
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsAddAndRevert(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	txHash := []byte("txHash")
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: txHash, Nonce: 0})
	txs = append(txs, &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: txHash, Nonce: 1})
	txs = append(txs, &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: txHash, Nonce: 2})
	txs = append(txs, &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: txHash, Nonce: 3})
	txs = append(txs, &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: txHash, Nonce: 4})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)
	irp.mutInterResultsForBlock.Lock()
	assert.Equal(t, len(irp.mapTxToResult[string(txHash)]), len(txs))
	irp.mutInterResultsForBlock.Unlock()

	irp.RemoveProcessedResultsFor([][]byte{txHash})
	irp.mutInterResultsForBlock.Lock()
	assert.Equal(t, len(irp.mapTxToResult[string(txHash)]), 0)
	assert.Equal(t, len(irp.interResultsForBlock), 0)
	irp.mutInterResultsForBlock.Unlock()
}

func TestIntermediateResultsProcessor_CreateAllInterMiniBlocksNothingInCache(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 0, len(mbs))
}

func TestIntermediateResultsProcessor_CreateAllInterMiniBlocksNotCrossShard(t *testing.T) {
	t.Parallel()

	nrShards := 5
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	scr := &smartContractResult.SmartContractResult{RcvAddr: []byte("rcv"), SndAddr: []byte("snd"), Value: big.NewInt(0), PrevTxHash: []byte("hash")}
	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)
	txs = append(txs, scr)

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 1, len(mbs))
}

func TestIntermediateResultsProcessor_CreateAllInterMiniBlocksCrossShard(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	miniBlockTest := &block.MiniBlock{}
	for _, mb := range mbs {
		if mb.ReceiverShardID == shardCoordinator.SelfId()+1 {
			miniBlockTest = mb
		}
	}
	assert.Equal(t, 5, len(miniBlockTest.TxHashes))
}

func TestIntermediateResultsProcessor_GetNumOfCrossInterMbsAndTxsShouldWork(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}

		shardID, _ := strconv.Atoi(string(address))
		return uint32(shardID)
	}

	irp, _ := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 0, SndAddr: snd, RcvAddr: []byte("0"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 1, SndAddr: snd, RcvAddr: []byte("1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 2, SndAddr: snd, RcvAddr: []byte("1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 3, SndAddr: snd, RcvAddr: []byte("2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 4, SndAddr: snd, RcvAddr: []byte("2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 5, SndAddr: snd, RcvAddr: []byte("2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 6, SndAddr: snd, RcvAddr: []byte("3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 7, SndAddr: snd, RcvAddr: []byte("3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 8, SndAddr: snd, RcvAddr: []byte("3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{Nonce: 9, SndAddr: snd, RcvAddr: []byte("3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})

	_ = irp.AddIntermediateTransactions(txs)

	numMbs, numTxs := irp.GetNumOfCrossInterMbsAndTxs()
	assert.Equal(t, 3, numMbs)
	assert.Equal(t, 9, numTxs)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksNilBody(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := &block.Body{}
	err = irp.VerifyInterMiniBlocks(body)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyShouldpassAsNotCrossSrcFromThisShard(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{
		Type:            block.SmartContractResultBlock,
		ReceiverShardID: shardCoordinator.SelfId(),
		SenderShardID:   shardCoordinator.SelfId() + 1})

	err = irp.VerifyInterMiniBlocks(body)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyMissingMiniblock(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := &block.Body{}
	otherShard := shardCoordinator.SelfId() + 1
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock, ReceiverShardID: otherShard})

	err = irp.VerifyInterMiniBlocks(body)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyMiniBlockMissmatch(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := &block.Body{}
	otherShard := shardCoordinator.SelfId() + 1
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{Type: block.SmartContractResultBlock, ReceiverShardID: otherShard})

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}
		return otherShard
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	err = irp.VerifyInterMiniBlocks(body)
	assert.Equal(t, process.ErrMiniBlockHashMismatch, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyShouldPass(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")
	otherShard := shardCoordinator.SelfId() + 1
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}
		return otherShard
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	miniBlock := &block.MiniBlock{
		SenderShardID:   shardCoordinator.SelfId(),
		ReceiverShardID: otherShard,
		Type:            block.SmartContractResultBlock}

	for i := 0; i < len(txs); i++ {
		txHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, txs[i])
		miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
	}

	sort.Slice(miniBlock.TxHashes, func(a, b int) bool {
		return bytes.Compare(miniBlock.TxHashes[a], miniBlock.TxHashes[b]) < 0
	})

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	err = irp.VerifyInterMiniBlocks(body)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_SaveCurrentIntermediateTxToStorageShouldSave(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	putCounter := 0
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{
			PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
				if unitType == dataRetriever.UnsignedTransactionUnit {
					putCounter++
				}
				return nil
			},
		},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	err = irp.SaveCurrentIntermediateTxToStorage()
	assert.Nil(t, err)
	assert.Equal(t, len(txs), putCounter)
}

func TestIntermediateResultsProcessor_CreateMarshalizedDataNothingToMarshal(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	irp, err := NewIntermediateResultsProcessor(
		hasher,
		marshalizer,
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	// nothing to marshal
	mrsTxs, err := irp.CreateMarshalizedData(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mrsTxs))

	// nothing saved in local cacher to marshal
	mrsTxs, err = irp.CreateMarshalizedData(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mrsTxs))
}

func TestIntermediateResultsProcessor_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	irp, err := NewIntermediateResultsProcessor(
		hasher,
		marshalizer,
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txHashes := make([][]byte, 0)
	txs := make([]data.TransactionHandler, 0)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	currHash, _ := core.CalculateHash(marshalizer, hasher, txs[0])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[1])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[2])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[3])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[4])
	txHashes = append(txHashes, currHash)

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mrsTxs, err := irp.CreateMarshalizedData(txHashes)
	assert.Nil(t, err)
	assert.Equal(t, len(txs), len(mrsTxs))

	for i := 0; i < len(mrsTxs); i++ {
		unMrsScr := &smartContractResult.SmartContractResult{}
		_ = marshalizer.Unmarshal(unMrsScr, mrsTxs[i])

		assert.Equal(t, unMrsScr, txs[i])
	}
}

func TestIntermediateResultsProcessor_GetAllCurrentUsedTxs(t *testing.T) {
	t.Parallel()

	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	irp, err := NewIntermediateResultsProcessor(
		hasher,
		marshalizer,
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{},
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2"), Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: snd, Nonce: 1, Value: big.NewInt(0), PrevTxHash: []byte("txHash")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: snd, Nonce: 2, Value: big.NewInt(0), PrevTxHash: []byte("txHash")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	usedTxs := irp.GetAllCurrentFinishedTxs()
	assert.Equal(t, 4, len(usedTxs))
}

func TestIntermediateResultsProcessor_SplitMiniBlocksIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	var gasLimit uint64
	nrShards := 5
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	irp, _ := NewIntermediateResultsProcessor(
		hasher,
		marshalizer,
		shardCoordinator,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
		&mock.TxForCurrentBlockStub{},
		&mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return gasLimit
			},
		},
	)

	tx1 := transaction.Transaction{Nonce: 0, GasLimit: 100}
	tx2 := transaction.Transaction{Nonce: 1, GasLimit: 100}
	tx3 := transaction.Transaction{Nonce: 2, GasLimit: 100}
	tx4 := transaction.Transaction{Nonce: 3, GasLimit: 100}
	tx5 := transaction.Transaction{Nonce: 4, GasLimit: 100}
	irp.interResultsForBlock["hash1"] = &txInfo{tx: &tx1}
	irp.interResultsForBlock["hash2"] = &txInfo{tx: &tx2}
	irp.interResultsForBlock["hash3"] = &txInfo{tx: &tx3}
	irp.interResultsForBlock["hash4"] = &txInfo{tx: &tx4}
	irp.interResultsForBlock["hash5"] = &txInfo{tx: &tx5}

	miniBlocks := make([]*block.MiniBlock, 0)

	mb1 := block.MiniBlock{
		ReceiverShardID: 1,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
	}
	miniBlocks = append(miniBlocks, &mb1)

	mb2 := block.MiniBlock{
		ReceiverShardID: 2,
		TxHashes:        [][]byte{[]byte("hash3"), []byte("hash4"), []byte("hash5"), []byte("hash6")},
	}
	miniBlocks = append(miniBlocks, &mb2)

	gasLimit = 300
	splitMiniBlocks := irp.splitMiniBlocksIfNeeded(miniBlocks)
	assert.Equal(t, 2, len(splitMiniBlocks))

	gasLimit = 199
	splitMiniBlocks = irp.splitMiniBlocksIfNeeded(miniBlocks)
	assert.Equal(t, 5, len(splitMiniBlocks))
}
