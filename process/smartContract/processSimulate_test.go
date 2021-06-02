package smartContract

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSmartContractProcessorSimulate_newScProcessorErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = nil
	sc, err := NewSmartContractProcessorSimulate(arguments)

	require.Nil(t, sc)
	require.NotNil(t, err)
}

func TestNewSmartContractProcessorSimulate_shouldCheckValuesAlwaysFalse(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessorSimulate(arguments)

	require.False(t, sc.shouldCheckValues())
}
