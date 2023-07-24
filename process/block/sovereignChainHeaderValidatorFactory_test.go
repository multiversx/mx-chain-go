package block

import (
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSovereignHeaderValidatorFactory(t *testing.T) {
	shvf, err := NewSovereignHeaderValidatorFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, shvf)

	sf, _ := NewShardHeaderValidatorFactory()
	shvf, err = NewSovereignHeaderValidatorFactory(sf)
	require.Nil(t, err)
	require.NotNil(t, shvf)
}

func TestSovereignHeaderValidatorFactory_CreateHeaderValidator(t *testing.T) {
	sf, _ := NewShardHeaderValidatorFactory()
	shvf, err := NewSovereignHeaderValidatorFactory(sf)

	hv, err := shvf.CreateHeaderValidator(ArgsHeaderValidator{
		Hasher:      nil,
		Marshalizer: nil,
	})
	require.Nil(t, hv)
	require.NotNil(t, err)

	hv, err = shvf.CreateHeaderValidator(ArgsHeaderValidator{
		Hasher:      &testscommon.HasherStub{},
		Marshalizer: &testscommon.ProtoMarshalizerMock{},
	})
	require.NotNil(t, hv)
	require.Nil(t, err)
}

func TestSovereignHeaderValidatorFactory_IsInterfaceNil(t *testing.T) {
	sf, _ := NewShardHeaderValidatorFactory()
	shvf, _ := NewSovereignHeaderValidatorFactory(sf)
	require.False(t, shvf.IsInterfaceNil())
}
