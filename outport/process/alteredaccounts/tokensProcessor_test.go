package alteredaccounts

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/require"
)

func TestTokenProcessorProcessEventNotEnoughTopics(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	enableEpochsHandler.IsScToScEventLogEnabledField = true
	tp := newTokensProcessor(&mock.ShardCoordinatorStub{}, enableEpochsHandler)

	markedAccounts := make(map[string]*markedAlteredAccount)
	tp.processEvent(&transaction.Event{
		Identifier: []byte(core.BuiltInFunctionMultiESDTNFTTransfer),
		Address:    []byte("addr"),
		Topics:     [][]byte{[]byte("0"), []byte("1"), []byte("2")},
	}, markedAccounts)

	require.Equal(t, 0, len(markedAccounts))
}

func TestTokenProcessorProcessEventMultiTransferV2(t *testing.T) {
	t.Parallel()

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	enableEpochsHandler.IsScToScEventLogEnabledField = true
	tp := newTokensProcessor(&mock.ShardCoordinatorStub{}, enableEpochsHandler)

	markedAccounts := make(map[string]*markedAlteredAccount)
	tp.processEvent(&transaction.Event{
		Identifier: []byte(core.BuiltInFunctionMultiESDTNFTTransfer),
		Address:    []byte("addr"),
		Topics:     [][]byte{[]byte("token1"), big.NewInt(0).Bytes(), []byte("2"), []byte("token2"), big.NewInt(1).Bytes(), []byte("3"), []byte("receiver")},
	}, markedAccounts)

	require.Equal(t, 2, len(markedAccounts))
	markedAccount := &markedAlteredAccount{
		tokens: map[string]*markedAlteredAccountToken{
			"token1": {
				identifier: "token1",
				nonce:      0,
			},
			"token2" + string(big.NewInt(1).Bytes()): {
				identifier: "token2",
				nonce:      1,
			},
		},
	}
	require.Equal(t, markedAccount, markedAccounts["addr"])
	require.Equal(t, markedAccount, markedAccounts["receiver"])
}
