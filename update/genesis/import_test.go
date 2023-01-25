package genesis

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/trie/factory"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO increase code coverage

func TestNewStateImport(t *testing.T) {
	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = &testscommon.StorageManagerStub{}
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
				AddressConverter:    &testscommon.PubkeyConverterMock{},
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
				AddressConverter:    &testscommon.PubkeyConverterMock{},
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
				AddressConverter:    &testscommon.PubkeyConverterMock{},
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
				AddressConverter:    &testscommon.PubkeyConverterMock{},
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

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = &testscommon.StorageManagerStub{}
	trieStorageManagers[factory.PeerAccountTrie] = &testscommon.StorageManagerStub{}

	args := ArgsNewStateImport{
		HardforkStorer:      &mock.HardforkStorerStub{},
		Hasher:              &hashingMocks.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		TrieStorageManagers: trieStorageManagers,
		ShardID:             0,
		StorageConfig:       config.StorageConfig{},
		AddressConverter:    &testscommon.PubkeyConverterMock{},
	}

	importState, _ := NewStateImport(args)
	require.False(t, check.IfNil(importState))

	err := importState.ImportAll()
	require.Nil(t, err)
}

func TestStateImport_ImportUnFinishedMetaBlocksShouldWork(t *testing.T) {
	t.Parallel()

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = &testscommon.StorageManagerStub{}

	hasher := &hashingMocks.HasherMock{}
	marshahlizer := &mock.MarshalizerMock{}
	metaBlock := &block.MetaBlock{
		Round:   1,
		ChainID: []byte("chainId"),
	}
	metaBlockHash, _ := core.CalculateHash(marshahlizer, hasher, metaBlock)

	args := ArgsNewStateImport{
		HardforkStorer: &mock.HardforkStorerStub{
			GetCalled: func(identifier string, key []byte) ([]byte, error) {
				return marshahlizer.Marshal(metaBlock)
			},
		},
		Hasher:              hasher,
		Marshalizer:         marshahlizer,
		TrieStorageManagers: trieStorageManagers,
		ShardID:             0,
		StorageConfig:       config.StorageConfig{},
		AddressConverter:    &testscommon.PubkeyConverterMock{},
	}

	importState, _ := NewStateImport(args)
	require.False(t, check.IfNil(importState))

	key := fmt.Sprintf("meta@chainId@%s", hex.EncodeToString(metaBlockHash))
	err := importState.importUnFinishedMetaBlocks(UnFinishedMetaBlocksIdentifier, [][]byte{
		[]byte(key),
	})

	require.Nil(t, err)
	assert.Equal(t, importState.importedUnFinishedMetaBlocks[string(metaBlockHash)], metaBlock)
}
