package debug

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("debug")

// GenericMiniBlockHash -
var GenericMiniBlockHash = make([]byte, 0)

// DetectorForIncoming -
var DetectorForIncoming = NewDoubleTransactionsDetector("DetectorForIncoming")

// DetectorForProcessing -
var DetectorForProcessing = NewDoubleTransactionsDetector("DetectorForProcessing")

// DetectorForGetSortedTransactions -
var DetectorForGetSortedTransactions = NewDoubleTransactionsDetector("DetectorForGetSortedTransactions")

// DetectorForSortTransactionsBySenderAndNonce -
var DetectorForSortTransactionsBySenderAndNonce = NewDoubleTransactionsDetector("DetectorForSortTransactionsBySenderAndNonce")

// DetectorForProcessBadTransaction -
var DetectorForProcessBadTransaction = NewDoubleTransactionsDetector("DetectorForProcessBadTransaction")

// ClearAll -
func ClearAll() {
	DetectorForIncoming.Clear()
	DetectorForProcessing.Clear()
	DetectorForGetSortedTransactions.Clear()
	DetectorForSortTransactionsBySenderAndNonce.Clear()
	DetectorForProcessBadTransaction.Clear()
}

// PrintAll -
func PrintAll() {
	DetectorForIncoming.PrintReport()
	DetectorForProcessing.PrintReport()
	DetectorForGetSortedTransactions.PrintReport()
	DetectorForSortTransactionsBySenderAndNonce.PrintReport()
	DetectorForProcessBadTransaction.PrintReport()
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
func (detector *doubleTransactionsDetector) AddTxHash(txHash []byte, miniblockHash []byte) {
	detector.mut.Lock()
	defer detector.mut.Unlock()

	mb := detector.db[string(miniblockHash)]
	if mb == nil {
		mb = &miniblockInfo{
			txs: make(map[string]int),
		}
		detector.db[string(miniblockHash)] = mb
	}

	mb.txs[string(txHash)]++
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
