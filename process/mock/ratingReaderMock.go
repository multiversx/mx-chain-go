package mock

// RatingReaderMock -
type RatingReaderMock struct {
	GetRatingCalled                  func(string) uint32
	UpdateRatingFromTempRatingCalled func([]string) error
	RatingsMap                       map[string]uint32
}

// GetRating -
func (rrm *RatingReaderMock) GetRating(pk string) uint32 {
	if rrm.GetRatingCalled != nil {
		return rrm.GetRatingCalled(pk)
	}

	return 0
}

func (rrm *RatingReaderMock) UpdateRatingFromTempRating(pks []string) error {
	if rrm.UpdateRatingFromTempRatingCalled != nil {
		return rrm.UpdateRatingFromTempRatingCalled(pks)
	}
	return nil
}

// IsInterfaceNil -
func (rrm *RatingReaderMock) IsInterfaceNil() bool {
	return rrm == nil
}
