package debug

import (
	"runtime/debug"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("debug")

// DetectorForProcessBadTransaction -
var DetectorForProcessBadTransaction = NewDoubleTransactionsDetector("DetectorForProcessBadTransaction")

// DetectorProcessBlockTransaction -
var DetectorProcessBlockTransaction = NewDoubleTransactionsDetector("DetectorProcessBlockTransaction")

// DetectorComputeSortedTxs -
var DetectorComputeSortedTxs = NewDoubleTransactionsDetector("DetectorComputeSortedTxs")

// DetectorComputeSortedPlusRemainingTxs -
var DetectorComputeSortedPlusRemainingTxs = NewDoubleTransactionsDetector("DetectorComputeSortedPlusRemainingTxs")

// DetectorComputeTxsFromMe -
var DetectorComputeTxsFromMe = NewDoubleTransactionsDetector("DetectorComputeTxsFromMe")

// ClearAll -
func ClearAll() {
	DetectorForProcessBadTransaction.Clear()
	DetectorProcessBlockTransaction.Clear()
	DetectorComputeSortedTxs.Clear()
	DetectorComputeSortedPlusRemainingTxs.Clear()
	DetectorComputeTxsFromMe.Clear()
}

// PrintAll -
func PrintAll() {
	DetectorForProcessBadTransaction.PrintReport()
	DetectorProcessBlockTransaction.PrintReport()
	DetectorComputeSortedTxs.PrintReport()
	DetectorComputeSortedPlusRemainingTxs.PrintReport()
	DetectorComputeTxsFromMe.PrintReport()
}

type miniblockInfo struct {
	txs map[string]int
}

type doubleTransactionsDetector struct {
	mut  sync.Mutex
	db   map[string]*miniblockInfo
	name string
}

// NewDoubleTransactionsDetector -
func NewDoubleTransactionsDetector(name string) *doubleTransactionsDetector {
	return &doubleTransactionsDetector{
		db:   make(map[string]*miniblockInfo),
		name: name,
	}
}

// AddTxHash -
func (detector *doubleTransactionsDetector) AddTxHash(txHash []byte, miniblockHash []byte) bool {
	detector.mut.Lock()
	defer detector.mut.Unlock()

	mb := detector.db[string(miniblockHash)]
	if mb == nil {
		mb = &miniblockInfo{
			txs: make(map[string]int),
		}
		detector.db[string(miniblockHash)] = mb
	}

	if mb.txs[string(txHash)] > 0 {
		log.Warn("doubleTransactionsDetector - "+detector.name,
			"tx hash", txHash, "current count", mb.txs[string(txHash)], "mb hash", miniblockHash,
			"stack", string(debug.Stack()))
	}
	mb.txs[string(txHash)]++

	return mb.txs[string(txHash)] > 1
}

// Clear -
func (detector *doubleTransactionsDetector) Clear() {
	detector.mut.Lock()
	defer detector.mut.Unlock()

	detector.db = make(map[string]*miniblockInfo)
}

// PrintReport -
func (detector *doubleTransactionsDetector) PrintReport() {
	detector.mut.Lock()
	defer detector.mut.Unlock()

	found := false
	for mbHash, mb := range detector.db {
		for txHash, counter := range mb.txs {
			if counter == 1 {
				continue
			}

			log.Error("doubleTransactionsDetector - "+detector.name,
				"tx hash", []byte(txHash), "count", counter, "mb hash", []byte(mbHash))
			found = true
		}
	}

	if !found {
		log.Debug("doubleTransactionsDetector - " + detector.name + " - all good")
	}
}
