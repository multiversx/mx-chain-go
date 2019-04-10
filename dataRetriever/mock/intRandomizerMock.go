package mock

type IntRandomizerMock struct {
	IntnCalled func(n int) int
}

func (irm *IntRandomizerMock) Intn(n int) int {
	return irm.IntnCalled(n)
}
