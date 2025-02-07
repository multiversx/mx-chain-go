package block

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewShardHeaderValidatorFactory(t *testing.T) {
	t.Parallel()

	shvf := NewShardHeaderValidatorFactory()
	require.NotNil(t, shvf)
	require.Implements(t, new(HeaderValidatorCreator), shvf)
}

func TestShardHeaderValidatorFactory_CreateHeaderValidator(t *testing.T) {
	t.Parallel()

	shvf := NewShardHeaderValidatorFactory()

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

func TestShardHeaderValidatorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	shvf := NewShardHeaderValidatorFactory()
	require.False(t, shvf.IsInterfaceNil())
}
