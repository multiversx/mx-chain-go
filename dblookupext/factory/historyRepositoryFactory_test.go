package factory_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext/factory"
	"github.com/multiversx/mx-chain-go/process"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNewHistoryRepositoryFactory(t *testing.T) {
	args := getArgs()

	argsNilStorageService := getArgs()
	argsNilStorageService.Store = nil
	hrf, err := factory.NewHistoryRepositoryFactory(argsNilStorageService)
	require.Equal(t, core.ErrNilStore, err)
	require.Nil(t, hrf)

	argsNilMarshalizer := getArgs()
	argsNilMarshalizer.Marshalizer = nil
	hrf, err = factory.NewHistoryRepositoryFactory(argsNilMarshalizer)
	require.Equal(t, core.ErrNilMarshalizer, err)
	require.Nil(t, hrf)

	argsNilHasher := getArgs()
	argsNilHasher.Hasher = nil
	hrf, err = factory.NewHistoryRepositoryFactory(argsNilHasher)
	require.Equal(t, core.ErrNilHasher, err)
	require.Nil(t, hrf)

	argsNilUint64Converter := getArgs()
	argsNilUint64Converter.Uint64ByteSliceConverter = nil
	hrf, err = factory.NewHistoryRepositoryFactory(argsNilUint64Converter)
	require.Equal(t, process.ErrNilUint64Converter, err)
	require.Nil(t, hrf)

	hrf, err = factory.NewHistoryRepositoryFactory(args)
	require.NoError(t, err)
	require.False(t, check.IfNil(hrf))
}

func TestHistoryRepositoryFactory_CreateShouldCreateDisabledRepository(t *testing.T) {
	hrf, _ := factory.NewHistoryRepositoryFactory(getArgs())

	repository, err := hrf.Create()
	require.NoError(t, err)
	require.NotNil(t, repository)
	require.False(t, repository.IsEnabled())
}

func TestHistoryRepositoryFactory_CreateShouldCreateRegularRepository(t *testing.T) {
	args := getArgs()
	args.Config.Enabled = true
	args.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}

	hrf, _ := factory.NewHistoryRepositoryFactory(args)

	repository, err := hrf.Create()
	require.NoError(t, err)
	require.NotNil(t, repository)
	require.True(t, repository.IsEnabled())
}

func TestHistoryRepositoryFactory_CreateMissingStorersReturnsError(t *testing.T) {
	t.Parallel()

	t.Run("missing ESDTSuppliesUnit", testWithMissingStorer(dataRetriever.ESDTSuppliesUnit))
	t.Run("missing TxLogsUnit", testWithMissingStorer(dataRetriever.TxLogsUnit))
	t.Run("missing RoundHdrHashDataUnit", testWithMissingStorer(dataRetriever.RoundHdrHashDataUnit))
	t.Run("missing MiniblocksMetadataUnit", testWithMissingStorer(dataRetriever.MiniblocksMetadataUnit))
	t.Run("missing EpochByHashUnit", testWithMissingStorer(dataRetriever.EpochByHashUnit))
	t.Run("missing MiniblockHashByTxHashUnit", testWithMissingStorer(dataRetriever.MiniblockHashByTxHashUnit))
	t.Run("missing ResultsHashesByTxHashUnit", testWithMissingStorer(dataRetriever.ResultsHashesByTxHashUnit))
}

func testWithMissingStorer(missingUnit dataRetriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := getArgs()
		args.Config.Enabled = true
		args.Store = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == missingUnit {
					return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
				}
				return &storageStubs.StorerStub{}, nil
			},
		}
		hrf, _ := factory.NewHistoryRepositoryFactory(args)
		repository, err := hrf.Create()
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
		require.True(t, strings.Contains(err.Error(), missingUnit.String()))
		require.True(t, check.IfNil(repository))
	}
}

func getArgs() *factory.ArgsHistoryRepositoryFactory {
	return &factory.ArgsHistoryRepositoryFactory{
		SelfShardID:              0,
		Config:                   config.DbLookupExtensionsConfig{},
		Store:                    &storageStubs.ChainStorerStub{},
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &hashingMocks.HasherMock{},
		Uint64ByteSliceConverter: &processMock.Uint64ByteSliceConverterMock{},
	}
}
