package process

import (
	"fmt"
)

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
	// BuiltInFunctionCall defines ID of a builtin function call
	BuiltInFunctionCall
	// RelayedTx defines ID of a transaction of type relayed
	RelayedTx
	// RelayedTxV2 defines the ID of a slim relayed transaction version
	RelayedTxV2
	// RewardTx defines ID of a reward transaction
	RewardTx
	// InvalidTransaction defines unknown transaction type
	InvalidTransaction
)

func (transactionType TransactionType) String() string {
	switch transactionType {
	case MoveBalance:
		return "MoveBalance"
	case SCDeployment:
		return "SCDeployment"
	case SCInvoking:
		return "SCInvoking"
	case BuiltInFunctionCall:
		return "BuiltInFunctionCall"
	case RelayedTx:
		return "RelayedTx"
	case RelayedTxV2:
		return "RelayedTxV2"
	case RewardTx:
		return "RewardTx"
	case InvalidTransaction:
		return "InvalidTransaction"
	default:
		return fmt.Sprintf("type %d", transactionType)
	}
}

// BlockFinality defines the block finality which is used in meta-chain/shards (the real finality in shards is given
// by meta-chain)
const BlockFinality = 1

// MetaBlockValidity defines the block validity which is when checking a metablock
const MetaBlockValidity = 1

// EpochChangeGracePeriod defines the allowed round numbers till the shard has to change the epoch
const EpochChangeGracePeriod = 1

// MaxHeaderRequestsAllowed defines the maximum number of missing cross-shard headers (gaps) which could be requested
// in one round, when node processes a received block
const MaxHeaderRequestsAllowed = 20

// NumTxPerSenderBatchForFillingMiniblock defines the number of transactions to be drawn
// from the transactions pool, for a specific sender, in a single pass.
// Drawing transactions for a miniblock happens in multiple passes, until "MaxItemsInBlock" are drawn.
const NumTxPerSenderBatchForFillingMiniblock = 10

// NonceDifferenceWhenSynced defines the difference between probable highest nonce seen from network and node's last
// committed block nonce, after which, node is considered himself not synced
const NonceDifferenceWhenSynced = 0

// MaxSyncWithErrorsAllowed defines the maximum allowed number of sync with errors,
// before a special action to be applied
const MaxSyncWithErrorsAllowed = 10

// MaxHeadersToRequestInAdvance defines the maximum number of headers which will be requested in advance,
// if they are missing
const MaxHeadersToRequestInAdvance = 20

// RoundModulusTrigger defines a round modulus on which a trigger for an action will be released
const RoundModulusTrigger = 5

// RoundModulusTriggerWhenSyncIsStuck defines a round modulus on which a trigger for an action when sync is stuck will be released
const RoundModulusTriggerWhenSyncIsStuck = 20

// MaxRoundsWithoutCommittedBlock defines the maximum rounds to wait for a new block to be committed,
// before a special action to be applied
const MaxRoundsWithoutCommittedBlock = 10

// MinForkRound represents the minimum fork round set by a notarized header received
const MinForkRound = uint64(0)

// MaxMetaNoncesBehind defines the maximum difference between the current meta block nonce and the processed meta block
// nonce before a shard is considered stuck
const MaxMetaNoncesBehind = 15

// MaxMetaNoncesBehindForGlobalStuck defines the maximum difference between the current meta block nonce and the processed
// meta block nonce for any shard, where the chain is considered stuck and enters recovery
const MaxMetaNoncesBehindForGlobalStuck = 30

// MaxShardNoncesBehind defines the maximum difference between the current shard block nonce and the last notarized
// shard block nonce by meta, before meta is considered stuck
const MaxShardNoncesBehind = 15

// MaxRoundsWithoutNewBlockReceived defines the maximum number of rounds to wait for a new block to be received,
// before a special action to be applied
const MaxRoundsWithoutNewBlockReceived = 10

// MaxMetaHeadersAllowedInOneShardBlock defines the maximum number of meta headers allowed to be included in one shard block
const MaxMetaHeadersAllowedInOneShardBlock = 50

// MaxShardHeadersAllowedInOneMetaBlock defines the maximum number of shard headers allowed to be included in one meta block
const MaxShardHeadersAllowedInOneMetaBlock = 60

// MinShardHeadersFromSameShardInOneMetaBlock defines the minimum number of shard headers from the same shard,
// which would be included in one meta block if they are available
const MinShardHeadersFromSameShardInOneMetaBlock = 10

// MaxNumOfTxsToSelect defines the maximum number of transactions that should be selected from the cache
const MaxNumOfTxsToSelect = 30000

// MaxGasBandwidthPerBatchPerSender defines the maximum gas bandwidth that should be selected for a sender per batch from the cache
const MaxGasBandwidthPerBatchPerSender = 5000000

// MaxHeadersToWhitelistInAdvance defines the maximum number of headers whose miniblocks will be whitelisted in advance
const MaxHeadersToWhitelistInAdvance = 300

// MaxGasFeeHigherFactorAccepted defines the maximum higher factor of gas fee put inside a transaction compared with
// the real gas used, after which the transaction will be considered an attack and all the gas will be consumed and
// nothing will be refunded to the sender
const MaxGasFeeHigherFactorAccepted = 10
