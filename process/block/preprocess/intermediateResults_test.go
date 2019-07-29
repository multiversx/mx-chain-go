package preprocess

import (
	"bytes"
	"sort"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewIntermediateResultsProcessor_NilHashes(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		nil,
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		block.TxBlock,
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
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		block.TxBlock,
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
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		block.TxBlock,
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateResultsProcessor_NilAddressConverter(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		nil,
		&mock.ChainStorerMock{},
		block.TxBlock,
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewIntermediateResultsProcessor_NilStorer(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AddressConverterMock{},
		nil,
		block.TxBlock,
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateResultsProcessor_Good(t *testing.T) {
	t.Parallel()

	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.AddressConverterMock{},
		&mock.ChainStorerMock{},
		block.TxBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_getShardIdsFromAddressesErrAdrConv(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	sndAddr := []byte("sndAddress")
	dstAddr := []byte("dstAddress")

	adrConv.Fail = true

	sndId, dstId, err := irp.getShardIdsFromAddresses(sndAddr, dstAddr)
	assert.Equal(t, uint32(nrShards), sndId, dstId)
	assert.NotNil(t, err)
}

func TestIntermediateResultsProcessor_getShardIdsFromAddressesGood(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	sndAddr := []byte("sndAddress")
	dstAddr := []byte("dstAddress")

	sndId, dstId, err := irp.getShardIdsFromAddresses(sndAddr, dstAddr)
	assert.Equal(t, uint32(0), sndId, dstId)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactions(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	err = irp.AddIntermediateTransactions(nil)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsWrongType(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &transaction.Transaction{})

	err = irp.AddIntermediateTransactions(txs)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsAddrConvError(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{})

	adrConv.Fail = true

	err = irp.AddIntermediateTransactions(txs)
	assert.NotNil(t, err)
}

func TestIntermediateResultsProcessor_AddIntermediateTransactionsAddrGood(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_CreateAllInterMiniBlocksNothingInCache(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 0, len(mbs))
}

func TestIntermediateResultsProcessor_CreateAllInterMiniBlocksNotCrossShard(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(uint32(nrShards)),
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})
	txs = append(txs, &smartContractResult.SmartContractResult{})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 0, len(mbs))
}

func TestIntermediateResultsProcessor_CreateAllInterMiniBlocksCrossShard(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 5, len(mbs[shardCoordinator.SelfId()+1].TxHashes))
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksNilBody(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	err = irp.VerifyInterMiniBlocks(nil)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyShouldpassAsNotCrossSrcFromThisShard(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := block.Body{}
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock, ReceiverShardID: shardCoordinator.SelfId()})

	err = irp.VerifyInterMiniBlocks(body)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyMissingMiniblock(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := block.Body{}
	otherShard := shardCoordinator.SelfId() + 1
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock, ReceiverShardID: otherShard})

	err = irp.VerifyInterMiniBlocks(body)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyMiniBlockMissmatch(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	body := block.Body{}
	otherShard := shardCoordinator.SelfId() + 1
	body = append(body, &block.MiniBlock{Type: block.SmartContractResultBlock, ReceiverShardID: otherShard})

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), snd) {
			return shardCoordinator.SelfId()
		}
		return otherShard
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	err = irp.VerifyInterMiniBlocks(body)
	assert.Equal(t, process.ErrMiniBlockHashMismatch, err)
}

func TestIntermediateResultsProcessor_VerifyInterMiniBlocksBodyShouldPass(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")
	otherShard := shardCoordinator.SelfId() + 1
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), snd) {
			return shardCoordinator.SelfId()
		}
		return otherShard
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5")})

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

	body := block.Body{}
	body = append(body, miniBlock)

	err = irp.VerifyInterMiniBlocks(body)
	assert.Nil(t, err)
}

func TestIntermediateResultsProcessor_SaveCurrentIntermediateTxToStorageShouldSave(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	putCounter := 0
	irp, err := NewIntermediateResultsProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{
			PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
				if unitType == dataRetriever.UnsignedTransactionUnit {
					putCounter++
				}
				return nil
			},
		},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4")})
	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5")})

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	err = irp.SaveCurrentIntermediateTxToStorage()
	assert.Nil(t, err)
	assert.Equal(t, len(txs), putCounter)
}

func TestIntermediateResultsProcessor_CreateMarshalizedDataNothingToMarshal(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	irp, err := NewIntermediateResultsProcessor(
		hasher,
		marshalizer,
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	// nothing to marshal
	mrsTxs, err := irp.CreateMarshalizedData(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mrsTxs))

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("bad1"), []byte("bad2"), []byte("bad3"))

	// nothing saved in local cacher to marshal
	mrsTxs, err = irp.CreateMarshalizedData(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mrsTxs))
}

func TestIntermediateResultsProcessor_CreateMarshalizedData(t *testing.T) {
	t.Parallel()

	nrShards := 5
	adrConv := &mock.AddressConverterMock{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(uint32(nrShards))
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	irp, err := NewIntermediateResultsProcessor(
		hasher,
		marshalizer,
		shardCoordinator,
		adrConv,
		&mock.ChainStorerMock{},
		block.SmartContractResultBlock,
	)

	assert.NotNil(t, irp)
	assert.Nil(t, err)

	snd := []byte("snd")

	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), snd) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	txHashes := make([][]byte, 0)
	txs := make([]data.TransactionHandler, 0)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr1")})
	currHash, _ := core.CalculateHash(marshalizer, hasher, txs[0])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr2")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[1])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr3")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[2])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr4")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[3])
	txHashes = append(txHashes, currHash)

	txs = append(txs, &smartContractResult.SmartContractResult{SndAddr: snd, RcvAddr: []byte("recvaddr5")})
	currHash, _ = core.CalculateHash(marshalizer, hasher, txs[4])
	txHashes = append(txHashes, currHash)

	err = irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mrsTxs, err := irp.CreateMarshalizedData(txHashes)
	assert.Nil(t, err)
	assert.Equal(t, len(txs), len(mrsTxs))
}
