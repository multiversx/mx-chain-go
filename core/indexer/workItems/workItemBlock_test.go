package workItems_test

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/require"
)

func generateTxs(numTxs int) map[string]data.TransactionHandler {
	txs := make(map[string]data.TransactionHandler, numTxs)
	for i := 0; i < numTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:     uint64(i),
			Value:     big.NewInt(int64(i)),
			RcvAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			SndAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			GasPrice:  10000000,
			GasLimit:  1000,
			Data:      []byte("dasjdksakjdksajdjksajkdjkasjdksajkdasjdksakjdksajdjksajkdjkasjdksajkdasjdksakjdksajdjksajkdjkasjdksajk"),
			Signature: []byte("randomSignatureasdasldkasdsahjgdlhjaskldsjkaldjklasjkdjskladjkl;sajkl"),
		}
		txs[fmt.Sprintf("%d", i)] = tx
	}

	return txs
}

func TestItemBlock_SaveHeaderShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(header data.HeaderHandler, signersIndexes []uint64, body *dataBlock.Body, notarizedHeadersHashes []string, txsSize int) error {
				return localErr
			},
		},
		&mock.MarshalizerMock{},
		true,
		&dataBlock.Body{
			MiniBlocks: dataBlock.MiniBlockSlice{{}},
		},
		&dataBlock.Header{},
		nil,
		[]uint64{},
		[]string{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.Equal(t, localErr, err)
}

func TestItemBlock_SaveNoMiniblocksShoulCallSaveHeader(t *testing.T) {
	countCalled := 0
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(header data.HeaderHandler, signersIndexes []uint64, body *dataBlock.Body, notarizedHeadersHashes []string, txsSize int) error {
				countCalled++
				return nil
			},
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) (map[string]bool, error) {
				countCalled++
				return nil, nil
			},
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error {
				countCalled++
				return nil
			},
		},
		&mock.MarshalizerMock{},
		true,
		&dataBlock.Body{},
		&dataBlock.Header{},
		nil,
		[]uint64{},
		[]string{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.Equal(t, 1, countCalled)
}

func TestItemBlock_SaveMiniblocksShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) (map[string]bool, error) {
				return nil, localErr
			},
		},
		&mock.MarshalizerMock{},
		true,
		&dataBlock.Body{
			MiniBlocks: dataBlock.MiniBlockSlice{{}},
		},
		&dataBlock.Header{},
		nil,
		[]uint64{},
		[]string{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.Equal(t, localErr, err)
}

func TestItemBlock_ShouldNotSaveTransaction(t *testing.T) {
	called := false
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error {
				called = true
				return nil
			},
		},
		&mock.MarshalizerMock{},
		false,
		&dataBlock.Body{
			MiniBlocks: dataBlock.MiniBlockSlice{{}},
		},
		&dataBlock.Header{},
		nil,
		[]uint64{},
		[]string{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.False(t, called)
}

func TestItemBlock_SaveTransactionsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error {
				return localErr
			},
		},
		&mock.MarshalizerMock{},
		true,
		&dataBlock.Body{
			MiniBlocks: dataBlock.MiniBlockSlice{{}},
		},
		&dataBlock.Header{},
		nil,
		[]uint64{},
		[]string{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.Equal(t, localErr, err)
}

func TestItemBlock_SaveShouldWork(t *testing.T) {
	countCalled := 0
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(header data.HeaderHandler, signersIndexes []uint64, body *dataBlock.Body, notarizedHeadersHashes []string, txsSize int) error {
				countCalled++
				return nil
			},
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) (map[string]bool, error) {
				countCalled++
				return nil, nil
			},
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error {
				countCalled++
				return nil
			},
		},
		&mock.MarshalizerMock{},
		true,
		&dataBlock.Body{
			MiniBlocks: dataBlock.MiniBlockSlice{{}},
		},
		&dataBlock.Header{},
		nil,
		[]uint64{},
		[]string{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.Equal(t, 3, countCalled)
}

func TestComputeSizeOfTxsDuration(t *testing.T) {
	res := testing.Benchmark(benchmarkComputeSizeOfTxsDuration)

	fmt.Println("Time to calculate size of txs :", time.Duration(res.NsPerOp()))
}

func benchmarkComputeSizeOfTxsDuration(b *testing.B) {
	numTxs := 20000
	txs := generateTxs(numTxs)
	gogoMarsh := &marshal.GogoProtoMarshalizer{}

	for i := 0; i < b.N; i++ {
		workItems.ComputeSizeOfTxs(gogoMarsh, txs)
	}
}

func TestComputeSizeOfTxs(t *testing.T) {
	const kb = 1024
	numTxs := 20000

	txs := generateTxs(numTxs)
	gogoMarsh := &marshal.GogoProtoMarshalizer{}
	lenTxs := workItems.ComputeSizeOfTxs(gogoMarsh, txs)

	keys := reflect.ValueOf(txs).MapKeys()
	oneTxBytes, _ := gogoMarsh.Marshal(txs[keys[0].String()])
	oneTxSize := len(oneTxBytes)
	expectedSize := numTxs * oneTxSize
	expectedSizeDeltaPlus := expectedSize + int(0.01*float64(expectedSize))
	expectedSizeDeltaMinus := expectedSize - int(0.01*float64(expectedSize))

	require.Greater(t, lenTxs, expectedSizeDeltaMinus)
	require.Less(t, lenTxs, expectedSizeDeltaPlus)
	fmt.Printf("Size of %d transactions : %d Kbs \n", numTxs, lenTxs/kb)
}
