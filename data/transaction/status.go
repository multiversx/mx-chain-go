package transaction

// TxStatus is the status of a transaction
type TxStatus string

const (
	// TxStatusReceived = transaction was received but not yet executed
	TxStatusReceived TxStatus = "received"
	// TxStatusPartiallyExecuted represent the status of a transaction which was received and executed on source shard
	TxStatusPartiallyExecuted TxStatus = "partially-executed"
	// TxStatusExecuted represents the status of a transaction which was received and executed
	TxStatusExecuted TxStatus = "executed"
	// TxStatusNotExecuted represents the status of a transaction which was received and not executed
	TxStatusNotExecuted TxStatus = "not-executed"
	// TxStatusInvalid represents the status of a transaction which was considered invalid
	TxStatusInvalid TxStatus = "invalid"
)
