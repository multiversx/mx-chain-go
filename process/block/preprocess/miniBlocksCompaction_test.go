package preprocess

import (
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func TestMiniBlocksCompaction_CompactAndExpandMiniBlocksShouldResultTheSameMiniBlocks(t *testing.T) {
	t.Parallel()

	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	requestTransaction := func(shardID uint32, txHashes [][]byte) {}
	txs, _ := NewTransactionPreprocessor(
		txPool,
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint64) error {
				return nil
			},
		},
		mock.NewMultiShardsCoordinatorMock(2),
		&mock.AccountsStub{},
		requestTransaction,
		feeHandlerMock(),
		miniBlocksCompacterMock(),
	)

	keygen := signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())
	_, accPk := keygen.GeneratePair()
	pkBytes, _ := accPk.ToByteArray()

	strCache00 := process.ShardCacherIdentifier(0, 0)
	strCache01 := process.ShardCacherIdentifier(0, 1)

	txHashesInMb1 := [][]byte{[]byte("tx00"), []byte("tx01"), []byte("tx02")}
	txHashesInMb2 := [][]byte{[]byte("tx10"), []byte("tx11"), []byte("tx12")}
	txHashesInMb3 := [][]byte{[]byte("tx20"), []byte("tx21"), []byte("tx22")}
	txHashesInMb4 := [][]byte{[]byte("tx30"), []byte("tx31"), []byte("tx32")}

	mapHashesAndTxs := map[string]data.TransactionHandler{
		string(txHashesInMb1[0]): &transaction.Transaction{Nonce: 0, SndAddr: pkBytes},
		string(txHashesInMb1[1]): &transaction.Transaction{Nonce: 1, SndAddr: pkBytes},
		string(txHashesInMb1[2]): &transaction.Transaction{Nonce: 2, SndAddr: pkBytes},
		string(txHashesInMb2[0]): &transaction.Transaction{Nonce: 3, SndAddr: pkBytes},
		string(txHashesInMb2[1]): &transaction.Transaction{Nonce: 4, SndAddr: pkBytes},
		string(txHashesInMb2[2]): &transaction.Transaction{Nonce: 5, SndAddr: pkBytes},
		string(txHashesInMb3[0]): &transaction.Transaction{Nonce: 6, SndAddr: pkBytes},
		string(txHashesInMb3[1]): &transaction.Transaction{Nonce: 7, SndAddr: pkBytes},
		string(txHashesInMb3[2]): &transaction.Transaction{Nonce: 8, SndAddr: pkBytes},
		string(txHashesInMb4[0]): &transaction.Transaction{Nonce: 9, SndAddr: pkBytes},
		string(txHashesInMb4[1]): &transaction.Transaction{Nonce: 10, SndAddr: pkBytes},
		string(txHashesInMb4[2]): &transaction.Transaction{Nonce: 11, SndAddr: pkBytes},
	}

	txPool.AddData(txHashesInMb1[0], mapHashesAndTxs[string(txHashesInMb1[0])], strCache00)
	txPool.AddData(txHashesInMb1[1], mapHashesAndTxs[string(txHashesInMb1[1])], strCache00)
	txPool.AddData(txHashesInMb1[2], mapHashesAndTxs[string(txHashesInMb1[2])], strCache00)
	mb1 := block.MiniBlock{
		TxHashes:        txHashesInMb1,
		ReceiverShardID: 0,
		SenderShardID:   0,
		Type:            0,
	}

	txPool.AddData(txHashesInMb2[0], mapHashesAndTxs[string(txHashesInMb2[0])], strCache01)
	txPool.AddData(txHashesInMb2[1], mapHashesAndTxs[string(txHashesInMb2[1])], strCache01)
	txPool.AddData(txHashesInMb2[2], mapHashesAndTxs[string(txHashesInMb2[2])], strCache01)
	mb2 := block.MiniBlock{
		TxHashes:        txHashesInMb2,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	txPool.AddData(txHashesInMb3[0], mapHashesAndTxs[string(txHashesInMb3[0])], strCache00)
	txPool.AddData(txHashesInMb3[1], mapHashesAndTxs[string(txHashesInMb3[1])], strCache00)
	txPool.AddData(txHashesInMb3[2], mapHashesAndTxs[string(txHashesInMb3[2])], strCache00)
	mb3 := block.MiniBlock{
		TxHashes:        txHashesInMb3,
		ReceiverShardID: 0,
		SenderShardID:   0,
		Type:            0,
	}

	txPool.AddData(txHashesInMb4[0], mapHashesAndTxs[string(txHashesInMb4[0])], strCache01)
	txPool.AddData(txHashesInMb4[1], mapHashesAndTxs[string(txHashesInMb4[1])], strCache01)
	txPool.AddData(txHashesInMb4[2], mapHashesAndTxs[string(txHashesInMb4[2])], strCache01)
	mb4 := block.MiniBlock{
		TxHashes:        txHashesInMb4,
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
	}

	_ = txs.ProcessMiniBlock(&mb1, haveTimeTrue, 0)
	_ = txs.ProcessMiniBlock(&mb2, haveTimeTrue, 0)
	_ = txs.ProcessMiniBlock(&mb3, haveTimeTrue, 0)
	_ = txs.ProcessMiniBlock(&mb4, haveTimeTrue, 0)

	mbsOrig := block.MiniBlockSlice{}
	mbsOrig = append(mbsOrig, &mb1, &mb2, &mb3, &mb4)

	mbsValues := make([]block.MiniBlock, 0)
	for _, mb := range mbsOrig {
		mbsValues = append(mbsValues, *mb)
	}

	compactedMbs := txs.miniBlocksCompacter.Compact(mbsOrig, mapHashesAndTxs)
	expandedMbs, err := txs.miniBlocksCompacter.Expand(compactedMbs, mapHashesAndTxs)
	assert.Nil(t, err)

	assert.Equal(t, len(mbsValues), len(expandedMbs))
	for i := 0; i < len(mbsValues); i++ {
		assert.True(t, reflect.DeepEqual(mbsValues[i], *expandedMbs[i]))
	}
}
