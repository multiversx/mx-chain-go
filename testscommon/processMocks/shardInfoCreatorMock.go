package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ShardInfoCreatorMock is a mock implementation of ShardInfoCreator interface
type ShardInfoCreatorMock struct {
	CreateShardInfoV3Called             func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataHandler, error)
	CreateShardInfoFromLegacyMetaCalled func(metaHeader data.MetaHeaderHandler, shardHeaders []data.ShardHeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataHandler, error)
}

// CreateShardInfoV3 -
func (scm *ShardInfoCreatorMock) CreateShardInfoV3(
	metaHeader data.MetaHeaderHandler,
	shardHeaders []data.HeaderHandler,
	shardHeaderHashes [][]byte,
) ([]data.ShardDataHandler, error) {
	if scm.CreateShardInfoV3Called != nil {
		return scm.CreateShardInfoV3Called(metaHeader, shardHeaders, shardHeaderHashes)
	}
	return nil, nil
}

// CreateShardInfoFromLegacyMeta -
func (scm *ShardInfoCreatorMock) CreateShardInfoFromLegacyMeta(
	metaHeader data.MetaHeaderHandler,
	shardHeaders []data.ShardHeaderHandler,
	shardHeaderHashes [][]byte,
) ([]data.ShardDataHandler, error) {
	if scm.CreateShardInfoFromLegacyMetaCalled != nil {
		return scm.CreateShardInfoFromLegacyMetaCalled(metaHeader, shardHeaders, shardHeaderHashes)
	}
	return nil, nil
}

// IsInterfaceNil -
func (scm *ShardInfoCreatorMock) IsInterfaceNil() bool {
	return scm == nil
}
