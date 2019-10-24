package core

// UnVersionedAppString represents the default app version that indicate that the binary wasn't build by setting
// the appVersion flag
const UnVersionedAppString = "undefined"

// NodeType represents the node's role in the network
type NodeType string

// NodeTypeObserver signals that a node is running as observer node
const NodeTypeObserver NodeType = "observer"

// NodeTypeValidator signals that a node is running as validator node
const NodeTypeValidator NodeType = "validator"

// pkPrefixSize specifies the max numbers of chars to be displayed from one publc key
const pkPrefixSize = 12

// MaxBulkTransactionSize specifies the maximum size of one bulk with txs which can be send over the network
//TODO convert this const into a var and read it from config when this code moves to another binary
const MaxBulkTransactionSize = 2 << 17 //128KB bulks

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic = "consensus"

// GenesisBlockNonce is the nonce of the genesis block
const GenesisBlockNonce = 0

// MetricCurrentRound is the metric for monitoring the current round of a node
const MetricCurrentRound = "erd_current_round"

// MetricNonce is the metric for monitoring the nonce of a node
const MetricNonce = "erd_nonce"

// MetricProbableHighestNonce is the metric for monitoring the max speculative nonce received by the node by listening on the network
const MetricProbableHighestNonce = "erd_probable_highest_nonce"

// MetricNumConnectedPeers is the metric for monitoring the number of connected peers
const MetricNumConnectedPeers = "erd_num_connected_peers"

// MetricSynchronizedRound is the metric for monitoring the synchronized round of a node
const MetricSynchronizedRound = "erd_synchronized_round"

// MetricIsSyncing is the metric for monitoring if a node is syncing
const MetricIsSyncing = "erd_is_syncing"

// MetricPublicKeyBlockSign is the metric for monitoring public key of a node used in block signing
const MetricPublicKeyBlockSign = "erd_public_key_block_sign"

// MetricPublicKeyTxSign is the metric for monitoring public key of a node used in tx signing
// (balance account held by the node)
const MetricPublicKeyTxSign = "erd_public_key_tx_sign"

// MetricShardId is the metric for monitoring shard id of a node
const MetricShardId = "erd_shard_id"

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

// MetricConnectedNodes is the metric for monitoring total connected peers on the network
const MetricConnectedNodes = "erd_connected_nodes"

// MetricCpuLoadPercent is the metric for monitoring CPU load [%]
const MetricCpuLoadPercent = "erd_cpu_load_percent"

// MetricMemLoadPercent is the metric for monitoring memory load [%]
const MetricMemLoadPercent = "erd_mem_load_percent"

// MetricMemTotal is the metric for monitoring total memory bytes
const MetricMemTotal = "erd_mem_total"

// MetricMemUsedGolang is the metric that stores the total memory used by golang in bytes
const MetricMemUsedGolang = "erd_mem_used_golang"

// MetricMemUsedSystem is the metric that stores the total memory used by the system in bytes
const MetricMemUsedSystem = "erd_mem_used_sys"

// MetricNetworkRecvPercent is the metric for monitoring network receive load [%]
const MetricNetworkRecvPercent = "erd_network_recv_percent"

// MetricNetworkRecvBps is the metric for monitoring network received bytes per second
const MetricNetworkRecvBps = "erd_network_recv_bps"

// MetricNetworkRecvBpsPeak is the metric for monitoring network received peak bytes per second
const MetricNetworkRecvBpsPeak = "erd_network_recv_bps_peak"

// MetricNetworkSentPercent is the metric for monitoring network sent load [%]
const MetricNetworkSentPercent = "erd_network_sent_percent"

// MetricNetworkSentBps is the metric for monitoring network sent bytes per second
const MetricNetworkSentBps = "erd_network_sent_bps"

// MetricNetworkSentBpsPeak is the metric for monitoring network sent peak bytes per second
const MetricNetworkSentBpsPeak = "erd_network_sent_bps_peak"

// MetricRoundTime is the metric for round time in seconds
const MetricRoundTime = "erd_round_time"

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
const MetricCrossCheckBlockHeight = "erd_metric_cross_check_block_height"

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

// MaxMiniBlocksInBlock specifies the max number of mini blocks which can be added in one block
const MaxMiniBlocksInBlock = 100

//MetricHighestFinalBlockInShard is the metric that stores the highest nonce block notarized by metachain for current shard
const MetricHighestFinalBlockInShard = "erd_highest_notarized_block_by_metachain_for_current_shard"

//MetricLatestTagSoftwareVersion is the metric that stores the latest tag software version
const MetricLatestTagSoftwareVersion = "erd_latest_tag_software_version"

//MetricCountConsensusAcceptedBlocks is the metric for monitoring number of blocks accepted when the node was in consensus group
const MetricCountConsensusAcceptedBlocks = "erd_count_consensus_accepted_blocks"

//MetricRewardsValue is the metric that stores rewards value
const MetricRewardsValue = "erd_rewards_value"

//MetricNodeDisplayName is the metric that stores the name of the node
const MetricNodeDisplayName = "erd_node_display_name"

//MetricConsensusGroupSize is the metric for consensus group size
const MetricConsensusGroupSize = "erd_metric_consensus_group_size"

//MetricNumValidators is the metric for the number of validators
const MetricNumValidators = "erd_metric_num_validators"

//MetricLeaderPercentage is the metric for leader rewards percentage
const MetricLeaderPercentage = "erd_metric_leader_percentage"

//MetricCommunityPercentage is the metric for community rewards percentage
const MetricCommunityPercentage = "erd_metric_community_percentage"
