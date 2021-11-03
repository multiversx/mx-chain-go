package notifier

import coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"

// ProofTxDataExtractor defines a component which can extract necessary tx data from a given proof
// This data is meant to be used for a slashing notifier to be added in a transaction.data field
type ProofTxDataExtractor interface {
	GetProofTxData(proof coreSlash.SlashingProofHandler) (*ProofTxData, error)
	IsInterfaceNil() bool
}
