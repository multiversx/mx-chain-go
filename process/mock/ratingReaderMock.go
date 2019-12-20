package mock

type RatingReaderMock struct {
	GetRatingCalled  func(string) uint32
	GetRatingsCalled func([]string) map[string]uint32
	RatingsMap       map[string]uint32
}

func (rrm *RatingReaderMock) GetRating(pk string) uint32 {
	if rrm.GetRatingCalled != nil {
		return rrm.GetRatingCalled(pk)
	}

	return 0
}

func (rrm *RatingReaderMock) GetRatings(pks []string) map[string]uint32 {
	if rrm.GetRatingsCalled != nil {
		return rrm.GetRatingsCalled(pks)
	}

	return map[string]uint32{}
}

func (rrm *RatingReaderMock) IsInterfaceNil() bool {
	return rrm == nil
}
