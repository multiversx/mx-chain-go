package block

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignHeaderValidatorFactory(t *testing.T) {
	t.Parallel()

	shvf, err := NewSovereignHeaderValidatorFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, shvf)

	sf, _ := NewShardHeaderValidatorFactory()
	shvf, err = NewSovereignHeaderValidatorFactory(sf)
	require.Nil(t, err)
	require.NotNil(t, shvf)
	require.Implements(t, new(HeaderValidatorCreator), shvf)
}

func TestSovereignHeaderValidatorFactory_CreateHeaderValidator(t *testing.T) {
	t.Parallel()

	sf, _ := NewShardHeaderValidatorFactory()
	shvf, _ := NewSovereignHeaderValidatorFactory(sf)

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
	require.Implements(t, new(process.HeaderConstructionValidator), hv)
}

func TestSovereignHeaderValidatorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sf, _ := NewShardHeaderValidatorFactory()
	shvf, _ := NewSovereignHeaderValidatorFactory(sf)
	require.False(t, shvf.IsInterfaceNil())
}
