package dataRetriever

// TxPoolNumSendersToEvictInOneStep instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToEvictInOneStep = uint32(100)

// TxPoolLargeNumOfTxsForASender instructs tx pool eviction algorithm to tag a sender with more transactions than this value
// as a "sender with a large number of transactions"
const TxPoolLargeNumOfTxsForASender = uint32(500)

// TxPoolNumTxsToEvictFromASender instructs tx pool eviction algorithm to remove this many transactions
// for "a sender with a large number of transactions" when eviction takes place
const TxPoolNumTxsToEvictFromASender = uint32(100)

// TxPoolMinSizeInBytes is the lower limit of the tx cache / eviction parameter "sizeInBytes"
const TxPoolMinSizeInBytes = uint32(40960)
