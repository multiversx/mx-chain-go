package external

// ProposerResolver defines how a struct will determine a block proposer
type ProposerResolver interface {
	ResolveProposer(shardId uint32, roundIndex uint32, prevRandomSeed []byte) ([]byte, error)
}
