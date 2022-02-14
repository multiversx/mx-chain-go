package common

import (
	"math"
	"time"
)

// PeerType represents the type of a peer
type PeerType string

// EligibleList represents the list of peers who participate in consensus inside a shard
const EligibleList PeerType = "eligible"

// WaitingList represents the list of peers who don't participate in consensus but will join the next epoch
const WaitingList PeerType = "waiting"

// LeavingList represents the list of peers who were taken out of eligible and waiting because of rating
const LeavingList PeerType = "leaving"

// InactiveList represents the list of peers who were taken out because they were leaving
const InactiveList PeerType = "inactive"

// JailedList represents the list of peers who have stake but are in jail
const JailedList PeerType = "jailed"

// ObserverList represents the list of peers who don't participate in consensus but will join the next epoch
const ObserverList PeerType = "observer"

// NewList -
const NewList PeerType = "new"

// CombinedPeerType - represents the combination of two peerTypes
const CombinedPeerType = "%s (%s)"

// UnVersionedAppString represents the default app version that indicate that the binary wasn't build by setting
// the appVersion flag
const UnVersionedAppString = "undefined"

// DisabledShardIDAsObserver defines the uint32 identifier which tells that the node hasn't configured any preferred
// shard to start in as observer
const DisabledShardIDAsObserver = uint32(0xFFFFFFFF) - 7

// MaxTxNonceDeltaAllowed specifies the maximum difference between an account's nonce and a received transaction's nonce
// in order to mark the transaction as valid.
const MaxTxNonceDeltaAllowed = 30000

// MaxBulkTransactionSize specifies the maximum size of one bulk with txs which can be send over the network
// TODO convert this const into a var and read it from config when this code moves to another binary
const MaxBulkTransactionSize = 1 << 18 // 256KB bulks

// MaxTxsToRequest specifies the maximum number of txs to request
const MaxTxsToRequest = 1000

// NodesSetupJsonFileName specifies the name of the json file which contains the setup of the nodes
const NodesSetupJsonFileName = "nodesSetup.json"

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic = "consensus"

// GenesisTxSignatureString is the string used to generate genesis transaction signature as 128 hex characters
const GenesisTxSignatureString = "GENESISGENESISGENESISGENESISGENESISGENESISGENESISGENESISGENESISG"

// HeartbeatTopic is the topic used for heartbeat signaling
const HeartbeatTopic = "heartbeat"

// PathShardPlaceholder represents the placeholder for the shard ID in paths
const PathShardPlaceholder = "[S]"

// PathEpochPlaceholder represents the placeholder for the epoch number in paths
const PathEpochPlaceholder = "[E]"

// PathIdentifierPlaceholder represents the placeholder for the identifier in paths
const PathIdentifierPlaceholder = "[I]"

// MetricCurrentRound is the metric for monitoring the current round of a node
const MetricCurrentRound = "erd_current_round"

// MetricNonce is the metric for monitoring the nonce of a node
const MetricNonce = "erd_nonce"

// MetricProbableHighestNonce is the metric for monitoring the max speculative nonce received by the node by listening on the network
const MetricProbableHighestNonce = "erd_probable_highest_nonce"

// MetricNumConnectedPeers is the metric for monitoring the number of connected peers
const MetricNumConnectedPeers = "erd_num_connected_peers"

// MetricNumConnectedPeersClassification is the metric for monitoring the number of connected peers split on the connection type
const MetricNumConnectedPeersClassification = "erd_num_connected_peers_classification"

// MetricSynchronizedRound is the metric for monitoring the synchronized round of a node
const MetricSynchronizedRound = "erd_synchronized_round"

// MetricIsSyncing is the metric for monitoring if a node is syncing
const MetricIsSyncing = "erd_is_syncing"

// MetricPublicKeyBlockSign is the metric for monitoring public key of a node used in block signing
const MetricPublicKeyBlockSign = "erd_public_key_block_sign"

// MetricShardId is the metric for monitoring shard id of a node
const MetricShardId = "erd_shard_id"

// MetricNumShardsWithoutMetachain is the metric for monitoring the number of shards (excluding meta)
const MetricNumShardsWithoutMetachain = "erd_num_shards_without_meta"

// MetricTxPoolLoad is the metric for monitoring number of transactions from pool of a node
const MetricTxPoolLoad = "erd_tx_pool_load"

