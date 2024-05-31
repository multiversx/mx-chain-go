package rating

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignRatingsDataFactory(t *testing.T) {
	factory := NewSovereignRatingsDataFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(RatingsDataFactory), factory)
}

func TestSovereignRatingsDataFactory_CreateRatingsData(t *testing.T) {
	factory := NewSovereignRatingsDataFactory()

	ratingsDataArg := createSovereignRatingsDataArgs()
	blockCreator, err := factory.CreateRatingsData(ratingsDataArg)
	require.Nil(t, err)
	require.IsType(t, &sovereignRatingsData{}, blockCreator)
}
