package workItems

import "github.com/ElrondNetwork/elrond-go/core/statistics"

type itemTpsBenchmark struct {
	indexer      saveTpsBenchmark
	tpsBenchmark statistics.TPSBenchmark
}

// NewItemTpsBenchmark will create a new instance of itemTpsBenchmark
func NewItemTpsBenchmark(indexer saveTpsBenchmark, tpsBenchmark statistics.TPSBenchmark) WorkItemHandler {
	return &itemTpsBenchmark{
		indexer:      indexer,
		tpsBenchmark: tpsBenchmark,
	}
}

// Save will save information about tps benchmark in elasticsearch database
func (wit *itemTpsBenchmark) Save() error {
	err := wit.indexer.SaveShardStatistics(wit.tpsBenchmark)
	if err != nil {
		log.Warn("itemTpsBenchmark.Save", "could not index tps benchmark", err.Error())
		return err
	}

	return nil
}
