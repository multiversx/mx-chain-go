package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ dataRetriever.PoolsHolder = (*dataPool)(nil)

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
	currBlockValidatorInfo dataRetriever.ValidatorInfoCacher
	smartContracts         storage.Cacher
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
	CurrentBlockValidatorInfo dataRetriever.ValidatorInfoCacher
	SmartContracts            storage.Cacher
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
	if check.IfNil(args.CurrentBlockValidatorInfo) {
		return nil, dataRetriever.ErrNilCurrBlockValidatorInfo
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
		currBlockValidatorInfo: args.CurrentBlockValidatorInfo,
		smartContracts:         args.SmartContracts,
		validatorsInfo:         args.ValidatorsInfo,
	}, nil
}

// CurrentBlockTxs returns the holder for current block transactions
func (dp *dataPool) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return dp.currBlockTxs
}

// CurrentBlockValidatorInfo returns the holder for current block validator info
func (dp *dataPool) CurrentBlockValidatorInfo() dataRetriever.ValidatorInfoCacher {
	return dp.currBlockValidatorInfo
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

// ValidatorsInfo returns the holder for validators info
func (dp *dataPool) ValidatorsInfo() dataRetriever.ShardedDataCacherNotifier {
	return dp.validatorsInfo
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *dataPool) IsInterfaceNil() bool {
	return dp == nil
}
