package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRatingReader_GetRatingShouldWork(t *testing.T) {
	rr := &RatingReader{}

	testKey := "teskKey"
	testRatingValue := uint32(4)
	getRatingCalled := false
	rr.getRating = func(s string) uint32 {
		if s == testKey {
			getRatingCalled = true
			return testRatingValue
		}
		return 0
	}

	rating := rr.GetRating(testKey)
	assert.True(t, getRatingCalled)
	assert.Equal(t, testRatingValue, rating)
}

func TestRatingReader_IsInterfaceNil(t *testing.T) {
	rr := &RatingReader{}

	assert.False(t, rr.IsInterfaceNil())
}
