package data

// ConsensusRewardData holds the required data for rewarding validators in a specific round and epoch
type ConsensusRewardData struct {
	Round     uint64
	Epoch     uint32
	PubKeys   []string
	Addresses []string
}
