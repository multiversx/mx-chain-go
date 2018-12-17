package mock

type RoundMock struct {
	RoundIndex int
}

func (rm *RoundMock) Index() int {
	return rm.RoundIndex
}
