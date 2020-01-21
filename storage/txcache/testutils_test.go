package txcache

import (
	"encoding/binary"
	"testing"

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

func (cache *TxCache) getRawScoreOfSender(sender string) float64 {
	list, ok := cache.txListBySender.getListForSender(sender)
	if !ok {
		panic("sender not in cache")
	}

	rawScore := list.computeRawScore()
	return rawScore
}

func addManyTransactionsWithUniformDistribution(cache *TxCache, nSenders int, nTransactionsPerSender int) {
	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := createFakeSenderAddress(senderTag)

		for txNonce := nTransactionsPerSender; txNonce > 0; txNonce-- {
			txHash := createFakeTxHash(sender, txNonce)
			tx := createTx(string(sender), uint64(txNonce))
			cache.AddTx([]byte(txHash), tx)
		}
	}
}

func createTx(sender string, nonce uint64) *transaction.Transaction {
	return &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
	}
}

func createTxWithParams(sender string, nonce uint64, dataLength uint64, gasLimit uint64, gasPrice uint64) *transaction.Transaction {
	payloadLength := int(dataLength) - int(estimatedSizeOfBoundedTxFields)
	if payloadLength < 0 {
		panic("createTxWithData(): invalid length for dummy tx")
	}

	return &transaction.Transaction{
		SndAddr:  []byte(sender),
		Nonce:    nonce,
		Data:     make([]byte, payloadLength),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
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