// MetricCountLeader is the metric for monitoring number of rounds when a node was leader
const MetricCountLeader = "erd_count_leader"

// MetricCountConsensus is the metric for monitoring number of rounds when a node was in consensus group
const MetricCountConsensus = "erd_count_consensus"

// MetricCountAcceptedBlocks is the metric for monitoring number of blocks that was accepted proposed by a node
const MetricCountAcceptedBlocks = "erd_count_accepted_blocks"

// MetricNodeType is the metric for monitoring the type of the node
const MetricNodeType = "erd_node_type"

// MetricLiveValidatorNodes is the metric for monitoring live validators on the network
const MetricLiveValidatorNodes = "erd_live_validator_nodes"

// MetricConnectedNodes is the metric for monitoring total connected nodes on the network
const MetricConnectedNodes = "erd_connected_nodes"

// MetricCpuLoadPercent is the metric for monitoring CPU load [%]
const MetricCpuLoadPercent = "erd_cpu_load_percent"

// MetricMemLoadPercent is the metric for monitoring memory load [%]
const MetricMemLoadPercent = "erd_mem_load_percent"

// MetricMemTotal is the metric for monitoring total memory bytes
const MetricMemTotal = "erd_mem_total"

// MetricMemUsedGolang is a metric for monitoring the memory ("total")
const MetricMemUsedGolang = "erd_mem_used_golang"

// MetricMemUsedSystem is a metric for monitoring the memory ("sys mem")
const MetricMemUsedSystem = "erd_mem_used_sys"

// MetricMemHeapInUse is a metric for monitoring the memory ("heap in use")
const MetricMemHeapInUse = "erd_mem_heap_inuse"

// MetricMemStackInUse is a metric for monitoring the memory ("stack in use")
const MetricMemStackInUse = "erd_mem_stack_inuse"

// MetricNetworkRecvPercent is the metric for monitoring network receive load [%]
const MetricNetworkRecvPercent = "erd_network_recv_percent"

// MetricNetworkRecvBps is the metric for monitoring network received bytes per second
const MetricNetworkRecvBps = "erd_network_recv_bps"

// MetricNetworkRecvBpsPeak is the metric for monitoring network received peak bytes per second
const MetricNetworkRecvBpsPeak = "erd_network_recv_bps_peak"

// MetricNetworkRecvBytesInCurrentEpochPerHost is the metric for monitoring network received bytes in current epoch per host
const MetricNetworkRecvBytesInCurrentEpochPerHost = "erd_network_recv_bytes_in_epoch_per_host"

// MetricNetworkSendBytesInCurrentEpochPerHost is the metric for monitoring network send bytes in current epoch per host
const MetricNetworkSendBytesInCurrentEpochPerHost = "erd_network_sent_bytes_in_epoch_per_host"

// MetricNetworkSentPercent is the metric for monitoring network sent load [%]
const MetricNetworkSentPercent = "erd_network_sent_percent"

// MetricNetworkSentBps is the metric for monitoring network sent bytes per second
const MetricNetworkSentBps = "erd_network_sent_bps"

// MetricNetworkSentBpsPeak is the metric for monitoring network sent peak bytes per second
const MetricNetworkSentBpsPeak = "erd_network_sent_bps_peak"

// MetricRoundTime is the metric for round time in seconds
const MetricRoundTime = "erd_round_time"

// MetricEpochNumber is the metric for the number of epoch
const MetricEpochNumber = "erd_epoch_number"

// MetricAppVersion is the metric for the current app version
const MetricAppVersion = "erd_app_version"

// MetricNumTxInBlock is the metric for the number of transactions in the proposed block
const MetricNumTxInBlock = "erd_num_tx_block"

// MetricConsensusState is the metric for consensus state of node proposer,participant or not consensus group
const MetricConsensusState = "erd_consensus_state"

// MetricNumMiniBlocks is the metric for number of miniblocks in a block
const MetricNumMiniBlocks = "erd_num_mini_blocks"

// MetricConsensusRoundState is the metric for consensus round state for a block
const MetricConsensusRoundState = "erd_consensus_round_state"

// MetricCrossCheckBlockHeight is the metric that store cross block height
const MetricCrossCheckBlockHeight = "erd_cross_check_block_height"

// MetricNumProcessedTxs is the metric that stores the number of transactions processed
const MetricNumProcessedTxs = "erd_num_transactions_processed"

