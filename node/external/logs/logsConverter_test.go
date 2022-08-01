package logs

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestLogsConverter_TxLogToApiResourceShouldWork(t *testing.T) {
	pkConverter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	logsConverter := newLogsConverter(pkConverter)

	contractAddressBech32 := "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt"
	contractAddress, _ := pkConverter.Decode(contractAddressBech32)

	txLog := &transaction.Log{
		Address: contractAddress,
		Events: []*transaction.Event{
			{
				Address:    contractAddress,
				Identifier: []byte("foo"),
				Topics:     [][]byte{{0xa}, {0xb}},
				Data:       []byte("data"),
			},
		},
	}

	expectedApiResource := &transaction.ApiLogs{
		Address: contractAddressBech32,
		Events: []*transaction.Events{
			{
				Address:    contractAddressBech32,
				Identifier: "foo",
				Topics:     [][]byte{{0xa}, {0xb}},
				Data:       []byte("data"),
			},
		},
	}

	apiResource := logsConverter.txLogToApiResource([]byte("aaaabbbb"), txLog)
	require.Equal(t, expectedApiResource, apiResource)
}
