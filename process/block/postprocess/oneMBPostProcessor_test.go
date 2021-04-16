package postprocess

import (
	"bytes"
	"sort"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewOneMBPostProcessor_NilHasher(t *testing.T) {
	t.Parallel()

	irp, err := NewOneMiniBlockPostProcessor(
		nil,
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewOneMBPostProcessor_NilMarshalizer(t *testing.T) {
	t.Parallel()

	irp, err := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewOneMBPostProcessor_NilShardCoord(t *testing.T) {
	t.Parallel()

	irp, err := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewOneMBPostProcessor_NilStorer(t *testing.T) {
	t.Parallel()

	irp, err := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		nil,
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewOneMBPostProcessor_NilEconomicsFeeHandler(t *testing.T) {
	t.Parallel()

	irp, err := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		nil,
	)

	assert.Nil(t, irp)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewOneMBPostProcessor_OK(t *testing.T) {
	t.Parallel()

	irp, err := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, irp)
}

func TestOneMBPostProcessor_CreateAllInterMiniBlocks(t *testing.T) {
	t.Parallel()

	irp, _ := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 0, len(mbs))
}

func TestOneMBPostProcessor_CreateAllInterMiniBlocksOneMinBlock(t *testing.T) {
	t.Parallel()

	irp, _ := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &transaction.Transaction{})
	txs = append(txs, &transaction.Transaction{})

	err := irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	mbs := irp.CreateAllInterMiniBlocks()
	assert.Equal(t, 1, len(mbs))
}

func TestOneMBPostProcessor_VerifyNilBody(t *testing.T) {
	t.Parallel()

	irp, _ := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	err := irp.VerifyInterMiniBlocks(&block.Body{})
	assert.Nil(t, err)
}

func TestOneMBPostProcessor_VerifyTooManyBlock(t *testing.T) {
	t.Parallel()

	irp, _ := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr1")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr2")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr3")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr4")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr5")})

	err := irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	miniBlock := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 0,
		Type:            block.TxBlock}

	for i := 0; i < len(txs); i++ {
		txHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, txs[i])
		miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
	}

	sort.Slice(miniBlock.TxHashes, func(a, b int) bool {
		return bytes.Compare(miniBlock.TxHashes[a], miniBlock.TxHashes[b]) < 0
	})

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	err = irp.VerifyInterMiniBlocks(body)
	assert.Equal(t, process.ErrTooManyReceiptsMiniBlocks, err)
}

func TestOneMBPostProcessor_VerifyNilMiniBlocks(t *testing.T) {
	t.Parallel()

	irp, _ := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	miniBlock := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 0,
		Type:            block.TxBlock}
	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, miniBlock)

	err := irp.VerifyInterMiniBlocks(body)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestOneMBPostProcessor_VerifyOk(t *testing.T) {
	t.Parallel()

	irp, _ := NewOneMiniBlockPostProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(5),
		&mock.ChainStorerMock{},
		block.TxBlock,
		dataRetriever.TransactionUnit,
		&mock.FeeHandlerStub{},
	)

	txs := make([]data.TransactionHandler, 0)
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr1")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr2")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr3")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr4")})
	txs = append(txs, &transaction.Transaction{SndAddr: []byte("snd"), RcvAddr: []byte("recvaddr5")})

	err := irp.AddIntermediateTransactions(txs)
	assert.Nil(t, err)

	miniBlock := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 0,
		Type:            block.TxBlock}

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
