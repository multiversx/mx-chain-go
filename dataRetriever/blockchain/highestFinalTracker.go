package blockchain

type currentCoordinates struct {
	highestFinalNonce    uint64
	currentBlockNonce    uint64
	currentBlockHash     []byte
	currentBlockRootHash []byte
}

type coordinatesRecord struct {
	blockNonce    uint64
	blockHash     []byte
	blockRootHash []byte
}

type highestFinalTracker struct {
	records            []coordinatesRecord
	highestFinalRecord coordinatesRecord
}

func newHighestFinalTracker() *highestFinalTracker {
	return &highestFinalTracker{
		records: make([]coordinatesRecord, 0),
	}
}

func (tracker *highestFinalTracker) trackCoordinates(coordinates currentCoordinates) {
	// Record the new coordinates
	tracker.records = append(tracker.records, coordinatesRecord{
		blockNonce:    coordinates.currentBlockNonce,
		blockHash:     coordinates.currentBlockHash,
		blockRootHash: coordinates.currentBlockRootHash,
	})

	// Search for a record that refers to the highest final block
	// If found, also forget previous records
	forgettingIndex := 0

	for i, record := range tracker.records {
		if record.blockNonce == coordinates.highestFinalNonce {
			tracker.highestFinalRecord = record
			forgettingIndex = i

			log.Trace("highestFinalTracker: found highest final in records",
				"nonce", record.blockNonce,
				"hash", record.blockHash,
				"rootHash", record.blockRootHash)
		}
	}

	tracker.records = tracker.records[forgettingIndex:]
	log.Trace("highestFinalTracker", "length", len(tracker.records))
}
