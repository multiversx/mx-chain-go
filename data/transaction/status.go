package transaction

// TransactionStatus is the type used to represent the status of a transaction
type TransactionStatus string

const (
	// TxStatusReceived represents the status of a transaction which was received but not yet executed
	TxStatusReceived TransactionStatus = "received"
	// TxStatusPartiallyExecuted represent the status of a transaction which was received and executed on source shard
	TxStatusPartiallyExecuted TransactionStatus = "partially-executed"
	// TxStatusExecuted represents the status of a transaction which was received and executed
	TxStatusExecuted TransactionStatus = "executed"
	// TxStatusNotExecuted represents the status of a transaction which was received and not executed
	TxStatusNotExecuted TransactionStatus = "not-executed"
	// TxStatusInvalid represents the status of a transaction which was considered invalid
	TxStatusInvalid TransactionStatus = "invalid"
)