// MetricCurrentBlockHash is the metric that stores the current block hash
const MetricCurrentBlockHash = "erd_current_block_hash"

// MetricCurrentRoundTimestamp is the metric that stores current round timestamp
const MetricCurrentRoundTimestamp = "erd_current_round_timestamp"

// MetricHeaderSize is the metric that stores the current block size
const MetricHeaderSize = "erd_current_block_size"

// MetricMiniBlocksSize is the metric that stores the current block size
const MetricMiniBlocksSize = "erd_mini_blocks_size"

// MetricNumShardHeadersFromPool is the metric that stores number of shard header from pool
const MetricNumShardHeadersFromPool = "erd_num_shard_headers_from_pool"

// MetricNumShardHeadersProcessed is the metric that stores number of shard header processed
const MetricNumShardHeadersProcessed = "erd_num_shard_headers_processed"

// MetricNumTimesInForkChoice is the metric that counts how many time a node was in fork choice
const MetricNumTimesInForkChoice = "erd_fork_choice_count"

// MetricHighestFinalBlock is the metric for the nonce of the highest final block
const MetricHighestFinalBlock = "erd_highest_final_nonce"

// MetricLatestTagSoftwareVersion is the metric that stores the latest tag software version
const MetricLatestTagSoftwareVersion = "erd_latest_tag_software_version"

// MetricCountConsensusAcceptedBlocks is the metric for monitoring number of blocks accepted when the node was in consensus group
const MetricCountConsensusAcceptedBlocks = "erd_count_consensus_accepted_blocks"

// MetricRewardsValue is the metric that stores rewards value
const MetricRewardsValue = "erd_rewards_value"

// MetricNodeDisplayName is the metric that stores the name of the node
const MetricNodeDisplayName = "erd_node_display_name"

// MetricConsensusGroupSize is the metric for consensus group size for the current shard/meta
const MetricConsensusGroupSize = "erd_consensus_group_size"

// MetricShardConsensusGroupSize is the metric for the shard consensus group size
const MetricShardConsensusGroupSize = "erd_shard_consensus_group_size"

// MetricMetaConsensusGroupSize is the metric for the metachain consensus group size
const MetricMetaConsensusGroupSize = "erd_meta_consensus_group_size"

// MetricNumNodesPerShard is the metric which holds the number of nodes in a shard
const MetricNumNodesPerShard = "erd_num_nodes_in_shard"

// MetricNumMetachainNodes is the metric which holds the number of nodes in metachain
const MetricNumMetachainNodes = "erd_num_metachain_nodes"

// MetricNumValidators is the metric for the number of validators
const MetricNumValidators = "erd_num_validators"

// MetricPeerType is the metric which tells the peer's type (in eligible list, in waiting list, or observer)
const MetricPeerType = "erd_peer_type"

// MetricPeerSubType is the metric which tells the peer's subtype (regular observer or full history observer)
const MetricPeerSubType = "erd_peer_subtype"

// MetricLeaderPercentage is the metric for leader rewards percentage
const MetricLeaderPercentage = "erd_leader_percentage"

// MetricDenomination is the metric for exposing the denomination
const MetricDenomination = "erd_denomination"

// MetricRoundAtEpochStart is the metric for storing the first round of the current epoch
const MetricRoundAtEpochStart = "erd_round_at_epoch_start"

// MetricNonceAtEpochStart is the metric for storing the first nonce of the current epoch
const MetricNonceAtEpochStart = "erd_nonce_at_epoch_start"

// MetricRoundsPerEpoch is the metric that tells the number of rounds in an epoch
const MetricRoundsPerEpoch = "erd_rounds_per_epoch"

// MetricRoundsPassedInCurrentEpoch is the metric that tells the number of rounds passed in current epoch
const MetricRoundsPassedInCurrentEpoch = "erd_rounds_passed_in_current_epoch"

// MetricNoncesPassedInCurrentEpoch is the metric that tells the number of nonces passed in current epoch
const MetricNoncesPassedInCurrentEpoch = "erd_nonces_passed_in_current_epoch"

// MetricReceivedProposedBlock is the metric that specify the moment in the round when the received block has reached the
// current node. The value is provided in percent (0 meaning it has been received just after the round started and
// 100 meaning that the block has been received in the last moment of the round)
const MetricReceivedProposedBlock = "erd_consensus_received_proposed_block"

