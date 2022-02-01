package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common/mock"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	processMock "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
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
	args.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &storageStubs.StorerStub{}
		},
	}

	hrf, _ := factory.NewHistoryRepositoryFactory(args)

	repository, err := hrf.Create()
	require.NoError(t, err)
	require.NotNil(t, repository)
	require.True(t, repository.IsEnabled())
}

func getArgs() *factory.ArgsHistoryRepositoryFactory {
	return &factory.ArgsHistoryRepositoryFactory{
		SelfShardID:              0,
		Config:                   config.DbLookupExtensionsConfig{},
		Store:                    &mock.ChainStorerMock{},
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &hashingMocks.HasherMock{},
		Uint64ByteSliceConverter: &processMock.Uint64ByteSliceConverterMock{},
	}
}
