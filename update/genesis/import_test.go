package genesis

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

//TODO increase code coverage

func TestNewStateImport(t *testing.T) {
	trieStorageManagers := make(map[string]data.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = &mock.StorageManagerStub{}
	tests := []struct {
		name    string
		args    ArgsNewStateImport
		exError error
	}{
		{
			name: "NilHarforkStorer",
			args: ArgsNewStateImport{
				HardforkStorer:      nil,
				Marshalizer:         &mock.MarshalizerMock{},
				Hasher:              &mock.HasherStub{},
				TrieStorageManagers: trieStorageManagers,
			},
			exError: update.ErrNilHardforkStorer,
		},
		{
			name: "NilMarshalizer",
			args: ArgsNewStateImport{
				HardforkStorer:      &mock.HardforkStorerStub{},
				Marshalizer:         nil,
				Hasher:              &mock.HasherStub{},
				TrieStorageManagers: trieStorageManagers,
			},
			exError: update.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			args: ArgsNewStateImport{
				HardforkStorer:      &mock.HardforkStorerStub{},
				Marshalizer:         &mock.MarshalizerMock{},
				Hasher:              nil,
				TrieStorageManagers: trieStorageManagers,
			},
			exError: update.ErrNilHasher,
		},
		{
			name: "Ok",
			args: ArgsNewStateImport{
				HardforkStorer:      &mock.HardforkStorerStub{},
				Marshalizer:         &mock.MarshalizerMock{},
				Hasher:              &mock.HasherStub{},
				TrieStorageManagers: trieStorageManagers,
			},
			exError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStateImport(tt.args)
			require.Equal(t, tt.exError, err)
		})
	}
}

func TestImportAll(t *testing.T) {
	t.Parallel()

	trieStorageManagers := make(map[string]data.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = &mock.StorageManagerStub{}
	trieStorageManagers[factory.PeerAccountTrie] = &mock.StorageManagerStub{}

	args := ArgsNewStateImport{
		HardforkStorer:      &mock.HardforkStorerStub{},
		Hasher:              &mock.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		TrieStorageManagers: trieStorageManagers,
		ShardID:             0,
		StorageConfig:       config.StorageConfig{},
	}

	importState, _ := NewStateImport(args)
	require.False(t, check.IfNil(importState))

	err := importState.ImportAll()
	require.Nil(t, err)
}
