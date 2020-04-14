package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/require"
)

func TestNewLatestDataProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	ldp, err := NewLatestDataProvider(getLatestDataProviderArgs())
	require.False(t, check.IfNil(ldp))
	require.NoError(t, err)
}

func getLatestDataProviderArgs() ArgsLatestDataProvider {
	return ArgsLatestDataProvider{
		GeneralConfig:         config.Config{},
		Marshalizer:           &mock.MarshalizerMock{},
		Hasher:                &mock.HasherMock{},
		BootstrapDataProvider: &mock.BootStrapDataProviderStub{},
		DirectoryReader:       &mock.DirectoryReaderStub{},
		WorkingDir:            "",
		ChainID:               "",
		DefaultDBPath:         "",
		DefaultEpochString:    "",
		DefaultShardString:    "",
	}
}
