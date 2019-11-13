package mock

type RaterMock struct {
	UpdateRatingCalled func()
	GetRatingCalled    func(string) int64
}

func (rm *RaterMock) UpdateRatings(updateValues map[string][]string) {
	if rm.UpdateRatingCalled != nil {
		rm.UpdateRatingCalled()
	}
}

func (rm *RaterMock) GetRating(pk string) int64 {
	return rm.GetRatingCalled(pk)
}

func (rm *RaterMock) GetRatings() map[string]int64 {
	return make(map[string]int64)
}
