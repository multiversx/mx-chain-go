package sovereign_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/sovereign"
	sovereignMock "github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

func TestNewSovereignTopicsCheckerFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil topics checker should fail", func(t *testing.T) {
		sovereignTopicsCheckerFactory, err := sovereign.NewSovereignTopicsCheckerFactory(nil)
		require.Equal(t, errors.ErrNilTopicsChecker, err)
		require.True(t, sovereignTopicsCheckerFactory.IsInterfaceNil())
	})

	t.Run("should work", func(t *testing.T) {
		sovereignTopicsCheckerFactory, err := sovereign.NewSovereignTopicsCheckerFactory(&sovereignMock.TopicsCheckerMock{})
		require.Nil(t, err)
		require.False(t, sovereignTopicsCheckerFactory.IsInterfaceNil())
	})
}

func TestSovereignTopicsCheckerFactory_CreateTopicsChecker(t *testing.T) {
	t.Parallel()

	sovereignTopicsCheckerFactory, _ := sovereign.NewSovereignTopicsCheckerFactory(&sovereignMock.TopicsCheckerMock{})
	sovereignTopicsChecker := sovereignTopicsCheckerFactory.CreateTopicsChecker()
	require.NotNil(t, sovereignTopicsChecker)
	require.Equal(t, "*sovereign.TopicsCheckerMock", fmt.Sprintf("%T", sovereignTopicsChecker))
}
