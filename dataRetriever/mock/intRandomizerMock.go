package mock

type IntRandomizerMock struct {
	IntnCalled func(n int) (int, error)
}

func (irm *IntRandomizerMock) Intn(n int) (int, error) {
	return irm.IntnCalled(n)
}
