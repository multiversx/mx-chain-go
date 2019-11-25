package integrationTests

type blockProcessorInitializer struct {
	InitBlockProcessorCalled func()
}

func (bpi *blockProcessorInitializer) InitBlockProcessor() {
	bpi.InitBlockProcessorCalled()
}
