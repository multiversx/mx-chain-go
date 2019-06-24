package external

// ProposerResolver defines how a struct will determine a block proposer
type ProposerResolver interface {
	ResolveProposer(shardId uint32, roundIndex uint32, prevRandomSeed []byte) ([]byte, error)
}

// ScDataGetter defines how data should be get from a SC account
type ScDataGetter interface {
	Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
}
