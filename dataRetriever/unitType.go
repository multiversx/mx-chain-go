package dataRetriever

import "fmt"

const (
	// TransactionUnit is the transactions storage unit identifier
	TransactionUnit UnitType = 0
	// MiniBlockUnit is the transaction block body storage unit identifier
	MiniBlockUnit UnitType = 1
	// PeerChangesUnit is the peer change block body storage unit identifier
	PeerChangesUnit UnitType = 2
	// BlockHeaderUnit is the Block Headers Storage unit identifier
	BlockHeaderUnit UnitType = 3
	// MetaBlockUnit is the metachain blocks storage unit identifier
	MetaBlockUnit UnitType = 4
	// UnsignedTransactionUnit is the unsigned transaction unit identifier
	UnsignedTransactionUnit UnitType = 5
	// RewardTransactionUnit is the reward transaction unit identifier
	RewardTransactionUnit UnitType = 6
	// MetaHdrNonceHashDataUnit is the meta header nonce-hash pair data unit identifier
	MetaHdrNonceHashDataUnit UnitType = 7
	// HeartbeatUnit is the heartbeat storage unit identifier
	HeartbeatUnit UnitType = 8
	// BootstrapUnit is the bootstrap storage unit identifier
	BootstrapUnit UnitType = 9
	//StatusMetricsUnit is the status metrics storage unit identifier
	StatusMetricsUnit UnitType = 10
	// TxLogsUnit is the transactions logs storage unit identifier
	TxLogsUnit UnitType = 11
	// MiniblocksMetadataUnit is the miniblocks metadata storage unit identifier
	MiniblocksMetadataUnit UnitType = 12
	// EpochByHashUnit is the epoch by hash storage unit identifier
	EpochByHashUnit UnitType = 13
	// MiniblockHashByTxHashUnit is the miniblocks hash by tx hash storage unit identifier
	MiniblockHashByTxHashUnit UnitType = 14
	// ReceiptsUnit is the receipts storage unit identifier
	ReceiptsUnit UnitType = 15
	// ResultsHashesByTxHashUnit is the results hashes by transaction storage unit identifier
	ResultsHashesByTxHashUnit UnitType = 16
	// TrieEpochRootHashUnit is the trie epoch <-> root hash storage unit identifier
	TrieEpochRootHashUnit UnitType = 17
	// ESDTSuppliesUnit is the ESDT supplies storage unit identifier
	ESDTSuppliesUnit UnitType = 18
	// RoundHdrHashDataUnit is the round- block header hash storage data unit identifier
	RoundHdrHashDataUnit UnitType = 19
	// UserAccountsUnit is the user accounts storage unit identifier
	UserAccountsUnit UnitType = 20
	// PeerAccountsUnit is the peer accounts storage unit identifier
	PeerAccountsUnit UnitType = 21
	// ScheduledSCRsUnit is the scheduled SCRs storage unit identifier
	ScheduledSCRsUnit UnitType = 22
	// EpochStartStaticUnit is the epochstart metablocks storage unit identifier
	EpochStartStaticUnit UnitType = 23

	// ShardHdrNonceHashDataUnit is the header nonce-hash pair data unit identifier
	//TODO: Add only unit types lower than 100
	ShardHdrNonceHashDataUnit UnitType = 100
	//TODO: Do not add unit type greater than 100 as the metachain creates this kind of unit type for each shard.
	//100 -> shard 0, 101 -> shard 1 and so on. This should be replaced with a factory which will manage the unit types
	//creation
)

// UnitType is the type for Storage unit identifiers
type UnitType uint8

// String returns the friendly name of the unit
func (ut UnitType) String() string {
	switch ut {
	case TransactionUnit:
		return "TransactionUnit"
	case MiniBlockUnit:
		return "MiniBlockUnit"
	case PeerChangesUnit:
		return "PeerChangesUnit"
	case BlockHeaderUnit:
		return "BlockHeaderUnit"
	case MetaBlockUnit:
		return "MetaBlockUnit"
	case UnsignedTransactionUnit:
		return "UnsignedTransactionUnit"
	case RewardTransactionUnit:
		return "RewardTransactionUnit"
	case MetaHdrNonceHashDataUnit:
		return "MetaHdrNonceHashDataUnit"
	case HeartbeatUnit:
		return "HeartbeatUnit"
	case BootstrapUnit:
		return "BootstrapUnit"
	case StatusMetricsUnit:
		return "StatusMetricsUnit"
	case TxLogsUnit:
		return "TxLogsUnit"
	case MiniblocksMetadataUnit:
		return "MiniblocksMetadataUnit"
	case EpochByHashUnit:
		return "EpochByHashUnit"
	case MiniblockHashByTxHashUnit:
		return "MiniblockHashByTxHashUnit"
	case ReceiptsUnit:
		return "ReceiptsUnit"
	case ResultsHashesByTxHashUnit:
		return "ResultsHashesByTxHashUnit"
	case TrieEpochRootHashUnit:
		return "TrieEpochRootHashUnit"
	case ESDTSuppliesUnit:
		return "ESDTSuppliesUnit"
	case RoundHdrHashDataUnit:
		return "RoundHdrHashDataUnit"
	case UserAccountsUnit:
		return "UserAccountsUnit"
	case PeerAccountsUnit:
		return "PeerAccountsUnit"
	case ScheduledSCRsUnit:
		return "ScheduledSCRsUnit"
	case EpochStartStaticUnit:
		return "EpochStartStaticUnit"
	}

	if ut < ShardHdrNonceHashDataUnit {
		return fmt.Sprintf("unknown type %d", ut)
	}

	return fmt.Sprintf("%s%d", "ShardHdrNonceHashDataUnit", ut-ShardHdrNonceHashDataUnit)
}
