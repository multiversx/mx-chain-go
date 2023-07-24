package block

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewShardHeaderValidatorFactory(t *testing.T) {
	shvf, err := NewShardHeaderValidatorFactory()
	require.Nil(t, err)
	require.NotNil(t, shvf)
}

func TestShardHeaderValidatorFactory_CreateHeaderValidator(t *testing.T) {
	shvf, _ := NewShardHeaderValidatorFactory()

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

func TestShardHeaderValidatorFactory_IsInterfaceNil(t *testing.T) {
	shvf, _ := NewShardHeaderValidatorFactory()
	require.False(t, shvf.IsInterfaceNil())
}
