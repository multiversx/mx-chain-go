package notifier

import coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"

type ProofTxData struct {
	Round     uint64
	ShardID   uint32
	Bytes     []byte
	SlashType coreSlash.SlashingType
}
