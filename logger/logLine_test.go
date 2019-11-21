package logger_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/logger/mock"
	protobuf "github.com/ElrondNetwork/elrond-go/logger/proto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

func TestLogLineWrapper_MarshalUnmarshalShouldWork(t *testing.T) {
	llw := generateLogLineWrapper()

	for marshName, marsh := range mock.TestingMarshalizers {
		testMarshalUnmarshal(t, marshName, marsh, llw)
	}
}

func generateLogLineWrapper() logger.LogLineWrapper {
	return logger.LogLineWrapper{
		LogLineMessage: protobuf.LogLineMessage{
			Message:   "test message",
			LogLevel:  4,
			Args:      []string{"arg1", "arg2", "arg3", "arg4"},
			Timestamp: 11223344,
		},
	}
}

func testMarshalUnmarshal(t *testing.T, marshName string, marsh marshal.Marshalizer, llw logger.LogLineWrapper) {
	llwCopyForAssert := llw

	buff, err := marsh.Marshal(&llw)
	assert.Nil(t, err)

	llwRecovered := &logger.LogLineWrapper{}
	err = marsh.Unmarshal(llwRecovered, buff)

	assert.Equal(t, &llwCopyForAssert, llwRecovered, fmt.Sprintf("for marshalizer %v", marshName))
}
