package common

// MetricP2PPeerNumReceivedMessages represents the current maximum number of received messages in the amount of time
// counted on a connected peer
const MetricP2PPeerNumReceivedMessages = "erd_p2p_peer_num_received_messages"

// MetricP2PPeerSizeReceivedMessages represents the current maximum size of received data (sum of all messages) in
// the amount of time counted on a connected peer
const MetricP2PPeerSizeReceivedMessages = "erd_p2p_peer_size_received_messages"

// MetricP2PPeerNumProcessedMessages represents the current maximum number of processed messages in the amount of time
// counted on a connected peer
const MetricP2PPeerNumProcessedMessages = "erd_p2p_peer_num_processed_messages"

// MetricP2PPeerSizeProcessedMessages represents the current maximum size of processed data (sum of all messages) in
// the amount of time counted on a connected peer
const MetricP2PPeerSizeProcessedMessages = "erd_p2p_peer_size_processed_messages"

// MetricP2PPeakPeerNumReceivedMessages represents the peak maximum number of received messages in the amount of time
// counted on a connected peer
const MetricP2PPeakPeerNumReceivedMessages = "erd_p2p_peak_peer_num_received_messages"

// MetricP2PPeakPeerSizeReceivedMessages represents the peak maximum size of received data (sum of all messages) in
// the amount of time counted on a connected peer
const MetricP2PPeakPeerSizeReceivedMessages = "erd_p2p_peak_peer_size_received_messages"

// MetricP2PPeakPeerNumProcessedMessages represents the peak maximum number of processed messages in the amount of time
// counted on a connected peer
const MetricP2PPeakPeerNumProcessedMessages = "erd_p2p_peak_peer_num_processed_messages"

// MetricP2PPeakPeerSizeProcessedMessages represents the peak maximum size of processed data (sum of all messages) in
// the amount of time counted on a connected peer
const MetricP2PPeakPeerSizeProcessedMessages = "erd_p2p_peak_peer_size_processed_messages"

// MetricP2PNumReceiverPeers represents the number of connected peer sent messages to the current peer (and have been
// received by the current peer) in the amount of time
const MetricP2PNumReceiverPeers = "erd_p2p_num_receiver_peers"

// MetricP2PPeakNumReceiverPeers represents the peak number of connected peer sent messages to the current peer
// (and have been received by the current peer) in the amount of time
const MetricP2PPeakNumReceiverPeers = "erd_p2p_peak_num_receiver_peers"
