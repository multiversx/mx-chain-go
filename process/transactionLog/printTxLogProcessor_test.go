package transactionLog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewPrintTxLogProcessor(t *testing.T) {
	t.Parallel()

	ptlp := NewPrintTxLogProcessor()
	require.NotNil(t, ptlp)
	require.False(t, ptlp.IsInterfaceNil())

	res := ptlp.GetAllCurrentLogs()
	require.NotNil(t, res)
	require.Len(t, res, 0)

	_, found := ptlp.GetLogFromCache(nil)
	require.False(t, found)

	ptlp.EnableLogToBeSavedInCache()
	ptlp.Clean()

	l, err := ptlp.GetLog(nil)
	require.Nil(t, l)
	require.Nil(t, err)

}

type plainFormatterExtended struct {
	logger.PlainFormatter
}

func (pf *plainFormatterExtended) Output(line logger.LogLineHandler) []byte {
	if line.GetMessage() != "printTxLogProcessor.SaveLog" && line.GetMessage() != "printTxLogProcessor.entry" {
		return nil
	}

	return pf.PlainFormatter.Output(line)
}

func TestPrintTxLogProcessor_SaveLog(t *testing.T) {
	log.SetLevel(logger.LogDebug)

	buff := &bytes.Buffer{}
	_ = logger.AddLogObserver(buff, &plainFormatterExtended{})

	ptlp := NewPrintTxLogProcessor()

	txLogEntry := []*vmcommon.LogEntry{
		{
			Address:    []byte("addr"),
			Identifier: []byte("identifier"),
			Topics:     [][]byte{[]byte("topic 1"), []byte("topic 2")},
		},
	}

	err := ptlp.SaveLog([]byte("hash"), &transaction.Transaction{}, txLogEntry)
	require.Nil(t, err)

	err = ptlp.SaveLog([]byte("hash"), &transaction.Transaction{}, nil)
	require.Nil(t, err)

	require.True(t, strings.Contains(buff.String(), "printTxLogProcessor.SaveLog"))
	require.True(t, strings.Contains(buff.String(), "printTxLogProcessor.entry"))
}
