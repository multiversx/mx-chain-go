package block

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewShardHeaderFactory_NilHeaderVersionHandlerShouldErr(t *testing.T) {
	t.Parallel()

	shf, err := NewShardHeaderFactory(nil)
	require.Nil(t, shf)
	require.True(t, check.IfNil(shf))
	require.Equal(t, ErrNilHeaderVersionHandler, err)
}

func TestNewShardHeaderFactory_OK(t *testing.T) {
	t.Parallel()

	shf, err := NewShardHeaderFactory(&testscommon.HeaderVersionHandlerStub{})
	require.Nil(t, err)
	require.False(t, check.IfNil(shf))
	require.NotNil(t, shf)
}

func TestNewShardHeaderFactory_CreateOK(t *testing.T) {
	t.Parallel()

	v1Version := "*"
	v2Version := "2"

	hvh := &testscommon.HeaderVersionHandlerStub{
		GetVersionCalled: func(epoch uint32) string {
			switch epoch {
			case 1:
				return v2Version
			}
			return v1Version
		},
	}

	shf, _ := NewShardHeaderFactory(hvh)

	epoch := uint32(0)
	header := shf.Create(epoch)
	require.NotNil(t, header)
	require.IsType(t, &block.Header{}, header)
	require.Equal(t, epoch, header.GetEpoch())
	require.Equal(t, []byte(v1Version), header.GetSoftwareVersion())

	epoch = uint32(1)
	header = shf.Create(epoch)
	require.NotNil(t, header)

	require.IsType(t, &block.HeaderV2{}, header)
	headerV2 := header.(*block.HeaderV2)
	require.NotNil(t, headerV2.Header)

	require.Equal(t, epoch, header.GetEpoch())
	require.Equal(t, []byte(v2Version), header.GetSoftwareVersion())
}
