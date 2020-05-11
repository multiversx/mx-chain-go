package txcache

import (
	"encoding/binary"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

const oneMilion = 1000000
const oneTrilion = oneMilion * oneMilion
const delta = 0.00000001

func toMicroERD(erd uint64) uint64 {
	return erd * 1000000
}

func kBToBytes(kB float32) uint64 {
	return uint64(kB * 1000)
}

func (cache *TxCache) getListForSender(sender string) *txListForSender {
	return cache.txListBySender.testGetListForSender(sender)
}

func (sendersMap *txListBySenderMap) testGetListForSender(sender string) *txListForSender {
	list, ok := sendersMap.getListForSender(sender)
	if !ok {
		panic("sender not in cache")
	}

	return list
}

func (cache *TxCache) getRawScoreOfSender(sender string) float64 {
	return cache.getListForSender(sender).computeRawScore()
}

func (cache *TxCache) getNumFailedSelectionsOfSender(sender string) int {
	return int(cache.getListForSender(sender).numFailedSelections.Get())
}

func (cache *TxCache) setNumFailedSelectionsOfSender(sender string, numFailed int) {
	cache.getListForSender(sender).numFailedSelections.Set(int64(numFailed))
}

func (cache *TxCache) isSenderSweepable(sender string) bool {
	for _, item := range cache.sweepingListOfSenders {
		if item.sender == sender {
			return true
		}
	}

	return false
}

func addManyTransactionsWithUniformDistribution(cache *TxCache, nSenders int, nTransactionsPerSender int) {
	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := createFakeSenderAddress(senderTag)

		for txNonce := nTransactionsPerSender; txNonce > 0; txNonce-- {
			txHash := createFakeTxHash(sender, txNonce)
			tx := createTx(txHash, string(sender), uint64(txNonce))
			cache.AddTx(tx)
		}
	}
}

func createTx(hash []byte, sender string, nonce uint64) *WrappedTransaction {
	tx := &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}

	return &WrappedTransaction{
		Tx:     tx,
		TxHash: hash,
	}
}

func createTxWithParams(hash []byte, sender string, nonce uint64, dataLength uint64, gasLimit uint64, gasPrice uint64) *WrappedTransaction {
	payloadLength := int(dataLength) - int(estimatedSizeOfBoundedTxFields)
	if payloadLength < 0 {
		panic("createTxWithData(): invalid length for dummy tx")
	}

	tx := &transaction.Transaction{
		SndAddr:  []byte(sender),
		Nonce:    nonce,
		Data:     make([]byte, payloadLength),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
	}

	return &WrappedTransaction{
		Tx:     tx,
		TxHash: hash,
	}
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

func measureWithStopWatch(b *testing.B, function func()) {
	sw := core.NewStopWatch()
	sw.Start("time")
	function()
	sw.Stop("time")

	duration := sw.GetMeasurementsMap()["time"]
	b.ReportMetric(duration, "time@stopWatch")
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
