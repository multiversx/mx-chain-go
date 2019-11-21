package logger

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// logLineWrapperFormatter converts the LogLineHandler into its marshalized form
type logLineWrapperFormatter struct {
	marshalizer marshal.Marshalizer
}

// NewLogLineWrapperFormatter creates a new logLineWrapperFormatter that is able to marshalize the provided logLine
func NewLogLineWrapperFormatter(marshalizer marshal.Marshalizer) (*logLineWrapperFormatter, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	return &logLineWrapperFormatter{
		marshalizer: marshalizer,
	}, nil
}

// Output converts the provided LogLineHandler into a slice of bytes ready for output
func (llwf *logLineWrapperFormatter) Output(line LogLineHandler) []byte {
	if check.IfNil(line) {
		return nil
	}

	buff, err := llwf.marshalizer.Marshal(line)
	if err == nil {
		return buff
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (llwf *logLineWrapperFormatter) IsInterfaceNil() bool {
	return llwf == nil
}
