package mock

type ProposerResolverStub struct {
	ResolveProposerCalled func(shardId uint32, roundIndex uint32, prevRandomSeed []byte) ([]byte, error)
}

func (prs *ProposerResolverStub) ResolveProposer(shardId uint32, roundIndex uint32, prevRandomSeed []byte) ([]byte, error) {
	return prs.ResolveProposerCalled(shardId, roundIndex, prevRandomSeed)
}
