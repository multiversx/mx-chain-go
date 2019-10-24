package factory

// InterceptedDataType represents a type of intercepted data instantiated by the Create func
type InterceptedDataType string

// InterceptedTx is the type for intercepted transaction
const InterceptedTx InterceptedDataType = "intercepted transaction"

// InterceptedUnsignedTx is the type for intercepted unsigned transaction (smart contract results)
const InterceptedUnsignedTx InterceptedDataType = "intercepted unsigned transaction"

// InterceptedShardHeader is the type for intercepted shard header
const InterceptedShardHeader InterceptedDataType = "intercepted shard header"

// InterceptedMetaHeader is the type for intercepted meta header
const InterceptedMetaHeader InterceptedDataType = "intercepted meta header"

// InterceptedTxBlockBody is the type for intercepted tx block body
const InterceptedTxBlockBody InterceptedDataType = "intercepted block body"
