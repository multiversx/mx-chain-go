package processingOnlyNode

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgsTestOnlyProcessingNode represents the DTO struct for the NewTestOnlyProcessingNode constructor function
type ArgsTestOnlyProcessingNode struct {
	NumShards uint32
	ShardID   uint32
}

type testOnlyProcessingNode struct {
	Marshaller       coreData.Marshaller
	Hasher           coreData.Hasher
	ShardCoordinator sharding.Coordinator
}

// NewTestOnlyProcessingNode creates a new instance of a node that is able to only process transactions
func NewTestOnlyProcessingNode(args ArgsTestOnlyProcessingNode) (*testOnlyProcessingNode, error) {
	instance := &testOnlyProcessingNode{}

	err := instance.addBasicComponents(args)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (node *testOnlyProcessingNode) addBasicComponents(args ArgsTestOnlyProcessingNode) error {
	node.Marshaller = &marshal.GogoProtoMarshalizer{}
	node.Hasher = blake2b.NewBlake2b()

	var err error
	node.ShardCoordinator, err = sharding.NewMultiShardCoordinator(args.ShardID, args.NumShards)
	if err != nil {
		return err
	}

	return nil
}
