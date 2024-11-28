package txcache

import (
	"encoding/binary"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
)

const oneMilion = 1000000
const oneBillion = oneMilion * 1000
const estimatedSizeOfBoundedTxFields = uint64(128)

// The GitHub Actions runners are (extremely) slow.
const selectionLoopMaximumDuration = 30 * time.Second

func (cache *TxCache) areInternalMapsConsistent() bool {
	internalMapByHash := cache.txByHash
	internalMapBySender := cache.txListBySender

	senders := internalMapBySender.getSenders()
	numInMapByHash := len(internalMapByHash.keys())
	numInMapBySender := 0
	numMissingInMapByHash := 0

	for _, sender := range senders {
		numInMapBySender += int(sender.countTx())

		for _, hash := range sender.getTxsHashes() {
			_, ok := internalMapByHash.getTx(string(hash))
			if !ok {
				numMissingInMapByHash++
			}
		}
	}

	isFine := (numInMapByHash == numInMapBySender) && (numMissingInMapByHash == 0)
	return isFine
}

func (cache *TxCache) getHashesForSender(sender string) []string {
	return cache.getListForSender(sender).getTxHashesAsStrings()
}

func (cache *TxCache) getListForSender(sender string) *txListForSender {
	return cache.txListBySender.testGetListForSender(sender)
}

func (txMap *txListBySenderMap) testGetListForSender(sender string) *txListForSender {
	list, ok := txMap.getListForSender(sender)
	if !ok {
		panic("sender not in cache")
	}

	return list
}

func (listForSender *txListForSender) getTxHashesAsStrings() []string {
	hashes := listForSender.getTxsHashes()
	return hashesAsStrings(hashes)
}

func (listForSender *txListForSender) getTxsHashes() [][]byte {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	result := make([][]byte, 0, listForSender.countTx())

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		result = append(result, value.TxHash)
	}

	return result
}

func hashesAsStrings(hashes [][]byte) []string {
	result := make([]string, len(hashes))
	for i := 0; i < len(hashes); i++ {
		result[i] = string(hashes[i])
	}

	return result
}

func hashesAsBytes(hashes []string) [][]byte {
	result := make([][]byte, len(hashes))
	for i := 0; i < len(hashes); i++ {
		result[i] = []byte(hashes[i])
	}

	return result
}

func addManyTransactionsWithUniformDistribution(cache *TxCache, nSenders int, nTransactionsPerSender int) {
	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := createFakeSenderAddress(senderTag)

		for nonce := nTransactionsPerSender - 1; nonce >= 0; nonce-- {
			transactionHash := createFakeTxHash(sender, nonce)
			gasPrice := oneBillion + rand.Intn(3*oneBillion)
			transaction := createTx(transactionHash, string(sender), uint64(nonce)).withGasPrice(uint64(gasPrice))

			cache.AddTx(transaction)
		}
	}
}

func createBunchesOfTransactionsWithUniformDistribution(nSenders int, nTransactionsPerSender int) []bunchOfTransactions {
	bunches := make([]bunchOfTransactions, 0, nSenders)
	host := txcachemocks.NewMempoolHostMock()

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		bunch := make(bunchOfTransactions, 0, nTransactionsPerSender)
		sender := createFakeSenderAddress(senderTag)

		for nonce := 0; nonce < nTransactionsPerSender; nonce++ {
			transactionHash := createFakeTxHash(sender, nonce)
			gasPrice := oneBillion + rand.Intn(3*oneBillion)
			transaction := createTx(transactionHash, string(sender), uint64(nonce)).withGasPrice(uint64(gasPrice))
			transaction.precomputeFields(host)

			bunch = append(bunch, transaction)
		}

		bunches = append(bunches, bunch)
	}

	return bunches
}

func createTx(hash []byte, sender string, nonce uint64) *WrappedTransaction {
	tx := &transaction.Transaction{
		SndAddr:  []byte(sender),
		Nonce:    nonce,
		GasLimit: 50000,
		GasPrice: oneBillion,
	}

	return &WrappedTransaction{
		Tx:     tx,
		TxHash: hash,
		Size:   int64(estimatedSizeOfBoundedTxFields),
	}
}

func (wrappedTx *WrappedTransaction) withSize(size uint64) *WrappedTransaction {
	dataLength := size - estimatedSizeOfBoundedTxFields
	tx := wrappedTx.Tx.(*transaction.Transaction)
	tx.Data = make([]byte, dataLength)
	wrappedTx.Size = int64(size)
	return wrappedTx
}

func (wrappedTx *WrappedTransaction) withData(data []byte) *WrappedTransaction {
	tx := wrappedTx.Tx.(*transaction.Transaction)
	tx.Data = data
	wrappedTx.Size = int64(len(data)) + int64(estimatedSizeOfBoundedTxFields)
	return wrappedTx
}

func (wrappedTx *WrappedTransaction) withDataLength(dataLength int) *WrappedTransaction {
	tx := wrappedTx.Tx.(*transaction.Transaction)
	tx.Data = make([]byte, dataLength)
	wrappedTx.Size = int64(dataLength) + int64(estimatedSizeOfBoundedTxFields)
	return wrappedTx
}

func (wrappedTx *WrappedTransaction) withGasPrice(gasPrice uint64) *WrappedTransaction {
	tx := wrappedTx.Tx.(*transaction.Transaction)
	tx.GasPrice = gasPrice
	return wrappedTx
}

func (wrappedTx *WrappedTransaction) withGasLimit(gasLimit uint64) *WrappedTransaction {
	tx := wrappedTx.Tx.(*transaction.Transaction)
	tx.GasLimit = gasLimit
	return wrappedTx
}

func (wrappedTx *WrappedTransaction) withValue(value *big.Int) *WrappedTransaction {
	tx := wrappedTx.Tx.(*transaction.Transaction)
	tx.Value = value
	return wrappedTx
}

func createFakeSenderAddress(senderTag int) []byte {
	bytes := make([]byte, 32)
	binary.LittleEndian.PutUint64(bytes, uint64(senderTag))
	binary.LittleEndian.PutUint64(bytes[24:], uint64(senderTag))
	return bytes
}

func createFakeTxHash(fakeSenderAddress []byte, nonce int) []byte {
	bytes := make([]byte, 32)
	copy(bytes, fakeSenderAddress)
	binary.LittleEndian.PutUint64(bytes[8:], uint64(nonce))
	binary.LittleEndian.PutUint64(bytes[16:], uint64(nonce))
	return bytes
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
// Reference: https://stackoverflow.com/a/32843750/1475331
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
