package block

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMetaHeaderFactory_NilHeaderVersionHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mhf, err := NewMetaHeaderFactory(nil)
	require.Nil(t, mhf)
	require.True(t, check.IfNil(mhf))
	require.Equal(t, ErrNilHeaderVersionHandler, err)
}

func TestNewMetaHeaderFactory_OK(t *testing.T) {
	t.Parallel()

	mhf, err := NewMetaHeaderFactory(&testscommon.HeaderVersionHandlerStub{})
	require.Nil(t, err)
	require.False(t, check.IfNil(mhf))
	require.NotNil(t, mhf)
}

func TestNewMetaHeaderFactory_CreateOK(t *testing.T) {
	t.Parallel()

	hvh := &testscommon.HeaderVersionHandlerStub{
		GetVersionCalled: func(epoch uint32) string {
			switch epoch {
			case 1:
				return "2"
			}
			return "*"
		},
	}

	mhf, _ := NewMetaHeaderFactory(hvh)

	epoch := uint32(0)
	header := mhf.Create(epoch)
	require.NotNil(t, header)
	require.IsType(t, &block.MetaBlock{}, header)
	require.Equal(t, epoch, header.GetEpoch())

	epoch = uint32(1)
	header = mhf.Create(epoch)
	require.NotNil(t, header)
	require.IsType(t, &block.MetaBlock{}, header)
	require.Equal(t, epoch, header.GetEpoch())
}
