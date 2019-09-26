package factory

// InterceptedDataType represents a type of intercepted data instantiated by the Create func
type InterceptedDataType string

// InterceptedTx is the type for intercepted transaction
const InterceptedTx InterceptedDataType = "intercepted transaction"

// InterceptedShardHeader is the type for intercepted shard header
const InterceptedShardHeader InterceptedDataType = "intercepted shard header"

// InterceptedUnsignedTx is the type for intercepted unsigned transaction (smart contract results)
const InterceptedUnsignedTx InterceptedDataType = "intercepted unsigned transaction"
