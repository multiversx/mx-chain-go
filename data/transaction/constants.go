package transaction

// TxType represents a transaction type
type TxType string

const (
	// TxTypeNormal represents the identifier for a regular transaction
	TxTypeNormal TxType = "normal"

	// TxTypeUnsigned represents the identifier for a unsigned transaction
	TxTypeUnsigned TxType = "unsigned"

	// TxTypeReward represents the identifier for a reward transaction
	TxTypeReward TxType = "reward"

	// TxTypeInvalid represents the identifier for an invalid transaction
	TxTypeInvalid TxType = "invalid"
)
