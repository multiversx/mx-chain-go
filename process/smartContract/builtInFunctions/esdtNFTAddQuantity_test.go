package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestNewESDTNFTAddQuantityFunc(t *testing.T) {
	t.Parallel()

	// nil marshalizer
	eqf, err := NewESDTNFTAddQuantityFunc(10, nil, nil, nil)
	require.True(t, check.IfNil(eqf))
	require.Equal(t, process.ErrNilMarshalizer, err)

	// nil pause handler
	eqf, err = NewESDTNFTAddQuantityFunc(10, &mock.MarshalizerMock{}, nil, nil)
	require.True(t, check.IfNil(eqf))
	require.Equal(t, process.ErrNilPauseHandler, err)

	// nil roles handler
	eqf, err = NewESDTNFTAddQuantityFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, nil)
	require.True(t, check.IfNil(eqf))
	require.Equal(t, process.ErrNilRolesHandler, err)

	// should work
	eqf, err = NewESDTNFTAddQuantityFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})
	require.False(t, check.IfNil(eqf))
	require.NoError(t, err)
}

func TestEsdtNFTAddQuantity_SetNewGasConfig_NilGasCost(t *testing.T) {
	t.Parallel()

	defaultGasCost := uint64(10)
	eqf, _ := NewESDTNFTAddQuantityFunc(defaultGasCost, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	eqf.SetNewGasConfig(nil)
	require.Equal(t, defaultGasCost, eqf.funcGasCost)
}

func TestEsdtNFTAddQuantity_SetNewGasConfig_ShouldWork(t *testing.T) {
	t.Parallel()

	defaultGasCost := uint64(10)
	newGasCost := uint64(37)
	eqf, _ := NewESDTNFTAddQuantityFunc(defaultGasCost, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{}, &mock.ESDTRoleHandlerStub{})

	eqf.SetNewGasConfig(
		&process.GasCost{
			BuiltInCost: process.BuiltInCost{
				ESDTNFTAddQuantity: newGasCost,
			},
		},
	)

	require.Equal(t, newGasCost, eqf.funcGasCost)
}
