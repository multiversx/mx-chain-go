package components

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// CreateStore creates a storage service for shard nodes
func CreateStore(numOfShards uint32) dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.StatusMetricsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.TxLogsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.UserAccountsUnit, CreateMemUnitForTries())
	store.AddStorer(dataRetriever.UserAccountsCheckpointsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.PeerAccountsUnit, CreateMemUnitForTries())
	store.AddStorer(dataRetriever.PeerAccountsCheckpointsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ESDTSuppliesUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.RoundHdrHashDataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniblocksMetadataUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniblockHashByTxHashUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.EpochByHashUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ResultsHashesByTxHashUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, CreateMemUnit())

	for i := uint32(0); i < numOfShards; i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, CreateMemUnit())
	}

	return store
}
