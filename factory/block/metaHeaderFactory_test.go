package block

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMetaHeaderFactory_NilHeaderVersionHandlerShouldErr(t *testing.T) {
	mhf, err := NewMetaHeaderFactory(nil)
	require.Nil(t, mhf)
	require.Equal(t, ErrNilHeaderVersionHandler, err)
}

func TestNewMetaHeaderFactory_OK(t *testing.T) {
	mhf, err := NewMetaHeaderFactory(&testscommon.HeaderVersionHandlerMock{})
	require.Nil(t, err)
	require.NotNil(t, mhf)
}

func TestNewMetaHeaderFactory_CreateOK(t *testing.T) {
	hvh := &testscommon.HeaderVersionHandlerMock{
		GetVersionCalled: func(epoch uint32) string {
			switch epoch {
			case 1:
				return "2"
			}
			return "*"
		},
	}

	mhf, _ := NewMetaHeaderFactory(hvh)

	header := mhf.Create(0)
	require.NotNil(t, header)
	require.IsType(t, &block.MetaBlock{}, header)

	header = mhf.Create(1)
	require.NotNil(t, header)
	require.IsType(t, &block.MetaBlock{}, header)
}
