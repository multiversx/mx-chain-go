package storage

import "github.com/ElrondNetwork/elrond-go-storage/common"

// TxPoolNumTxsToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many transactions when eviction takes place
const TxPoolNumTxsToPreemptivelyEvict = common.TxPoolNumTxsToPreemptivelyEvict

// DefaultEpochString is the default folder root name for node per epoch databases
const DefaultEpochString = common.DefaultEpochString

// DefaultShardString is the default folder root name for per shard databases
const DefaultShardString = common.DefaultShardString

// MaxRetriesToCreateDB represents the maximum number of times to try to create DB if it failed
const MaxRetriesToCreateDB = common.MaxRetriesToCreateDB

// SleepTimeBetweenCreateDBRetries represents the number of seconds to sleep between DB creates
const SleepTimeBetweenCreateDBRetries = common.SleepTimeBetweenCreateDBRetries
