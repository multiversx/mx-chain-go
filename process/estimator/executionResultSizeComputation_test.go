package estimator_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/require"
)

const maxSizeInBytes = uint32(core.MegabyteSize * 10 / 100)

func TestNewExecResultSizeComputation(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		ers, err := estimator.NewExecResultSizeComputationHandler(nil, maxSizeInBytes)
		require.Nil(t, ers)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ers, err := estimator.NewExecResultSizeComputationHandler(&mock.ProtobufMarshalizerMock{}, maxSizeInBytes)
		require.Nil(t, err)
		require.NotNil(t, ers)
	})
}

func TestExecResultSizeComputation_IsMaxExecResSizeReached(t *testing.T) {
	t.Parallel()

	ers, _ := estimator.NewExecResultSizeComputationHandler(
		&mock.ProtobufMarshalizerMock{},
		maxSizeInBytes,
	)

	ersl := ers.NewComputation()

	testData := []struct {
		numNewExecRes int
		expected      bool
		name          string
	}{
		{numNewExecRes: 0, expected: false, name: "with numExecRes 0"},
		{numNewExecRes: 10, expected: false, name: "with numExecRes 10"},
		{numNewExecRes: 30, expected: false, name: "with numExecRes 30"},
		{numNewExecRes: 150, expected: false, name: "with numExecRes 150"},
		{numNewExecRes: 200, expected: true, name: "with numExecRes 200"},
		{numNewExecRes: 1000, expected: true, name: "with numExecRes 1000"},
		{numNewExecRes: 10000, expected: true, name: "with numExecRes 10000"},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			require.Equal(t, td.expected, ersl.IsMaxExecResSizeReached(td.numNewExecRes))
		})
	}
}
