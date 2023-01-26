package economics_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/require"
)

func TestNewBuiltInFunctionsCost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  func() *economics.ArgsBuiltInFunctionCost
		exErr error
	}{
		{
			name: "NilArguments",
			args: func() *economics.ArgsBuiltInFunctionCost {
				return nil
			},
			exErr: process.ErrNilArgsBuiltInFunctionsConstHandler,
		},
		{
			name: "NilArgumentsParser",
			args: func() *economics.ArgsBuiltInFunctionCost {
				return &economics.ArgsBuiltInFunctionCost{
					ArgsParser:  nil,
					GasSchedule: testscommon.NewGasScheduleNotifierMock(nil),
				}
			},
			exErr: process.ErrNilArgumentParser,
		},
		{
			name: "NilGasScheduleHandler",
			args: func() *economics.ArgsBuiltInFunctionCost {
				return &economics.ArgsBuiltInFunctionCost{
					ArgsParser:  &mock.ArgumentParserMock{},
					GasSchedule: nil,
				}
			},
			exErr: process.ErrNilGasSchedule,
		},
		{
			name: "ShouldWork",
			args: func() *economics.ArgsBuiltInFunctionCost {
				return &economics.ArgsBuiltInFunctionCost{
					ArgsParser:  &mock.ArgumentParserMock{},
					GasSchedule: testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
				}
			},
			exErr: nil,
		},
	}

	for _, test := range tests {
		_, err := economics.NewBuiltInFunctionsCost(test.args())
		require.Equal(t, test.exErr, err)
	}
}

func TestNewBuiltInFunctionsCost_GasConfig(t *testing.T) {
	t.Parallel()

	args := &economics.ArgsBuiltInFunctionCost{
		ArgsParser:  &mock.ArgumentParserMock{},
		GasSchedule: testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 0)),
	}

	builtInCostHandler, err := economics.NewBuiltInFunctionsCost(args)
	require.NotNil(t, err)
	require.Nil(t, builtInCostHandler)
	require.True(t, check.IfNil(builtInCostHandler))
}
