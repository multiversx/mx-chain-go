package consensus

import "time"

// TODO: better handling for duplicated constants, evaluare moving to core or to some other common repo

const (
	// TransactionTopic is the topic used for sharing transactions
	TransactionTopic = "transactions"
	// ShardBlocksTopic is the topic used for sharing block headers
	ShardBlocksTopic = "shardBlocks"
	// MiniBlocksTopic is the topic used for sharing mini blocks
	MiniBlocksTopic = "txBlockBodies"
	// MetachainBlocksTopic is the topic used for sharing metachain block headers
	MetachainBlocksTopic = "metachainBlocks"
	// ConsensusTopic is the topic used in consensus algorithm
	ConsensusTopic = "consensus"
)

// ExtraDelayForBroadcastBlockInfo represents the number of seconds to wait since a block has been broadcast and the
// moment when its components, like mini blocks and transactions, would be broadcast too
const ExtraDelayForBroadcastBlockInfo = 1 * time.Second

// ExtraDelayBetweenBroadcastMbsAndTxs represents the number of seconds to wait since miniblocks have been broadcast
// and the moment when theirs transactions would be broadcast too
const ExtraDelayBetweenBroadcastMbsAndTxs = 1 * time.Second

// InvalidMessageBlacklistDuration represents the time to keep a peer in the black list if it sends a message that
// does not follow the protocol: example not useing the same marshaler as the other peers
const InvalidMessageBlacklistDuration = time.Second * 3600

// MaxBulkTransactionSize specifies the maximum size of one bulk with txs which can be send over the network
// TODO convert this const into a var and read it from config when this code moves to another binary
const MaxBulkTransactionSize = 1 << 18 // 256KB bulks

// MetricCurrentRound is the metric for monitoring the current round of a node
const MetricCurrentRound = "erd_current_round"

// MetricCurrentRoundTimestamp is the metric that stores current round timestamp
const MetricCurrentRoundTimestamp = "erd_current_round_timestamp"

// MetricConsensusRoundState is the metric for consensus round state for a block
const MetricConsensusRoundState = "erd_consensus_round_state"

// MetricCountLeader is the metric for monitoring number of rounds when a node was leader
const MetricCountLeader = "erd_count_leader"

// MetricConsensusState is the metric for consensus state of node proposer,participant or not consensus group
const MetricConsensusState = "erd_consensus_state"

// MetricCountConsensus is the metric for monitoring number of rounds when a node was in consensus group
const MetricCountConsensus = "erd_count_consensus"

// MetricReceivedProposedBlock is the metric that specifies the moment in the round when the received block has reached the
// current node. The value is provided in percent (0 meaning it has been received just after the round started and
// 100 meaning that the block has been received in the last moment of the round)
const MetricReceivedProposedBlock = "erd_consensus_received_proposed_block"

// MetricRedundancyIsMainActive is the metrics that specifies data about the redundancy main machine
const MetricRedundancyIsMainActive = "erd_redundancy_is_main_active"

// MetricCountAcceptedBlocks is the metric for monitoring number of blocks that was accepted proposed by a node
const MetricCountAcceptedBlocks = "erd_count_accepted_blocks"

// MetricCreatedProposedBlock is the metric that specifies the percent of the block subround used for header and body
// creation (0 meaning that the block was created in no-time and 100 meaning that the block creation used all the
// subround spare duration)
const MetricCreatedProposedBlock = "erd_consensus_created_proposed_block"

//MetricProcessedProposedBlock is the metric that specify the percent of the block subround used for header and body
// processing (0 meaning that the block was processed in no-time and 100 meaning that the block processing used all the
// subround spare duration)
const MetricProcessedProposedBlock = "erd_consensus_processed_proposed_block"

// CommitMaxTime represents max time accepted for a commit action, after which a warn message is displayed
const CommitMaxTime = 3 * time.Second
