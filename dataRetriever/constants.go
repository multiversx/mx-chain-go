package dataRetriever

// TxPoolNumSendersToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToPreemptivelyEvict = uint32(100)

// TxPoolNumTxsToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many transactions when eviction takes place
const TxPoolNumTxsToPreemptivelyEvict = uint32(1000)

// UnsignedTxPoolName defines the name of the unsigned transactions pool
const UnsignedTxPoolName = "uTxPool"

// RewardTxPoolName defines the name of the reward transactions pool
const RewardTxPoolName = "rewardTxPool"
