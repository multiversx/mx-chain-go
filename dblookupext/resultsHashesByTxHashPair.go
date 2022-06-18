package dblookupext

// ResultsHashesByTxHashPair holds receipt and SCRs hashes for a given transaction
type ResultsHashesByTxHashPair struct {
	TxHash                  []byte
	ReceiptsHash            []byte
	ScResultsHashesAndEpoch []*ScResultsHashesAndEpoch
}
