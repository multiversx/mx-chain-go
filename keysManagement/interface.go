package keysManagement

// NodesCoordinator provides Validator methods needed for the peer processing
type NodesCoordinator interface {
	GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	GetAllWaitingValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error)
	IsInterfaceNil() bool
}

// ShardProvider defines a component able to provide self shard
type ShardProvider interface {
	SelfId() uint32
	IsInterfaceNil() bool
}

// CurrentEpochProvider defines a component able to provide current epoch
type CurrentEpochProvider interface {
	CurrentEpoch() uint32
	IsInterfaceNil() bool
}
