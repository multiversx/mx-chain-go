package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/stretchr/testify/assert"
)

func getArgs() TrieFactoryArgs {
	return TrieFactoryArgs{
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &mock.HasherMock{},
		PathManager: &mock.PathManagerStub{},
		ShardId:     "0",
		Cfg: config.StorageConfig{
			Cache: config.CacheConfig{
				Type:   "LRU",
				Size:   100,
				Shards: 2,
			},
			DB: config.DBConfig{
				FilePath:          "path",
				Type:              "LvlDBSerial",
				BatchDelaySeconds: 10,
				MaxBatchSize:      200,
				MaxOpenFiles:      10,
			},
			Bloom: config.BloomFilterConfig{},
		},
	}
}

func TestNewTrieFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Marshalizer = nil
	tf, err := NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Hasher = nil
	tf, err := NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.PathManager = nil
	tf, err := NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilPathManager, err)
}
