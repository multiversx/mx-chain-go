package mock

type RaterMock struct {
	ComputeRatingCalled      func(string, uint32) uint32
	GetRatingCalled          func(string) uint32
	GetRatingOptionKeyCalled func() []string
}

func (rm *RaterMock) ComputeRating(ratingOptionKey string, previousValue uint32) uint32 {
	if rm.ComputeRatingCalled != nil {
		return rm.ComputeRatingCalled(ratingOptionKey, previousValue)
	}
	return 1
}

func (rm *RaterMock) GetRating(pk string) uint32 {
	return rm.GetRatingCalled(pk)
}

func (rm *RaterMock) GetRatings([]string) map[string]uint32 {
	return make(map[string]uint32)
}

func (rm *RaterMock) IsInterfaceNil() bool {
	return false
}

func (rm *RaterMock) GetRatingOptionKeys() []string {
	if rm.GetRatingOptionKeyCalled != nil {
		return rm.GetRatingOptionKeyCalled()
	}
	return make([]string, 0)
}

func (rm *RaterMock) GetStartRating() uint32 {
	return 5
}
