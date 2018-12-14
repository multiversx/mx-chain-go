package mock

type RoundMock struct {
	IndexCalled func() int
}

func (rndm *RoundMock) Index() int {
	return rndm.IndexCalled()
}
