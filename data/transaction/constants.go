package transaction

// TxType represents a transaction type
type TxType string

const (
	TxTypeNormal   TxType = "normal"
	TxTypeUnsigned TxType = "unsigned"
	TxTypeReward   TxType = "reward"
	TxTypeInvalid  TxType = "invalid"
)
