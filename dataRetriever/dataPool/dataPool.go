package dataPool

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ dataRetriever.PoolsHolder = (*dataPool)(nil)

var log = logger.GetOrCreate("dataRetriever/dataPool")

type dataPool struct {
	transactions           dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions   dataRetriever.ShardedDataCacherNotifier
	rewardTransactions     dataRetriever.ShardedDataCacherNotifier
	headers                dataRetriever.HeadersPool
	miniBlocks             storage.Cacher
	peerChangesBlocks      storage.Cacher
	trieNodes              storage.Cacher
	trieNodesChunks        storage.Cacher
	currBlockTxs           dataRetriever.TransactionCacher
	currEpochValidatorInfo dataRetriever.ValidatorInfoCacher
	smartContracts         storage.Cacher
	peerAuthentications    storage.Cacher
	heartbeats             storage.Cacher
	validatorsInfo         dataRetriever.ShardedDataCacherNotifier
}

// DataPoolArgs represents the data pool's constructor structure
type DataPoolArgs struct {
	Transactions              dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactions      dataRetriever.ShardedDataCacherNotifier
	RewardTransactions        dataRetriever.ShardedDataCacherNotifier
	Headers                   dataRetriever.HeadersPool
	MiniBlocks                storage.Cacher
	PeerChangesBlocks         storage.Cacher
	TrieNodes                 storage.Cacher
	TrieNodesChunks           storage.Cacher
	CurrentBlockTransactions  dataRetriever.TransactionCacher
	CurrentEpochValidatorInfo dataRetriever.ValidatorInfoCacher
	SmartContracts            storage.Cacher
	PeerAuthentications       storage.Cacher
	Heartbeats                storage.Cacher
	ValidatorsInfo            dataRetriever.ShardedDataCacherNotifier
}

// NewDataPool creates a data pools holder object
func NewDataPool(args DataPoolArgs) (*dataPool, error) {
	if check.IfNil(args.Transactions) {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if check.IfNil(args.UnsignedTransactions) {
		return nil, dataRetriever.ErrNilUnsignedTransactionPool
	}
	if check.IfNil(args.RewardTransactions) {
		return nil, dataRetriever.ErrNilRewardTransactionPool
	}
	if check.IfNil(args.Headers) {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if check.IfNil(args.MiniBlocks) {
		return nil, dataRetriever.ErrNilTxBlockDataPool
	}
	if check.IfNil(args.PeerChangesBlocks) {
		return nil, dataRetriever.ErrNilPeerChangeBlockDataPool
	}
	if check.IfNil(args.CurrentBlockTransactions) {
		return nil, dataRetriever.ErrNilCurrBlockTxs
	}
	if check.IfNil(args.CurrentEpochValidatorInfo) {
		return nil, dataRetriever.ErrNilCurrentEpochValidatorInfo
	}
	if check.IfNil(args.TrieNodes) {
		return nil, dataRetriever.ErrNilTrieNodesPool
	}
	if check.IfNil(args.TrieNodesChunks) {
		return nil, dataRetriever.ErrNilTrieNodesChunksPool
	}
	if check.IfNil(args.SmartContracts) {
		return nil, dataRetriever.ErrNilSmartContractsPool
	}
	if check.IfNil(args.PeerAuthentications) {
		return nil, dataRetriever.ErrNilPeerAuthenticationPool
	}
	if check.IfNil(args.Heartbeats) {
		return nil, dataRetriever.ErrNilHeartbeatPool
	}
	if check.IfNil(args.ValidatorsInfo) {
		return nil, dataRetriever.ErrNilValidatorInfoPool
	}

	return &dataPool{
		transactions:           args.Transactions,
		unsignedTransactions:   args.UnsignedTransactions,
		rewardTransactions:     args.RewardTransactions,
		headers:                args.Headers,
		miniBlocks:             args.MiniBlocks,
		peerChangesBlocks:      args.PeerChangesBlocks,
		trieNodes:              args.TrieNodes,
		trieNodesChunks:        args.TrieNodesChunks,
		currBlockTxs:           args.CurrentBlockTransactions,
		currEpochValidatorInfo: args.CurrentEpochValidatorInfo,
		smartContracts:         args.SmartContracts,
		peerAuthentications:    args.PeerAuthentications,
		heartbeats:             args.Heartbeats,
		validatorsInfo:         args.ValidatorsInfo,
	}, nil
}

// CurrentBlockTxs returns the holder for current block transactions
func (dp *dataPool) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return dp.currBlockTxs
}

// CurrentEpochValidatorInfo returns the holder for current epoch validator info
func (dp *dataPool) CurrentEpochValidatorInfo() dataRetriever.ValidatorInfoCacher {
	return dp.currEpochValidatorInfo
}

// Transactions returns the holder for transactions
func (dp *dataPool) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return dp.transactions
}

// UnsignedTransactions returns the holder for unsigned transactions (cross shard result entities)
func (dp *dataPool) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return dp.unsignedTransactions
}

// RewardTransactions returns the holder for reward transactions (cross shard result entities)
func (dp *dataPool) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return dp.rewardTransactions
}

// Headers returns the holder for headers
func (dp *dataPool) Headers() dataRetriever.HeadersPool {
	return dp.headers
}

// MiniBlocks returns the holder for miniblocks
func (dp *dataPool) MiniBlocks() storage.Cacher {
	return dp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (dp *dataPool) PeerChangesBlocks() storage.Cacher {
	return dp.peerChangesBlocks
}

// TrieNodes returns the holder for trie nodes
func (dp *dataPool) TrieNodes() storage.Cacher {
	return dp.trieNodes
}

// TrieNodesChunks returns the holder for trie nodes chunks
func (dp *dataPool) TrieNodesChunks() storage.Cacher {
	return dp.trieNodesChunks
}

// SmartContracts returns the holder for smart contracts
func (dp *dataPool) SmartContracts() storage.Cacher {
	return dp.smartContracts
}

// PeerAuthentications returns the holder for peer authentications
func (dp *dataPool) PeerAuthentications() storage.Cacher {
	return dp.peerAuthentications
}

// Heartbeats returns the holder for heartbeats
func (dp *dataPool) Heartbeats() storage.Cacher {
	return dp.heartbeats
}

// ValidatorsInfo returns the holder for validators info
func (dp *dataPool) ValidatorsInfo() dataRetriever.ShardedDataCacherNotifier {
	return dp.validatorsInfo
}

// Close closes all the components
func (dp *dataPool) Close() error {
	var lastError error
	if !check.IfNil(dp.trieNodes) {
		log.Debug("closing trie nodes data pool....")
		err := dp.trieNodes.Close()
		if err != nil {
			log.Error("failed to close trie nodes data pool", "error", err.Error())
			lastError = err
		}
	}

	if !check.IfNil(dp.peerAuthentications) {
		log.Debug("closing peer authentications data pool....")
		err := dp.peerAuthentications.Close()
		if err != nil {
			log.Error("failed to close peer authentications data pool", "error", err.Error())
			lastError = err
		}
	}

	return lastError
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *dataPool) IsInterfaceNil() bool {
	return dp == nil
}
