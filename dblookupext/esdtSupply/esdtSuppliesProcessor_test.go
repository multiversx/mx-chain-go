package esdtSupply

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

const (
	testNftCreateValue     = 10
	testAddQuantityValue   = 50
	testBurnValue          = 30
	testFungibleTokenMint  = 100
	testFungibleTokenMint2 = 75
	testFungibleTokenBurn  = 25
)

func TestNewSuppliesProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewSuppliesProcessor(nil, &storageStubs.StorerStub{}, &storageStubs.StorerStub{})
	require.Equal(t, core.ErrNilMarshalizer, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, nil, &storageStubs.StorerStub{})
	require.Equal(t, core.ErrNilStore, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &storageStubs.StorerStub{}, nil)
	require.Equal(t, core.ErrNilStore, err)

	proc, err := NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &storageStubs.StorerStub{}, &storageStubs.StorerStub{})
	require.Nil(t, err)
	require.NotNil(t, proc)
	require.False(t, proc.IsInterfaceNil())
}

func TestProcessLogsSaveSupply(t *testing.T) {
	t.Parallel()

	token := []byte("nft-0001")
	logs := []*data.LogData{
		{
			TxHash: "txLog",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte("something"),
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(testNftCreateValue).Bytes(),
						},
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTAddQuantity),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(testAddQuantityValue).Bytes(),
						},
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTBurn),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(testBurnValue).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "log",
		},
	}

	putCalledNum := 0
	marshalizer := testscommon.MarshalizerMock{}
	suppliesStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == "processed-block" {
				pbn := ProcessedBlockNonce{Nonce: 5}
				pbnB, _ := marshalizer.Marshal(pbn)
				return pbnB, nil
			}

			return nil, storage.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			if string(key) == "processed-block" {
				return nil
			}

			var supplyKey string
			if putCalledNum%2 == 0 {
				supplyKey = string(token) + "-" + hex.EncodeToString(big.NewInt(2).Bytes())
			} else {
				supplyKey = string(token)
			}
			require.Equal(t, supplyKey, string(key))

			var supplyESDT SupplyESDT
			_ = marshalizer.Unmarshal(&supplyESDT, data)
			require.Equal(t, big.NewInt(30), supplyESDT.Supply)

			putCalledNum++
			return nil
		},
	}

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, &storageStubs.StorerStub{})
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(6, logs)
	require.Nil(t, err)

	require.Equal(t, 2, putCalledNum)
}

func TestProcessLogsSaveSupplyShouldUpdateSupplyMintedAndBurned(t *testing.T) {
	t.Parallel()

	token := []byte("nft-0001")
	logsCreate := []*data.LogData{
		{
			TxHash: "txLog",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte("something"),
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(testNftCreateValue).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "log",
		},
	}
	logsAddQuantity := []*data.LogData{
		{
			TxHash: "txLog",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte("something"),
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTAddQuantity),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(testAddQuantityValue).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "log",
		},
	}

	logsBurn := []*data.LogData{
		{
			TxHash: "txLog",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte("something"),
					},
					{
						Identifier: []byte(core.BuiltInFunctionESDTNFTBurn),
						Topics: [][]byte{
							token, big.NewInt(2).Bytes(), big.NewInt(testBurnValue).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "log",
		},
	}

	membDB := testscommon.NewMemDbMock()
	marshalizer := testscommon.MarshalizerMock{}
	numTimesCalled := 0
	suppliesStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == processedBlockKey {
				pbn := ProcessedBlockNonce{Nonce: 5}
				pbnB, _ := marshalizer.Marshal(pbn)
				return pbnB, nil
			}

			val, err := membDB.Get(key)
			if err != nil {
				return nil, storage.ErrKeyNotFound
			}
			return val, nil
		},
		PutCalled: func(key, data []byte) error {
			if string(key) == processedBlockKey {
				return nil
			}

			switch numTimesCalled {
			case 0:
			case 1:
				supplyEsdt := getSupplyESDT(marshalizer, data)
				require.Equal(t, big.NewInt(testNftCreateValue), supplyEsdt.Supply)
				require.Equal(t, big.NewInt(0), supplyEsdt.Burned)
				require.Equal(t, big.NewInt(testNftCreateValue), supplyEsdt.Minted)
			case 2:
			case 3:
				supplyEsdt := getSupplyESDT(marshalizer, data)
				require.Equal(t, big.NewInt(testNftCreateValue+testAddQuantityValue), supplyEsdt.Supply)
				require.Equal(t, big.NewInt(0), supplyEsdt.Burned)
				require.Equal(t, big.NewInt(testNftCreateValue+testAddQuantityValue), supplyEsdt.Minted)
			case 4:
			case 5:
				supplyEsdt := getSupplyESDT(marshalizer, data)
				require.Equal(t, big.NewInt(testNftCreateValue+testAddQuantityValue-testBurnValue), supplyEsdt.Supply)
				require.Equal(t, big.NewInt(testBurnValue), supplyEsdt.Burned)
				require.Equal(t, big.NewInt(testNftCreateValue+testAddQuantityValue), supplyEsdt.Minted)
			}

			_ = membDB.Put(key, data)
			numTimesCalled++

			return nil
		},
	}

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, &storageStubs.StorerStub{})
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(6, logsCreate)
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(7, logsAddQuantity)
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(8, logsBurn)
	require.Nil(t, err)

	require.Equal(t, 6, numTimesCalled)
}

