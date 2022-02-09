package debug

import (
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

func TestDoubleTransactionsDetector_AddTxHash(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")

	detector := NewDoubleTransactionsDetector("test")
	log.Debug("empty")
	detector.PrintReport()

	detector.AddTxHash([]byte("aaaa"), []byte("bbbb"))
	log.Debug("one aaaa in bbbb")
	detector.PrintReport()

	detector.AddTxHash([]byte("aaaa"), []byte("bbbb"))
	log.Debug("two aaaa in bbbb")
	detector.PrintReport()

	detector.AddTxHash([]byte("aaaa"), []byte("ccc"))
	log.Debug("two aaaa in bbbb and one aaaa in ccc")
	detector.PrintReport()

	detector.Clear()
	log.Debug("empty: cleared")
	detector.PrintReport()

	detector.AddTxHash([]byte("aaaa"), []byte("bbbb"))
	log.Debug("one aaaa in bbbb")
	detector.PrintReport()

	detector.AddTxHash([]byte("aaaa"), []byte("ccc"))
	log.Debug("one aaaa in bbbb and one aaaa in ccc")
	detector.PrintReport()
}