// MetricCreatedProposedBlock is the metric that specify the percent of the block subround used for header and body
// creation (0 meaning that the block was created in no-time and 100 meaning that the block creation used all the
// subround spare duration)
const MetricCreatedProposedBlock = "erd_consensus_created_proposed_block"

// MetricProcessedProposedBlock is the metric that specify the percent of the block subround used for header and body
// processing (0 meaning that the block was processed in no-time and 100 meaning that the block processing used all the
// subround spare duration)
const MetricProcessedProposedBlock = "erd_consensus_processed_proposed_block"

// MetricMinGasPrice is the metric that specifies min gas price
const MetricMinGasPrice = "erd_min_gas_price"

// MetricMinGasLimit is the metric that specifies the minimum gas limit
const MetricMinGasLimit = "erd_min_gas_limit"

// MetricRewardsTopUpGradientPoint is the metric that specifies the rewards top up gradient point
const MetricRewardsTopUpGradientPoint = "erd_rewards_top_up_gradient_point"

// MetricGasPriceModifier is the metric that specifies the gas price modifier
const MetricGasPriceModifier = "erd_gas_price_modifier"

// MetricTopUpFactor is the metric that specifies the top up factor
const MetricTopUpFactor = "erd_top_up_factor"

// MetricMinTransactionVersion is the metric that specifies the minimum transaction version
const MetricMinTransactionVersion = "erd_min_transaction_version"

// MetricGasPerDataByte is the metric that specifies the required gas for a data byte
const MetricGasPerDataByte = "erd_gas_per_data_byte"

// MetricMaxGasPerTransaction is the metric that specifies the maximum gas limit for a transaction
const MetricMaxGasPerTransaction = "erd_max_gas_per_transaction"

// MetricChainId is the metric that specifies current chain id
const MetricChainId = "erd_chain_id"

// MetricStartTime is the metric that specifies the genesis start time
const MetricStartTime = "erd_start_time"

// MetricRoundDuration is the metric that specifies the round duration in milliseconds
const MetricRoundDuration = "erd_round_duration"

// MetricTotalSupply holds the total supply value for the last epoch
const MetricTotalSupply = "erd_total_supply"

// MetricTotalBaseStakedValue holds the total base staked value
const MetricTotalBaseStakedValue = "erd_total_base_staked_value"

// MetricTopUpValue holds the total top up value
const MetricTopUpValue = "erd_total_top_up_value"

// MetricInflation holds the inflation value for the last epoch
const MetricInflation = "erd_inflation"

// MetricDevRewardsInEpoch holds the developers' rewards value for the last epoch
const MetricDevRewardsInEpoch = "erd_dev_rewards"

// MetricTotalFees holds the total fees value for the last epoch
const MetricTotalFees = "erd_total_fees"

// MetricEpochForEconomicsData holds the epoch for which economics data are computed
const MetricEpochForEconomicsData = "erd_epoch_for_economics_data"

// MetachainShardId will be used to identify a shard ID as metachain
const MetachainShardId = uint32(0xFFFFFFFF)

// BaseOperationCost represents the field name for base operation costs
const BaseOperationCost = "BaseOperationCost"

// BuiltInCost represents the field name for built in operation costs
const BuiltInCost = "BuiltInCost"

// MetaChainSystemSCsCost represents the field name for metachain system smart contract operation costs
const MetaChainSystemSCsCost = "MetaChainSystemSCsCost"

// ElrondAPICost represents the field name of the Elrond SC API (EEI) gas costs
const ElrondAPICost = "ElrondAPICost"

// AsyncCallStepField is the field name for the gas cost for any of the two steps required to execute an async call
const AsyncCallStepField = "AsyncCallStep"

// AsyncCallbackGasLockField is the field name for the gas amount to be locked
// before executing the destination async call, to be put aside for the async callback
const AsyncCallbackGasLockField = "AsyncCallbackGasLock"