func TestProcessLogs_RevertChangesShouldWorkForRevertingMinting(t *testing.T) {
	t.Parallel()

	token := []byte("BRT-1q2w3e")
	logsMintNoRevert := []*data.LogData{
		{
			TxHash: "txLog0",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte(core.BuiltInFunctionESDTLocalMint),
						Topics: [][]byte{
							token, nil, big.NewInt(testFungibleTokenMint).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "txLog1",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte(core.BuiltInFunctionESDTLocalMint),
						Topics: [][]byte{
							token, nil, big.NewInt(testFungibleTokenMint).Bytes(),
						},
					},
				},
			},
		},
	}

	mintLogToBeReverted := &transaction.Log{
		Events: []*transaction.Event{
			{
				Identifier: []byte(core.BuiltInFunctionESDTLocalMint),
				Topics: [][]byte{
					token, nil, big.NewInt(testFungibleTokenMint2).Bytes(),
				},
			},
		},
	}

	logsMintRevert := []*data.LogData{
		{
			TxHash:     "txLog3",
			LogHandler: mintLogToBeReverted,
		},
	}

	marshalizer := testscommon.MarshalizerMock{}

	logsStorer := genericMocks.NewStorerMockWithErrKeyNotFound(0)
	mintLogToBeRevertedBytes, err := marshalizer.Marshal(mintLogToBeReverted)
	require.NoError(t, err)
	err = logsStorer.Put([]byte("txHash3"), mintLogToBeRevertedBytes)
	require.NoError(t, err)

	suppliesStorer := genericMocks.NewStorerMockWithErrKeyNotFound(0)

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, logsStorer)
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(6, logsMintNoRevert)
	require.Nil(t, err)
	checkStoredValues(t, suppliesStorer, token, marshalizer, testFungibleTokenMint*2, testFungibleTokenMint*2, 0)

	err = suppliesProc.ProcessLogs(7, logsMintRevert)
	require.Nil(t, err)
	checkStoredValues(t, suppliesStorer, token, marshalizer,
		testFungibleTokenMint*2+testFungibleTokenMint2,
		testFungibleTokenMint*2+testFungibleTokenMint2, 0)

	revertedHeader := block.Header{Nonce: 7}
	blockBody := block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		},
	}
	err = suppliesProc.RevertChanges(&revertedHeader, &blockBody)
	require.NoError(t, err)
	checkStoredValues(t, suppliesStorer, token, marshalizer,
		testFungibleTokenMint*2,
		testFungibleTokenMint*2, 0)
}

