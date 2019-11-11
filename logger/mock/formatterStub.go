package mock

import "github.com/ElrondNetwork/elrond-go/logger"

type FormatterStub struct {
	OutputCalled func(line logger.LogLineHandler) []byte
}

func (fs *FormatterStub) Output(line logger.LogLineHandler) []byte {
	return fs.OutputCalled(line)
}

func (fs *FormatterStub) IsInterfaceNil() bool {
	return fs == nil
}
