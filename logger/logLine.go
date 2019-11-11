package logger

import (
	"io"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger/capnp"
	protobuf "github.com/ElrondNetwork/elrond-go/logger/proto"
	capn "github.com/glycerine/go-capnproto"
)

// LogLine is the structure used to hold a log line
type LogLine struct {
	Message   string
	LogLevel  LogLevel
	Args      []interface{}
	Timestamp time.Time
}

func newLogLine(message string, logLevel LogLevel, args ...interface{}) *LogLine {
	return &LogLine{
		Message:   message,
		LogLevel:  logLevel,
		Args:      args,
		Timestamp: time.Now(),
	}
}

// LogLineWrapper is a wrapper over protobuf.LogLineMessage that enables the structure to be used with capnp marshalizer
type LogLineWrapper struct {
	protobuf.LogLineMessage
}

// Save saves the serialized data of a log line message into a stream through Capnp protocol
func (llw *LogLineWrapper) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	logLineWrapperGoToCapn(seg, llw)
	_, err := seg.WriteTo(w)

	return err
}

// Load loads the data from the stream into a log line message object through Capnp protocol
func (llw *LogLineWrapper) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootLogLineMessageCapn(capMsg)
	logLineWrapperCapnToGo(z, llw)

	return nil
}

func logLineWrapperGoToCapn(seg *capn.Segment, src *LogLineWrapper) capnp.LogLineMessageCapn {
	dest := capnp.AutoNewLogLineMessageCapn(seg)

	dest.SetLogLevel(src.LogLevel)
	dest.SetMessage(src.Message)
	dest.SetTimestamp(src.Timestamp)

	args := seg.NewTextList(len(src.Args))
	for i, arg := range src.Args {
		args.Set(i, arg)
	}
	dest.SetArgs(args)

	return dest
}

func logLineWrapperCapnToGo(src capnp.LogLineMessageCapn, dest *LogLineWrapper) {
	dest.Message = src.Message()
	dest.Timestamp = src.Timestamp()
	dest.LogLevel = src.LogLevel()

	dest.Args = make([]string, src.Args().Len())
	for i := 0; i < src.Args().Len(); i++ {
		dest.Args[i] = src.Args().At(i)
	}

	return
}