func TestProcessLogs_RevertChangesShouldWorkForRevertingBurning(t *testing.T) {
	t.Parallel()

	token := []byte("BRT-1q2w3e")
	logsMintNoRevert := []*data.LogData{
		{
			TxHash: "txLog0",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte(core.BuiltInFunctionESDTLocalMint),
						Topics: [][]byte{
							token, nil, big.NewInt(testFungibleTokenMint).Bytes(),
						},
					},
				},
			},
		},
		{
			TxHash: "txLog1",
			LogHandler: &transaction.Log{
				Events: []*transaction.Event{
					{
						Identifier: []byte(core.BuiltInFunctionESDTLocalMint),
						Topics: [][]byte{
							token, nil, big.NewInt(testFungibleTokenMint).Bytes(),
						},
					},
				},
			},
		},
	}

	mintLogToBeReverted := &transaction.Log{
		Events: []*transaction.Event{
			{
				Identifier: []byte(core.BuiltInFunctionESDTLocalBurn),
				Topics: [][]byte{
					token, nil, big.NewInt(testFungibleTokenBurn).Bytes(),
				},
			},
		},
	}

	logsMintRevert := []*data.LogData{
		{
			TxHash:     "txLog3",
			LogHandler: mintLogToBeReverted,
		},
	}

	marshalizer := testscommon.MarshalizerMock{}

	logsStorer := genericMocks.NewStorerMockWithErrKeyNotFound(0)
	mintLogToBeRevertedBytes, err := marshalizer.Marshal(mintLogToBeReverted)
	require.NoError(t, err)
	err = logsStorer.Put([]byte("txHash3"), mintLogToBeRevertedBytes)
	require.NoError(t, err)

	suppliesStorer := genericMocks.NewStorerMockWithErrKeyNotFound(0)

	suppliesProc, err := NewSuppliesProcessor(marshalizer, suppliesStorer, logsStorer)
	require.Nil(t, err)

	err = suppliesProc.ProcessLogs(6, logsMintNoRevert)
	require.Nil(t, err)
	checkStoredValues(t, suppliesStorer, token, marshalizer, testFungibleTokenMint*2, testFungibleTokenMint*2, 0)

	err = suppliesProc.ProcessLogs(7, logsMintRevert)
	require.Nil(t, err)
	checkStoredValues(t,
		suppliesStorer,
		token,
		marshalizer,
		testFungibleTokenMint*2-testFungibleTokenBurn,
		testFungibleTokenMint*2,
		testFungibleTokenBurn)

	revertedHeader := block.Header{Nonce: 7}
	blockBody := block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("txHash3"),
				},
			},
		},
	}
	err = suppliesProc.RevertChanges(&revertedHeader, &blockBody)
	require.NoError(t, err)
	checkStoredValues(t,
		suppliesStorer,
		token,
		marshalizer,
		testFungibleTokenMint*2,
		testFungibleTokenMint*2,
		0)
}

func checkStoredValues(t *testing.T, suppliesStorer storage.Storer, token []byte, marshalizer marshal.Marshalizer, supply uint64, minted uint64, burnt uint64) {
	storedSupplyBytes, err := suppliesStorer.Get(token)
	require.NoError(t, err)

	var recoveredSupply SupplyESDT
	err = marshalizer.Unmarshal(&recoveredSupply, storedSupplyBytes)
	require.NoError(t, err)
	require.NotNil(t, recoveredSupply)

	require.Equal(t, supply, recoveredSupply.Supply.Uint64())
	require.Equal(t, minted, recoveredSupply.Minted.Uint64())
	require.Equal(t, burnt, recoveredSupply.Burned.Uint64())
}

func getSupplyESDT(marshalizer marshal.Marshalizer, data []byte) SupplyESDT {
	var supplyESDT SupplyESDT
	_ = marshalizer.Unmarshal(&supplyESDT, data)

	makePropertiesNotNil(&supplyESDT)
	return supplyESDT
}

func TestSupplyESDT_GetSupply(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.MarshalizerMock{}
	proc, _ := NewSuppliesProcessor(marshalizer, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == "my-token" {
				supply := &SupplyESDT{Supply: big.NewInt(123456)}
				return marshalizer.Marshal(supply)
			}
			return nil, errors.New("local err")
		},
	}, &storageStubs.StorerStub{})

	res, err := proc.GetESDTSupply("my-token")
	require.Nil(t, err)
	expectedESDTSupply := &SupplyESDT{
		Supply: big.NewInt(123456),
		Burned: big.NewInt(0),
		Minted: big.NewInt(0),
	}

	require.Equal(t, expectedESDTSupply, res)
}
