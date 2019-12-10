package mock

type RaterMock struct {
	ComputeRatingCalled  func(string, uint32) uint32
	GetRatingCalled      func(string) uint32
	GetRatingsCalled     func([]string) map[string]uint32
	GetStartRatingCalled func() uint32
}

func (rm *RaterMock) ComputeRating(ratingOptionKey string, previousValue uint32) uint32 {
	if rm.ComputeRatingCalled != nil {
		return rm.ComputeRatingCalled(ratingOptionKey, previousValue)
	}
	return 1
}

func (rm *RaterMock) GetRating(pk string) uint32 {
	if rm.GetRatingCalled != nil {
		return rm.GetRatingCalled(pk)
	}
	return 1
}

func (rm *RaterMock) GetRatings(pks []string) map[string]uint32 {
	if rm.GetRatingsCalled == nil {
		return rm.GetRatingsCalled(pks)
	}
	return make(map[string]uint32)
}

func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}

func (rm *RaterMock) GetStartRating() uint32 {
	if rm.GetStartRatingCalled == nil {
		return rm.GetStartRatingCalled()
	}
	return 5
}
