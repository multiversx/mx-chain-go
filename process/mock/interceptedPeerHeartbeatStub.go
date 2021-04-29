package mock

// InterceptedPeerHeartbeatStub -
type InterceptedPeerHeartbeatStub struct {
	InterceptedDataStub
	PublicKeyCalled  func() []byte
	SetShardIDCalled func(shardId uint32)
}

// PublicKey -
func (i *InterceptedPeerHeartbeatStub) PublicKey() []byte {
	if i.PublicKeyCalled != nil {
		return i.PublicKeyCalled()
	}

	return make([]byte, 0)
}

// SetShardID -
func (i *InterceptedPeerHeartbeatStub) SetShardID(shardId uint32) {
	if i.SetShardIDCalled != nil {
		i.SetShardIDCalled(shardId)
	}
}
