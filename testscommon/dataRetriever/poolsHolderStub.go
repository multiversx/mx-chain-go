package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
)

// PoolsHolderStub -
type PoolsHolderStub struct {
	HeadersCalled                func() dataRetriever.HeadersPool
	TransactionsCalled           func() dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactionsCalled   func() dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsCalled     func() dataRetriever.ShardedDataCacherNotifier
	MiniBlocksCalled             func() storage.Cacher
	MetaBlocksCalled             func() storage.Cacher
	CurrBlockTxsCalled           func() dataRetriever.TransactionCacher
	CurrEpochValidatorInfoCalled func() dataRetriever.ValidatorInfoCacher
	TrieNodesCalled              func() storage.Cacher
	TrieNodesChunksCalled        func() storage.Cacher
	PeerChangesBlocksCalled      func() storage.Cacher
	SmartContractsCalled         func() storage.Cacher
	PeerAuthenticationsCalled    func() storage.Cacher
	HeartbeatsCalled             func() storage.Cacher
	ValidatorsInfoCalled         func() dataRetriever.ShardedDataCacherNotifier
	ProofsCalled                 func() dataRetriever.ProofsPool
	CloseCalled                  func() error
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

	return cache.NewCacherStub()
}

// MetaBlocks -
func (holder *PoolsHolderStub) MetaBlocks() storage.Cacher {
	if holder.MetaBlocksCalled != nil {
		return holder.MetaBlocksCalled()
	}

	return cache.NewCacherStub()
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

	return cache.NewCacherStub()
}

// TrieNodesChunks -
func (holder *PoolsHolderStub) TrieNodesChunks() storage.Cacher {
	if holder.TrieNodesChunksCalled != nil {
		return holder.TrieNodesChunksCalled()
	}

	return cache.NewCacherStub()
}

// PeerChangesBlocks -
func (holder *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	if holder.PeerChangesBlocksCalled != nil {
		return holder.PeerChangesBlocksCalled()
	}

	return cache.NewCacherStub()
}

// SmartContracts -
func (holder *PoolsHolderStub) SmartContracts() storage.Cacher {
	if holder.SmartContractsCalled != nil {
		return holder.SmartContractsCalled()
	}

	return cache.NewCacherStub()
}

// PeerAuthentications -
func (holder *PoolsHolderStub) PeerAuthentications() storage.Cacher {
	if holder.PeerAuthenticationsCalled != nil {
		return holder.PeerAuthenticationsCalled()
	}

	return cache.NewCacherStub()
}

// Heartbeats -
func (holder *PoolsHolderStub) Heartbeats() storage.Cacher {
	if holder.HeartbeatsCalled != nil {
		return holder.HeartbeatsCalled()
	}

	return cache.NewCacherStub()
}

// ValidatorsInfo -
func (holder *PoolsHolderStub) ValidatorsInfo() dataRetriever.ShardedDataCacherNotifier {
	if holder.ValidatorsInfoCalled != nil {
		return holder.ValidatorsInfoCalled()
	}

	return testscommon.NewShardedDataStub()
}

// Proofs -
func (holder *PoolsHolderStub) Proofs() dataRetriever.ProofsPool {
	if holder.ProofsCalled != nil {
		return holder.ProofsCalled()
	}

	return nil
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
