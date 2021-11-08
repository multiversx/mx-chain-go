package notifier

import coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"

// ProofTxData represents necessary data to be used in a slashing commitment proof tx by a slashing notifier.
// Each field is required to be added in a transaction.data field
type ProofTxData struct {
	Round     uint64
	ShardID   uint32
	Bytes     []byte
	SlashType coreSlash.SlashingType
}