const (
	// MetricScDeployEnableEpoch represents the epoch when the deployment of smart contracts is enabled
	MetricScDeployEnableEpoch = "erd_smart_contract_deploy_enable_epoch"

	// MetricBuiltInFunctionsEnableEpoch represents the epoch when the built in functions is enabled
	MetricBuiltInFunctionsEnableEpoch = "erd_built_in_functions_enable_epoch"

	// MetricRelayedTransactionsEnableEpoch represents the epoch when the relayed transactions is enabled
	MetricRelayedTransactionsEnableEpoch = "erd_relayed_transactions_enable_epoch"

	// MetricPenalizedTooMuchGasEnableEpoch represents the epoch when the penalization for using too much gas is enabled
	MetricPenalizedTooMuchGasEnableEpoch = "erd_penalized_too_much_gas_enable_epoch"

	// MetricSwitchJailWaitingEnableEpoch represents the epoch when the system smart contract processing at end of epoch is enabled
	MetricSwitchJailWaitingEnableEpoch = "erd_switch_jail_waiting_enable_epoch"

	// MetricSwitchHysteresisForMinNodesEnableEpoch represents the epoch when the system smart contract changes its config to consider
	// also (minimum) hysteresis nodes for the minimum number of nodes
	MetricSwitchHysteresisForMinNodesEnableEpoch = "erd_switch_hysteresis_for_min_nodes_enable_epoch"

	// MetricBelowSignedThresholdEnableEpoch represents the epoch when the change for computing rating for validators below signed rating is enabled
	MetricBelowSignedThresholdEnableEpoch = "erd_below_signed_threshold_enable_epoch"

	// MetricTransactionSignedWithTxHashEnableEpoch represents the epoch when the node will also accept transactions that are
	// signed with the hash of transaction
	MetricTransactionSignedWithTxHashEnableEpoch = "erd_transaction_signed_with_txhash_enable_epoch"

	// MetricMetaProtectionEnableEpoch represents the epoch when the transactions to the metachain are checked to have enough gas
	MetricMetaProtectionEnableEpoch = "erd_meta_protection_enable_epoch"

	// MetricAheadOfTimeGasUsageEnableEpoch represents the epoch when the cost of smart contract prepare changes from compiler
	// per byte to ahead of time prepare per byte
	MetricAheadOfTimeGasUsageEnableEpoch = "erd_ahead_of_time_gas_usage_enable_epoch"

	// MetricGasPriceModifierEnableEpoch represents the epoch when the gas price modifier in fee computation is enabled
	MetricGasPriceModifierEnableEpoch = "erd_gas_price_modifier_enable_epoch"

	// MetricRepairCallbackEnableEpoch represents the epoch when the callback repair is activated for smart contract results
	MetricRepairCallbackEnableEpoch = "erd_repair_callback_enable_epoch"

	// MetricBlockGasAndFreeRecheckEnableEpoch represents the epoch when gas and fees used in each created or processed block are re-checked
	MetricBlockGasAndFreeRecheckEnableEpoch = "erd_block_gas_and_fee_recheck_enable_epoch"

	// MetricStakingV2EnableEpoch represents the epoch when staking v2 is enabled
	MetricStakingV2EnableEpoch = "erd_staking_v2_enable_epoch"

	// MetricStakeEnableEpoch represents the epoch when staking is enabled
	MetricStakeEnableEpoch = "erd_stake_enable_epoch"

	// MetricDoubleKeyProtectionEnableEpoch represents the epoch when double key protection is enabled
	MetricDoubleKeyProtectionEnableEpoch = "erd_double_key_protection_enable_epoch"

	// MetricEsdtEnableEpoch represents the epoch when ESDT is enabled
	MetricEsdtEnableEpoch = "erd_esdt_enable_epoch"

	// MetricGovernanceEnableEpoch  represents the epoch when governance is enabled
	MetricGovernanceEnableEpoch = "erd_governance_enable_epoch"

	// MetricDelegationManagerEnableEpoch represents the epoch when the delegation manager is enabled
	MetricDelegationManagerEnableEpoch = "erd_delegation_manager_enable_epoch"

	// MetricDelegationSmartContractEnableEpoch represents the epoch when delegation smart contract is enabled
	MetricDelegationSmartContractEnableEpoch = "erd_delegation_smart_contract_enable_epoch"

	// MetricCorrectLastUnjailedEnableEpoch represents the epoch when the correction on the last unjailed node is applied
	MetricCorrectLastUnjailedEnableEpoch = "erd_correct_last_unjailed_enable_epoch"

	// MetricBalanceWaitingListsEnableEpoch represents the epoch when the balance waiting lists on shards fix is applied
	MetricBalanceWaitingListsEnableEpoch = "erd_balance_waiting_lists_enable_epoch"

	// MetricReturnDataToLastTransferEnableEpoch represents the epoch when the return data to last transfer is applied
	MetricReturnDataToLastTransferEnableEpoch = "erd_return_data_to_last_transfer_enable_epoch"

	// MetricSenderInOutTransferEnableEpoch represents the epoch when the sender in out transfer is applied
	MetricSenderInOutTransferEnableEpoch = "erd_sender_in_out_transfer_enable_epoch"

	// MetricRelayedTransactionsV2EnableEpoch represents the epoch when the relayed transactions v2 is enabled
	MetricRelayedTransactionsV2EnableEpoch = "erd_relayed_transactions_v2_enable_epoch"

	// MetricUnbondTokensV2EnableEpoch represents the epoch when the unbond tokens v2 is applied
	MetricUnbondTokensV2EnableEpoch = "erd_unbond_tokens_v2_enable_epoch"

	// MetricSaveJailedAlwaysEnableEpoch represents the epoch the save jailed fix is applied
	MetricSaveJailedAlwaysEnableEpoch = "erd_save_jailed_always_enable_epoch"

	// MetricValidatorToDelegationEnableEpoch represents the epoch when the validator to delegation feature (staking v3.5) is enabled
	MetricValidatorToDelegationEnableEpoch = "erd_validator_to_delegation_enable_epoch"

	// MetricReDelegateBelowMinCheckEnableEpoch represents the epoch when the re-delegation below minimum value fix is applied
	MetricReDelegateBelowMinCheckEnableEpoch = "erd_redelegate_below_min_check_enable_epoch"

	// MetricIncrementSCRNonceInMultiTransferEnableEpoch represents the epoch when the fix for multi transfer SCR is enabled
	MetricIncrementSCRNonceInMultiTransferEnableEpoch = "erd_increment_scr_nonce_in_multi_transfer_enable_epoch"

	// MetricESDTMultiTransferEnableEpoch represents the epoch when the ESDT multi transfer feature is enabled
	MetricESDTMultiTransferEnableEpoch = "erd_esdt_multi_transfer_enable_epoch"

	// MetricGlobalMintBurnDisableEpoch represents the epoch when the global mint and burn feature is disabled
	MetricGlobalMintBurnDisableEpoch = "erd_global_mint_burn_disable_epoch"

	// MetricESDTTransferRoleEnableEpoch represents the epoch when the ESDT transfer role feature is enabled
	MetricESDTTransferRoleEnableEpoch = "erd_esdt_transfer_role_enable_epoch"

	// MetricBuiltInFunctionOnMetaEnableEpoch represents the epoch when the builtin functions on metachain are enabled
	MetricBuiltInFunctionOnMetaEnableEpoch = "erd_builtin_function_on_meta_enable_epoch"
)

