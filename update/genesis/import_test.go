package genesis

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO increase code coverage

func createArgsNewStateImport() ArgsNewStateImport {
	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = &storageManager.StorageManagerStub{}
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = &storageManager.StorageManagerStub{}
	return ArgsNewStateImport{
		HardforkStorer:      &mock.HardforkStorerStub{},
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &mock.HasherStub{},
		TrieStorageManagers: trieStorageManagers,
		AddressConverter:    &testscommon.PubkeyConverterMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		AccountCreator:      &stateMock.AccountsFactoryStub{},
	}
}

func TestNewStateImport(t *testing.T) {

	tests := []struct {
		name    string
		args    func() ArgsNewStateImport
		exError error
	}{
		{
			name: "NilHarforkStorer",
			args: func() ArgsNewStateImport {
				args := createArgsNewStateImport()
				args.HardforkStorer = nil
				return args
			},
			exError: update.ErrNilHardforkStorer,
		},
		{
			name: "NilMarshalizer",
			args: func() ArgsNewStateImport {
				args := createArgsNewStateImport()
				args.Marshalizer = nil
				return args
			},
			exError: update.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			args: func() ArgsNewStateImport {
				args := createArgsNewStateImport()
				args.Hasher = nil
				return args
			},
			exError: update.ErrNilHasher,
		},
		{
			name: "NilAccountCreator",
			args: func() ArgsNewStateImport {
				args := createArgsNewStateImport()
				args.AccountCreator = nil
				return args
			},
			exError: state.ErrNilAccountFactory,
		},
		{
			name: "Ok",
			args: func() ArgsNewStateImport {
				return createArgsNewStateImport()
			},
			exError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStateImport(tt.args())
			require.Equal(t, tt.exError, err)
		})
	}
}

func TestImportAll(t *testing.T) {
	t.Parallel()

	args := createArgsNewStateImport()
	importState, _ := NewStateImport(args)
	require.False(t, check.IfNil(importState))

	err := importState.ImportAll()
	require.Nil(t, err)
}

func TestStateImport_ImportUnFinishedMetaBlocksShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsNewStateImport()
	metaBlock := &block.MetaBlock{
		Round:   1,
		ChainID: []byte("chainId"),
	}
	metaBlockHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, metaBlock)

	args.HardforkStorer = &mock.HardforkStorerStub{
		GetCalled: func(identifier string, key []byte) ([]byte, error) {
			return args.Marshalizer.Marshal(metaBlock)
		},
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
