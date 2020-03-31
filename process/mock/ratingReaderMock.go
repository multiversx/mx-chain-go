package mock

// RatingReaderMock -
type RatingReaderMock struct {
	GetRatingCalled func(string) uint32
	RatingsMap      map[string]uint32
}

// GetRating -
func (rrm *RatingReaderMock) GetRating(pk string) uint32 {
	if rrm.GetRatingCalled != nil {
		return rrm.GetRatingCalled(pk)
	}

	return 0
}

// IsInterfaceNil -
func (rrm *RatingReaderMock) IsInterfaceNil() bool {
	return rrm == nil
}
