package preprocess

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

var r *rand.Rand
var mutex sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func TestSortTxByNonce_EmptyCacherShouldReturnEmpty(t *testing.T) {
	t.Parallel()
	cacher, _ := storageUnit.NewCache(storageUnit.LRUCache, 100, 1)
	transactions, txHashes := sortTxByNonce(cacher)
	require.Equal(t, 0, len(transactions))
	require.Equal(t, 0, len(txHashes))
}

func TestSortTxByNonce_OneTxShouldWork(t *testing.T) {
	t.Parallel()
	cacher, _ := storageUnit.NewCache(storageUnit.LRUCache, 100, 1)
	hash, tx := createRandTx(r)
	cacher.HasOrAdd(hash, tx)
	transactions, txHashes := sortTxByNonce(cacher)
	require.Equal(t, 1, len(transactions))
	require.Equal(t, 1, len(txHashes))
	require.True(t, hashInSlice(hash, txHashes))
	require.True(t, txInSlice(tx, transactions))
}

func createRandTx(rand *rand.Rand) ([]byte, *transaction.Transaction) {
	mutex.Lock()
	nonce := rand.Uint64()
	mutex.Unlock()
	tx := &transaction.Transaction{
		Nonce: nonce,
	}
	marshalizer := &mock.MarshalizerMock{}
	buffTx, _ := marshalizer.Marshal(tx)
	hash := mock.HasherMock{}.Compute(string(buffTx))
	return hash, tx
}

func hashInSlice(hash []byte, hashes [][]byte) bool {
	for _, h := range hashes {
		if bytes.Equal(h, hash) {
			return true
		}
	}
	return false
}

func txInSlice(tx *transaction.Transaction, transactions []data.TransactionHandler) bool {
	for _, t := range transactions {
		if reflect.DeepEqual(tx, t) {
			return true
		}
	}
	return false
}

func TestSortTxByNonce_MoreTransactionsShouldRetSameSize(t *testing.T) {
	t.Parallel()
	cache, genTransactions, _ := genCacherTransactionsHashes(100)
	transactions, txHashes := sortTxByNonce(cache)
	require.Equal(t, len(genTransactions), len(transactions))
	require.Equal(t, len(genTransactions), len(txHashes))
}

func TestSortTxByNonce_MoreTransactionsShouldContainSameElements(t *testing.T) {
	t.Parallel()
	cache, genTransactions, genHashes := genCacherTransactionsHashes(100)
	transactions, txHashes := sortTxByNonce(cache)
	for i := 0; i < len(genTransactions); i++ {
		require.True(t, hashInSlice(genHashes[i], txHashes))
		require.True(t, txInSlice(genTransactions[i], transactions))
	}
}

func TestSortTxByNonce_MoreTransactionsShouldContainSortedElements(t *testing.T) {
	t.Parallel()
	cache, _, _ := genCacherTransactionsHashes(100)
	transactions, _ := sortTxByNonce(cache)
	lastNonce := uint64(0)
	for i := 0; i < len(transactions); i++ {
		tx := transactions[i]
		require.True(t, lastNonce <= tx.GetNonce())
		fmt.Println(tx.GetNonce())
		lastNonce = tx.GetNonce()
	}
}

func TestSortTxByNonce_TransactionsWithSameNonceShouldGetSorted(t *testing.T) {
	t.Parallel()
	transactions := []*transaction.Transaction{
		{Nonce: 1, Signature: []byte("sig1")},
		{Nonce: 2, Signature: []byte("sig2")},
		{Nonce: 1, Signature: []byte("sig3")},
		{Nonce: 2, Signature: []byte("sig4")},
		{Nonce: 3, Signature: []byte("sig5")},
	}
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, uint32(len(transactions)), 1)
	for _, tx := range transactions {
		marshalizer := &mock.MarshalizerMock{}
		buffTx, _ := marshalizer.Marshal(tx)
		hash := mock.HasherMock{}.Compute(string(buffTx))

		cache.Put(hash, tx)
	}
	sortedTxs, _ := sortTxByNonce(cache)
	lastNonce := uint64(0)
	for i := 0; i < len(sortedTxs); i++ {
		tx := sortedTxs[i]
		require.True(t, lastNonce <= tx.GetNonce())
		lastNonce = tx.GetNonce()
	}
	require.Equal(t, len(sortedTxs), len(transactions))
	//test if one transaction from transactions might not be in sortedTx
	for _, tx := range transactions {
		found := false
		for _, stx := range sortedTxs {
			if reflect.DeepEqual(tx, stx) {
				found = true
				break
			}
		}
		if !found {
			require.Fail(t, "Not found tx in sorted slice for sig: "+string(tx.Signature))
		}
	}
}

func BenchmarkSortTxByNonce1(b *testing.B) {
	cache, _, _ := genCacherTransactionsHashes(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sortTxByNonce(cache)
	}
}

func genCacherTransactionsHashes(noOfTx int) (storage.Cacher, []*transaction.Transaction, [][]byte) {
	cacher, _ := storageUnit.NewCache(storageUnit.LRUCache, uint32(noOfTx), 1)
	genHashes := make([][]byte, 0)
	genTransactions := make([]*transaction.Transaction, 0)
	for i := 0; i < noOfTx; i++ {
		hash, tx := createRandTx(r)
		cacher.HasOrAdd(hash, tx)

		genHashes = append(genHashes, hash)
		genTransactions = append(genTransactions, tx)
	}
	return cacher, genTransactions, genHashes
}
