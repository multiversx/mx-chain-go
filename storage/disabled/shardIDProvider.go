package disabled

type shardIDProvider struct{}

// NewShardIDProvider returns a new disabled shard id provider instance
func NewShardIDProvider() *shardIDProvider {
	return &shardIDProvider{}
}

// ComputeId returns 0
func (s *shardIDProvider) ComputeId(key []byte) uint32 {
	return 0
}

// NumberOfShards returns 0
func (s *shardIDProvider) NumberOfShards() uint32 {
	return 0
}

// GetShardIDs returns nil
func (s *shardIDProvider) GetShardIDs() []uint32 {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *shardIDProvider) IsInterfaceNil() bool {
	return s == nil
}
