package dataRetriever

// TxPoolNumSendersToEvictInOneStep instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToEvictInOneStep = uint32(100)

// TxPoolALotOfTransactionsForASender instructs tx pool eviction algorithm to tag a sender with more transactions than this value
// as a "sender with a lot of transactions"
const TxPoolALotOfTransactionsForASender = uint32(500)

// TxPoolNumTxsToEvictForASenderWithALot instructs tx pool eviction algorithm to remove this many transactions
// for "a sender with a lot of transactions" when eviction takes place
const TxPoolNumTxsToEvictForASenderWithALot = uint32(100)

// TxPoolMinSizeInBytes is the lower limit of the tx cache / eviction parameter "sizeInBytes"
const TxPoolMinSizeInBytes = uint32(40960)
