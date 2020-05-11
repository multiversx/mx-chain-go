package dataRetriever

// TxPoolNumSendersToEvictInOneStep instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToEvictInOneStep = uint32(100)

// TxPoolMinSizeInBytes is the lower limit of the tx cache / eviction parameter "sizeInBytes"
const TxPoolMinSizeInBytes = uint32(40960)
