package storage

// PathShardPlaceholder represents the placeholder for the shard ID in paths
const PathShardPlaceholder = "[S]"

// PathEpochPlaceholder represents the placeholder for the epoch number in paths
const PathEpochPlaceholder = "[E]"

// PathIdentifierPlaceholder represents the placeholder for the identifier in paths
const PathIdentifierPlaceholder = "[I]"

// TxPoolNumTxsToPreemptivelyEvict instructs tx pool eviction algorithm to remove this many transactions when eviction takes place
const TxPoolNumTxsToPreemptivelyEvict = uint32(1000)

// DefaultDBPath is the default path for nodes databases
const DefaultDBPath = "db"

// DefaultEpochString is the default folder root name for node per epoch databases
const DefaultEpochString = "Epoch"

// DefaultStaticDbString is the default name for the static databases (not changing with epoch)
const DefaultStaticDbString = "Static"

// DefaultShardString is the default folder root name for per shard databases
const DefaultShardString = "Shard"