const (
	// StorerOrder defines the order of storers to be notified of a start of epoch event
	StorerOrder = iota
	// NodesCoordinatorOrder defines the order in which NodesCoordinator is notified of a start of epoch event
	NodesCoordinatorOrder
	// ConsensusOrder defines the order in which Consensus is notified of a start of epoch event
	ConsensusOrder
	// NetworkShardingOrder defines the order in which the network sharding subsystem is notified of a start of epoch event
	NetworkShardingOrder
	// IndexerOrder defines the order in which indexer is notified of a start of epoch event
	IndexerOrder
	// NetStatisticsOrder defines the order in which netStatistic component is notified of a start of epoch event
	NetStatisticsOrder
	// OldDatabaseCleanOrder defines the order in which oldDatabaseCleaner component is notified of a start of epoch event
	OldDatabaseCleanOrder
)

// NodeState specifies what type of state a node could have
type NodeState int

const (
	// NsSynchronized defines ID of a state of synchronized
	NsSynchronized NodeState = iota
	// NsNotSynchronized defines ID of a state of not synchronized
	NsNotSynchronized
	// NsNotCalculated defines ID of a state which is not calculated
	NsNotCalculated
)

// MetricP2PPeerInfo is the metric for the node's p2p info
const MetricP2PPeerInfo = "erd_p2p_peer_info"

// MetricP2PIntraShardValidators is the metric that outputs the intra-shard connected validators
const MetricP2PIntraShardValidators = "erd_p2p_intra_shard_validators"

// MetricP2PCrossShardValidators is the metric that outputs the cross-shard connected validators
const MetricP2PCrossShardValidators = "erd_p2p_cross_shard_validators"

