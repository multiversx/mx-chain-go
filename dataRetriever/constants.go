package dataRetriever

// TxPoolNumSendersToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToPreemptivelyEvict = uint32(100)

// TxPoolNumTxsToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many transactions when eviction takes place
const TxPoolNumTxsToPreemptivelyEvict = uint32(1000)
