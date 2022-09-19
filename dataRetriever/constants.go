package dataRetriever

// TxPoolNumSendersToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many senders when eviction takes place
const TxPoolNumSendersToPreemptivelyEvict = uint32(100)

// UnsignedTxPoolName defines the name of the unsigned transactions pool
const UnsignedTxPoolName = "uTxPool"

// RewardTxPoolName defines the name of the reward transactions pool
const RewardTxPoolName = "rewardTxPool"

// ValidatorsInfoPoolName defines the name of the validators info pool
const ValidatorsInfoPoolName = "validatorsInfoPool"
