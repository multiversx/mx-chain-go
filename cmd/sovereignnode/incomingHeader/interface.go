package incomingHeader

import "github.com/multiversx/mx-chain-core-go/data"

// HeadersPool should be able to add new headers in pool
type HeadersPool interface {
	AddHeaderInShard(headerHash []byte, header data.HeaderHandler, shardID uint32)
	IsInterfaceNil() bool
}

// TransactionPool should be able to add a new transaction in the pool
type TransactionPool interface {
	AddData(key []byte, data interface{}, sizeInBytes int, cacheId string)
	IsInterfaceNil() bool
}

// TopicsChecker should be able to check the topics validity
type TopicsChecker interface {
	CheckValidity(topics [][]byte) error
	IsInterfaceNil() bool
}
