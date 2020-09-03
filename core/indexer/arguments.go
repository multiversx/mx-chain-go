package indexer

import (
	"io"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

//DataIndexerArgs is struct that is used to store all components that are needed to create a indexer
type DataIndexerArgs struct {
	ShardID                  uint32
	Url                      string
	UserName                 string
	Password                 string
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	EpochStartNotifier       sharding.EpochStartEventNotifier
	NodesCoordinator         sharding.NodesCoordinator
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	IndexTemplates           map[string]io.Reader
	IndexPolicies            map[string]io.Reader
	Options                  *Options
}
