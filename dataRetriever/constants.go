package dataRetriever

// TxPoolNumSendersToEvictInOneStep instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToEvictInOneStep = uint32(100)

// TxPoolNumTxsToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many transactions when eviction takes place
const TxPoolNumTxsToPreemptivelyEvict = uint32(1000)
