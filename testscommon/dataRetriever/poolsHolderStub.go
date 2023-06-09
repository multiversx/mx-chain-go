package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// PoolsHolderStub -
type PoolsHolderStub struct {
	HeadersCalled                        func() dataRetriever.HeadersPool
	TransactionsCalled                   func() dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactionsCalled           func() dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsCalled             func() dataRetriever.ShardedDataCacherNotifier
	MiniBlocksCalled                     func() storage.Cacher
	MetaBlocksCalled                     func() storage.Cacher
	CurrBlockTxsCalled                   func() dataRetriever.TransactionCacher
	CurrEpochValidatorInfoCalled         func() dataRetriever.ValidatorInfoCacher
	TrieNodesCalled                      func() storage.Cacher
	TrieNodesChunksCalled                func() storage.Cacher
	PeerChangesBlocksCalled              func() storage.Cacher
	SmartContractsCalled                 func() storage.Cacher
	PeerAuthenticationsCalled            func() storage.Cacher
	HeartbeatsCalled                     func() storage.Cacher
	FullArchivePeerAuthenticationsCalled func() storage.Cacher
	FullArchiveHeartbeatsCalled          func() storage.Cacher
	ValidatorsInfoCalled                 func() dataRetriever.ShardedDataCacherNotifier
	CloseCalled                          func() error
}

// NewPoolsHolderStub -
func NewPoolsHolderStub() *PoolsHolderStub {
	return &PoolsHolderStub{}
}

// Headers -
func (holder *PoolsHolderStub) Headers() dataRetriever.HeadersPool {
	if holder.HeadersCalled != nil {
		return holder.HeadersCalled()
	}

	return nil
}

// Transactions -
func (holder *PoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	if holder.TransactionsCalled != nil {
		return holder.TransactionsCalled()
	}

	return testscommon.NewShardedDataStub()
}

// UnsignedTransactions -
func (holder *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	if holder.UnsignedTransactionsCalled != nil {
		return holder.UnsignedTransactionsCalled()
	}

	return testscommon.NewShardedDataStub()
}

// RewardTransactions -
func (holder *PoolsHolderStub) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	if holder.RewardTransactionsCalled != nil {
		return holder.RewardTransactionsCalled()
	}

	return testscommon.NewShardedDataStub()
}

// MiniBlocks -
func (holder *PoolsHolderStub) MiniBlocks() storage.Cacher {
	if holder.MiniBlocksCalled != nil {
		return holder.MiniBlocksCalled()
	}

	return testscommon.NewCacherStub()
}

// MetaBlocks -
func (holder *PoolsHolderStub) MetaBlocks() storage.Cacher {
	if holder.MetaBlocksCalled != nil {
		return holder.MetaBlocksCalled()
	}

	return testscommon.NewCacherStub()
}

// CurrentBlockTxs -
func (holder *PoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	if holder.CurrBlockTxsCalled != nil {
		return holder.CurrBlockTxsCalled()
	}

	return nil
}

// CurrentEpochValidatorInfo -
func (holder *PoolsHolderStub) CurrentEpochValidatorInfo() dataRetriever.ValidatorInfoCacher {
	if holder.CurrEpochValidatorInfoCalled != nil {
		return holder.CurrEpochValidatorInfoCalled()
	}

	return nil
}

// TrieNodes -
func (holder *PoolsHolderStub) TrieNodes() storage.Cacher {
	if holder.TrieNodesCalled != nil {
		return holder.TrieNodesCalled()
	}

	return testscommon.NewCacherStub()
}

// TrieNodesChunks -
func (holder *PoolsHolderStub) TrieNodesChunks() storage.Cacher {
	if holder.TrieNodesChunksCalled != nil {
		return holder.TrieNodesChunksCalled()
	}

	return testscommon.NewCacherStub()
}

// PeerChangesBlocks -
func (holder *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	if holder.PeerChangesBlocksCalled != nil {
		return holder.PeerChangesBlocksCalled()
	}

	return testscommon.NewCacherStub()
}

// SmartContracts -
func (holder *PoolsHolderStub) SmartContracts() storage.Cacher {
	if holder.SmartContractsCalled != nil {
		return holder.SmartContractsCalled()
	}

	return testscommon.NewCacherStub()
}

// PeerAuthentications -
func (holder *PoolsHolderStub) PeerAuthentications() storage.Cacher {
	if holder.PeerAuthenticationsCalled != nil {
		return holder.PeerAuthenticationsCalled()
	}

	return testscommon.NewCacherStub()
}

// Heartbeats -
func (holder *PoolsHolderStub) Heartbeats() storage.Cacher {
	if holder.HeartbeatsCalled != nil {
		return holder.HeartbeatsCalled()
	}

	return testscommon.NewCacherStub()
}

// FullArchivePeerAuthentications -
func (holder *PoolsHolderStub) FullArchivePeerAuthentications() storage.Cacher {
	if holder.FullArchivePeerAuthenticationsCalled != nil {
		return holder.FullArchivePeerAuthenticationsCalled()
	}

	return testscommon.NewCacherStub()
}

// FullArchiveHeartbeats -
func (holder *PoolsHolderStub) FullArchiveHeartbeats() storage.Cacher {
	if holder.FullArchiveHeartbeatsCalled != nil {
		return holder.FullArchiveHeartbeatsCalled()
	}

	return testscommon.NewCacherStub()
}

// ValidatorsInfo -
func (holder *PoolsHolderStub) ValidatorsInfo() dataRetriever.ShardedDataCacherNotifier {
	if holder.ValidatorsInfoCalled != nil {
		return holder.ValidatorsInfoCalled()
	}

	return testscommon.NewShardedDataStub()
}

// Close -
func (holder *PoolsHolderStub) Close() error {
	if holder.CloseCalled != nil {
		return holder.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *PoolsHolderStub) IsInterfaceNil() bool {
	return holder == nil
}
