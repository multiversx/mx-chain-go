package rating

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRatingsDataFactory(t *testing.T) {
	factory := NewRatingsDataFactory()
	require.False(t, factory.IsInterfaceNil())
	require.Implements(t, new(RatingsDataFactory), factory)
}

func TestRatingsDataFactory_CreateRatingsData(t *testing.T) {
	factory := NewRatingsDataFactory()

	ratingsDataArg := createDymmyRatingsData()
	ratingsDataArg.Config = createDummyRatingsConfig()
	blockCreator, err := factory.CreateRatingsData(ratingsDataArg)
	require.Nil(t, err)
	require.IsType(t, &RatingsData{}, blockCreator)
}
