package mock

type GasHandlerMock struct {
	InitGasConsumedCalled func()
	AddGasConsumedCalled  func(gasConsumed uint64)
	SetGasConsumedCalled  func(gasConsumed uint64)
	GetGasConsumedCalled  func() uint64
}

func (ghm *GasHandlerMock) InitGasConsumed() {
	ghm.InitGasConsumedCalled()
}

func (ghm *GasHandlerMock) AddGasConsumed(gasConsumed uint64) {
	ghm.AddGasConsumedCalled(gasConsumed)
}

func (ghm *GasHandlerMock) SetGasConsumed(gasConsumed uint64) {
	ghm.SetGasConsumedCalled(gasConsumed)
}

func (ghm *GasHandlerMock) GetGasConsumed() uint64 {
	return ghm.GetGasConsumedCalled()
}

func (ghm *GasHandlerMock) IsInterfaceNil() bool {
	if ghm == nil {
		return true
	}

	return false
}
