package process

// BlockHeaderState specifies which is the state of the block header received
type BlockHeaderState int

const (
	// BHReceived defines ID of a received block header
	BHReceived BlockHeaderState = iota
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

// ShardBlockFinality defines the block finality which is used in shards (the real finality in shards is given
// by meta-chain)
const ShardBlockFinality = 1

// MetaBlockFinality defines the block finality which is used in meta-chain
const MetaBlockFinality = 1

// MaxHeaderRequestsAllowed defines the maximum number of missing cross-shard headers (gaps) which could be requested
// in one round, when node processes a received block
const MaxHeaderRequestsAllowed = 10

// MaxItemsInBlock defines the maximum threshold which could be set, and represents the maximum number of items
// (hashes of: mini blocks, txs, meta-headers, shard-headers) which could be added in one block
const MaxItemsInBlock = 15000

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

// MaxRoundsWithoutReceivedBlock defines the maximum rounds to wait for a new block to be received, before a special
// action to be applied
const MaxRoundsWithoutReceivedBlock = 5

// MaxRoundsWithoutCommittedBlock defines the maximum rounds to wait for a new block to be committed, before a special
// action to be applied
const MaxRoundsWithoutCommittedBlock = 20