// MetricP2PIntraShardObservers is the metric that outputs the intra-shard connected observers
const MetricP2PIntraShardObservers = "erd_p2p_intra_shard_observers"

// MetricP2PCrossShardObservers is the metric that outputs the cross-shard connected observers
const MetricP2PCrossShardObservers = "erd_p2p_cross_shard_observers"

// MetricP2PFullHistoryObservers is the metric that outputs the full-history connected observers
const MetricP2PFullHistoryObservers = "erd_p2p_full_history_observers"

// MetricP2PUnknownPeers is the metric that outputs the unknown-shard connected peers
const MetricP2PUnknownPeers = "erd_p2p_unknown_shard_peers"

// MetricP2PNumConnectedPeersClassification is the metric for monitoring the number of connected peers split on the connection type
const MetricP2PNumConnectedPeersClassification = "erd_p2p_num_connected_peers_classification"

// HighestRoundFromBootStorage is the key for the highest round that is saved in storage
const HighestRoundFromBootStorage = "highestRoundFromBootStorage"

// TriggerRegistryKeyPrefix is the key prefix to save epoch start registry to storage
const TriggerRegistryKeyPrefix = "epochStartTrigger_"

// TriggerRegistryInitialKeyPrefix is the key prefix to save initial data to storage
const TriggerRegistryInitialKeyPrefix = "initial_value_epoch_"

// NodesCoordinatorRegistryKeyPrefix is the key prefix to save epoch start registry to storage
const NodesCoordinatorRegistryKeyPrefix = "indexHashed_"

// ShuffledOut signals that a restart is pending because the node was shuffled out
const ShuffledOut = "shuffledOut"

// WrongConfiguration signals that the node has a malformed configuration and cannot continue processing
const WrongConfiguration = "wrongConfiguration"

// ImportComplete signals that a node restart will be done because the import did complete
const ImportComplete = "importComplete"

// MaxRetriesToCreateDB represents the maximum number of times to try to create DB if it failed
const MaxRetriesToCreateDB = 10

// SleepTimeBetweenCreateDBRetries represents the number of seconds to sleep between DB creates
const SleepTimeBetweenCreateDBRetries = 5 * time.Second

// DefaultStatsPath is the default path where the node stats are logged
const DefaultStatsPath = "stats"

// DefaultDBPath is the default path for nodes databases
const DefaultDBPath = "db"

// DefaultEpochString is the default folder root name for node per epoch databases
const DefaultEpochString = "Epoch"

// DefaultStaticDbString is the default name for the static databases (not changing with epoch)
const DefaultStaticDbString = "Static"

// DefaultShardString is the default folder root name for per shard databases
const DefaultShardString = "Shard"

// MetachainShardName is the string identifier of the metachain shard
const MetachainShardName = "metachain"

// TemporaryPath is the default temporary path directory
const TemporaryPath = "temp"

// TimeToWaitForP2PBootstrap is the wait time for the P2P to bootstrap
const TimeToWaitForP2PBootstrap = 20 * time.Second

// MaxSoftwareVersionLengthInBytes represents the maximum length for the software version to be saved in block header
const MaxSoftwareVersionLengthInBytes = 10

// ExtraDelayForBroadcastBlockInfo represents the number of seconds to wait since a block has been broadcast and the
// moment when its components, like mini blocks and transactions, would be broadcast too
const ExtraDelayForBroadcastBlockInfo = 1 * time.Second

// ExtraDelayBetweenBroadcastMbsAndTxs represents the number of seconds to wait since miniblocks have been broadcast
// and the moment when theirs transactions would be broadcast too
const ExtraDelayBetweenBroadcastMbsAndTxs = 1 * time.Second

// ExtraDelayForRequestBlockInfo represents the number of seconds to wait since a block has been received and the
// moment when its components, like mini blocks and transactions, would be requested too if they are still missing
const ExtraDelayForRequestBlockInfo = ExtraDelayForBroadcastBlockInfo + ExtraDelayBetweenBroadcastMbsAndTxs + time.Second

// CommitMaxTime represents max time accepted for a commit action, after which a warn message is displayed
const CommitMaxTime = 3 * time.Second

// PutInStorerMaxTime represents max time accepted for a put action, after which a warn message is displayed
const PutInStorerMaxTime = time.Second

