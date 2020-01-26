package txpool

// Economics interface contains economics-related functions required by the txpool
type Economics interface {
	MinGasPrice() uint64
	IsInterfaceNil() bool
}

// Sharding interface contains shards-related functions required by the txpool
type Sharding interface {
	NumberOfShards() uint32
	IsInterfaceNil() bool
}
