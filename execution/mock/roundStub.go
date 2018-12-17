package mock

type RoundStub struct {
	IndexCalled func() int
}

func (rndm *RoundStub) Index() int {
	return rndm.IndexCalled()
}