// DefaultUnstakedEpoch represents the default epoch that is set for a validator that has not unstaked yet
const DefaultUnstakedEpoch = math.MaxUint32

// InvalidMessageBlacklistDuration represents the time to keep a peer in the black list if it sends a message that
// does not follow the protocol: example not useing the same marshaler as the other peers
const InvalidMessageBlacklistDuration = time.Second * 3600

// PublicKeyBlacklistDuration represents the time to keep a public key in the black list if it will degrade its
// rating to a minimum threshold due to improper messages
const PublicKeyBlacklistDuration = time.Second * 7200

// WrongP2PMessageBlacklistDuration represents the time to keep a peer id in the blacklist if it sends a message that
// do not follow this protocol
const WrongP2PMessageBlacklistDuration = time.Second * 7200

// MaxWaitingTimeToReceiveRequestedItem represents the maximum waiting time in seconds needed to receive the requested items
const MaxWaitingTimeToReceiveRequestedItem = 5 * time.Second

// DefaultLogProfileIdentifier represents the default log profile used when the logviewer/termui applications do not
// need to change the current logging profile
const DefaultLogProfileIdentifier = "[default log profile]"

// NotSetDestinationShardID represents the shardIdString when the destinationShardId is not set in the prefs
const NotSetDestinationShardID = "disabled"

// AdditionalScrForEachScCallOrSpecialTx specifies the additional number of smart contract results which should be
// considered by a node, when it includes sc calls or special txs in a miniblock.
// Ex.: normal txs -> aprox. 27000, sc calls or special txs -> aprox. 6250 = 27000 / (AdditionalScrForEachScCallOrSpecialTx + 1),
// considering that constant below is set to 3
const AdditionalScrForEachScCallOrSpecialTx = 3

// MaxRoundsWithoutCommittedStartInEpochBlock defines the maximum rounds to wait for start in epoch block to be committed,
// before a special action to be applied
const MaxRoundsWithoutCommittedStartInEpochBlock = 50

// DefaultResolversIdentifier represents the identifier that is used in conjunction with regular resolvers
// (that makes the node run properly)
const DefaultResolversIdentifier = "default resolver"

// DefaultInterceptorsIdentifier represents the identifier that is used in conjunction with regular interceptors
// (that makes the node run properly)
const DefaultInterceptorsIdentifier = "default interceptor"

// HardforkInterceptorsIdentifier represents the identifier that is used in the hardfork process
const HardforkInterceptorsIdentifier = "hardfork interceptor"

// HardforkResolversIdentifier represents the resolver that is used in the hardfork process
const HardforkResolversIdentifier = "hardfork resolver"

// EpochStartInterceptorsIdentifier represents the identifier that is used in the start-in-epoch process
const EpochStartInterceptorsIdentifier = "epoch start interceptor"

// GetNodeFromDBErrorString represents the string which is returned when a getting node from DB returns an error
const GetNodeFromDBErrorString = "getNodeFromDB error"

// TimeoutGettingTrieNodes defines the timeout in trie sync operation if no node is received
const TimeoutGettingTrieNodes = 2 * time.Minute // to consider syncing a very large trie node of 64MB at ~1MB/s

// TimeoutGettingTrieNodesInHardfork represents the maximum time allowed between 2 nodes fetches (and commits)
// during the hardfork process
const TimeoutGettingTrieNodesInHardfork = time.Minute * 10

// RetrialIntervalForOutportDriver is the interval in which the outport driver should try to call the driver again
const RetrialIntervalForOutportDriver = time.Second * 10

// NodeProcessingMode represents the processing mode in which the node was started
type NodeProcessingMode int

const (
	// Normal means that the node has started in the normal processing mode
	Normal NodeProcessingMode = iota

	// ImportDb means that the node has started in the import-db mode
	ImportDb
)

// ScheduledMode represents the name used to differentiate normal vs. scheduled mini blocks / transactions execution mode
const ScheduledMode = "Scheduled"

const (
	// ActiveDBKey is the key at which ActiveDBVal will be saved
	ActiveDBKey = "activeDB"

	// ActiveDBVal is the value that will be saved at ActiveDBKey
	ActiveDBVal = "yes"

	// TrieSyncedKey is the key at which TrieSyncedVal will be saved
	TrieSyncedKey = "synced"

	// TrieSyncedVal is the value that will be saved at TrieSyncedKey
	TrieSyncedVal = "yes"
)
