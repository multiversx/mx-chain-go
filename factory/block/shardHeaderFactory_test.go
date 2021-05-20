package block

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewShardHeaderFactory_NilHeaderVersionHandlerShouldErr(t *testing.T) {
	shf, err := NewShardHeaderFactory(nil)
	require.Nil(t, shf)
	require.Equal(t, ErrNilHeaderVersionHandler, err)
}

func TestNewShardHeaderFactory_OK(t *testing.T) {
	shf, err := NewShardHeaderFactory(&testscommon.HeaderVersionHandlerMock{})
	require.Nil(t, err)
	require.NotNil(t, shf)
}

func TestNewShardHeaderFactory_CreateOK(t *testing.T) {
	hvh := &testscommon.HeaderVersionHandlerMock{
		GetVersionCalled: func(epoch uint32) string {
			switch epoch {
			case 1:
				return "2"
			}
			return "*"
		},
	}

	shf, _ := NewShardHeaderFactory(hvh)

	header := shf.Create(0)
	require.NotNil(t, header)
	require.IsType(t, &block.Header{}, header)

	header = shf.Create(1)
	require.NotNil(t, header)
	require.IsType(t, &block.HeaderV2{}, header)
}
