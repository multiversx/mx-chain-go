package core

// pkPrefixSize specifies the max numbers of chars to be displayed from one publc key
const pkPrefixSize = 12

// MaxBulkTransactionSize specifies the maximum size of one bulk with txs which can be send over the network
//TODO convert this const into a var and read it from config when this code moves to another binary
const MaxBulkTransactionSize = 2 << 17 //128KB bulks

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic = "consensus"
