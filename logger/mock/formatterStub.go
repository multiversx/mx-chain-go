package mock

import "github.com/ElrondNetwork/elrond-go/logger"

type FormatterStub struct {
	OutputCalled func(line *logger.LogLine) []byte
}

func (fs *FormatterStub) Output(line *logger.LogLine) []byte {
	return fs.OutputCalled(line)
}
