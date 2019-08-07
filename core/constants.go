package core

// pkPrefixSize specifies the max numbers of chars to be displayed from one publc key
const pkPrefixSize = 12

// MaxBulkTransactionSize specifies the maximum size of one bulk with txs which can be send over the network
//TODO convert this const into a var and read it from config when this code moves to another binary
const MaxBulkTransactionSize = 2 << 17 //128KB bulks

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic = "consensus"

// MetricCurrentRound is the metric for monitoring the current round of a node
const MetricCurrentRound = "erd_current_round"

// MetricNonce is the metric for monitoring the nonce of a node
const MetricNonce = "erd_nonce"

// MetricNumConnectedPeers is the metric for monitoring the number of connected peers
const MetricNumConnectedPeers = "erd_num_connected_peers"

// MetricSynchronizedRound is the metric for monitoring the synchronized round of a node
const MetricSynchronizedRound = "erd_synchronized_round"

// MetricIsSyncing is the metric for monitoring if a node is syncing
const MetricIsSyncing = "erd_is_syncing"
