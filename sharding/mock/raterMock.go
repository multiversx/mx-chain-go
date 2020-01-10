package mock

type RaterMock struct {
	ComputeRatingCalled              func(string, uint32) uint32
	GetRatingCalled                  func(string) uint32
	UpdateRatingFromTempRatingCalled func(string)
	GetStartRatingCalled             func() uint32
	GetChancesCalled                 func(uint32) uint32
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

func (rm *RaterMock) UpdateRatingFromTempRating(pk string) {
	if rm.UpdateRatingFromTempRatingCalled != nil {
		rm.UpdateRatingFromTempRatingCalled(pk)
	}
}

func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}

func (rm *RaterMock) GetStartRating() uint32 {
	if rm.GetStartRatingCalled != nil {
		return rm.GetStartRatingCalled()
	}
	return 5
}

func (rm *RaterMock) GetChance(rating uint32) uint32 {
	if rm.GetChancesCalled != nil {
		return rm.GetChancesCalled(rating)
	}
	return 5
}
