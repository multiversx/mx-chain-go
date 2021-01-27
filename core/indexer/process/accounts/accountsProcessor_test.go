package accounts

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountsProcessor_ComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	tests := []struct {
		input  *big.Int
		output float64
	}{
		{
			input:  big.NewInt(200000000000000000),
			output: float64(20000000),
		},
		{
			input:  big.NewInt(57777777777),
			output: 5.7777777777,
		},
		{
			input:  big.NewInt(5777779),
			output: 0.0005777779,
		},
		{
			input:  big.NewInt(7),
			output: 0.0000000007,
		},
		{
			input:  big.NewInt(-7),
			output: 0.0,
		},
	}

	for _, tt := range tests {
		out := ap.computeBalanceAsFloat(tt.input)
		assert.Equal(t, tt.output, out)
	}
}

func TestGetESDTInfo_CannotRetriveValueShoudError(t *testing.T) {
	t.Parallel()

	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	localErr := errors.New("local error")
	wrapAccount := &types.AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						return nil, localErr
					},
				}
			},
		},
		TokenIdentifier: "token",
	}
	_, _, err := ap.getESDTInfo(wrapAccount)
	require.Equal(t, localErr, err)
}

func TestGetESDTInfo(t *testing.T) {
	t.Parallel()

	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
	}

	tokenIdentifier := "token-001"
	wrapAccount := &types.AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						return json.Marshal(esdtToken)
					},
				}
			},
		},
		TokenIdentifier: tokenIdentifier,
	}
	balance, prop, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1000), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
}

func TestAccountsProcessor_GetAccountsEGLDAccounts(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{}
	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	alteredAccounts := map[string]*types.AlteredAccount{
		addr: {
			IsESDTOperation: false,
			TokenIdentifier: "",
		},
	}
	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts)
	require.Equal(t, 0, len(esdtAccounts))
	require.Equal(t, []*types.AccountEGLD{
		{Account: mockAccount},
	}, accounts)
}

func TestAccountsProcessor_GetAccountsESDTAccount(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{}
	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	alteredAccounts := map[string]*types.AlteredAccount{
		addr: {
			IsESDTOperation: true,
			TokenIdentifier: "token",
		},
	}
	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts)
	require.Equal(t, 0, len(accounts))
	require.Equal(t, []*types.AccountESDT{
		{Account: mockAccount, TokenIdentifier: "token"},
	}, esdtAccounts)
}

func TestAccountsProcessor_PrepareAccountsMapEGLD(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		GetNonceCalled: func() uint64 {
			return 1
		},
		GetBalanceCalled: func() *big.Int {
			return big.NewInt(1000)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}

	egldAccount := &types.AccountEGLD{
		Account:  mockAccount,
		IsSender: false,
	}

	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	res := ap.PrepareAccountsMapEGLD([]*types.AccountEGLD{egldAccount})
	require.Equal(t, map[string]*types.AccountInfo{
		hex.EncodeToString([]byte(addr)): {
			Nonce:      1,
			Balance:    "1000",
			BalanceNum: ap.computeBalanceAsFloat(big.NewInt(1000)),
		},
	}, res)
}

func TestAccountsProcessor_PrepareAccountsMapESDT(t *testing.T) {
	t.Parallel()

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
	}

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return json.Marshal(esdtToken)
				},
			}
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (state.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	res := ap.PrepareAccountsMapESDT([]*types.AccountESDT{{Account: mockAccount, TokenIdentifier: "token"}})
	require.Equal(t, map[string]*types.AccountInfo{
		hex.EncodeToString([]byte(addr)): {
			Address:         hex.EncodeToString([]byte(addr)),
			Balance:         "1000",
			BalanceNum:      ap.computeBalanceAsFloat(big.NewInt(1000)),
			TokenIdentifier: "token",
			Properties:      hex.EncodeToString([]byte("ok")),
		},
	}, res)
}
