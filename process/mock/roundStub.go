package mock

type RoundStub struct {
	IndexCalled func() int32
}

func (rndm *RoundStub) Index() int32 {
	return rndm.IndexCalled()
}
