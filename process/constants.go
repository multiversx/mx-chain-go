package process

// BlockHeaderState specifies which is the state of the block header received
type BlockHeaderState int

const (
	// BHReceived defines ID of a received block header
	BHReceived BlockHeaderState = iota
	// BHReceivedTooLate defines ID of a late received block header
	BHReceivedTooLate
	// BHProcessed defines ID of a processed block header
	BHProcessed
	// BHProposed defines ID of a proposed block header
	BHProposed
	// BHNotarized defines ID of a notarized block header
	BHNotarized
)

// TransactionType specifies the type of the transaction
type TransactionType int

const (
	// MoveBalance defines ID of a payment transaction - moving balances
	MoveBalance TransactionType = iota
	// SCDeployment defines ID of a transaction to store a smart contract
	SCDeployment
	// SCInvoking defines ID of a transaction of type smart contract call
	SCInvoking
	// RewardTx defines ID of a reward transaction
	RewardTx
	// InvalidTransaction defines unknown transaction type
	InvalidTransaction
)

// BlockFinality defines the block finality which is used in meta-chain/shards (the real finality in shards is given
// by meta-chain)
const BlockFinality = 1

// ShardBlockFinality defines the block finality which is used in shards (the real finality in shards is given
// by meta-chain)
const ShardBlockFinality = 1

// MetaBlockFinality defines the block finality which is used in meta-chain
const MetaBlockFinality = 1

// MetaBlockValidity defines the block validity which is when checking a metablock
const MetaBlockValidity = 1

// EpochChangeGracePeriod defines the allowed round numbers till the shard has to change the epoch
const EpochChangeGracePeriod = 1

// MaxHeaderRequestsAllowed defines the maximum number of missing cross-shard headers (gaps) which could be requested
// in one round, when node processes a received block
const MaxHeaderRequestsAllowed = 10

// MaxItemsInBlock defines the maximum threshold which could be set, and represents the maximum number of items
// (hashes of: mini blocks, txs, meta-headers, shard-headers) which could be added in one block
const MaxItemsInBlock = 15000

// NumTxPerSenderBatchForFillingMiniblock defines the number of transactions to be drawn
// from the transactions pool, for a specific sender, in a single pass.
// Drawing transactions for a miniblock happens in multiple passes, until "MaxItemsInBlock" are drawn.
const NumTxPerSenderBatchForFillingMiniblock = 10

// MinItemsInBlock defines the minimum threshold which could be set, and represents the maximum number of items
// (hashes of: mini blocks, txs, meta-headers, shard-headers) which could be added in one block
const MinItemsInBlock = 15000

// NonceDifferenceWhenSynced defines the difference between probable highest nonce seen from network and node's last
// committed block nonce, after which, node is considered himself not synced
const NonceDifferenceWhenSynced = 0

// MaxRequestsWithTimeoutAllowed defines the maximum allowed number of requests with timeout,
// before a special action to be applied
const MaxRequestsWithTimeoutAllowed = 5

// MaxHeadersToRequestInAdvance defines the maximum number of headers which will be requested in advance,
// if they are missing
const MaxHeadersToRequestInAdvance = 10

// RoundModulusTrigger defines a round modulus on which a trigger for an action will be released
const RoundModulusTrigger = 10

// MaxOccupancyPercentageAllowed defines the maximum occupancy percentage allowed to be used,
// from the full pool capacity, for the received data which are not needed in the near future
const MaxOccupancyPercentageAllowed = float64(0.9)

// MaxRoundsWithoutCommittedBlock defines the maximum rounds to wait for a new block to be committed,
// before a special action to be applied
const MaxRoundsWithoutCommittedBlock = 20

// MaxNoncesWithoutCrossNotarized defines the maximum nonces to wait for a new block to be cross notarized,
// before a special action to be applied
const MaxNoncesWithoutCrossNotarized = 100

// MinForkRound represents the minimum fork round set by a notarized header received
const MinForkRound = uint64(0)

// TxPoolNumSendersToEvictInOneStep instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToEvictInOneStep = uint32(50)

// TxPoolALotOfTransactionsForASender instructs tx pool eviction algorithm to tag a sender with more transactions than this value
// as a "sender with a lot of transactions"
const TxPoolALotOfTransactionsForASender = uint32(500)

// TxPoolNumTxsToEvictForASenderWithALot instructs tx pool eviction algorithm to remove this many transactions
// for "a sender with a lot of transactions" when eviction takes place
const TxPoolNumTxsToEvictForASenderWithALot = uint32(100)

// TxPoolMinSizeInBytes is the lower limit of the tx cache / eviction parameter "sizeInBytes"
const TxPoolMinSizeInBytes = uint32(40960)

// MaxNonceDifferences represents the maximum nonce difference between received and committed header, so the received one
// to be stored in advance in block tracker
const MaxNonceDifferences = uint64(1000)
