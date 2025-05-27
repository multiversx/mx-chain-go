package esdtSupply

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestSaveCorrectionInfo(t *testing.T) {
	t.Parallel()

	called := false
	marshaller := &marshallerMock.MarshalizerMock{}
	supplyStorer := &storageStubs.StorerStub{
		PutCalled: func(key, data []byte) error {
			called = true
			return nil
		},
	}

	logsProc := newLogsProcessor(marshaller, supplyStorer)

	scp := newSupplyCorrectionProcessor(0, logsProc)
	err := scp.saveCorrectionInfo([]byte("key"), &Correction{WasApplied: true})
	require.Nil(t, err)
	require.True(t, called)
}

func TestGetCorrectionInfo(t *testing.T) {
	t.Parallel()

	correction := &Correction{
		WasApplied: true,
	}

	marshaller := &marshallerMock.MarshalizerMock{}
	correctionBytes, err := marshaller.Marshal(correction)
	require.Nil(t, err)

	key1 := "key1"

	supplyStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if string(key) == key1 {
				return nil, storage.ErrKeyNotFound
			}

			return correctionBytes, nil
		},
	}

	logsProc := newLogsProcessor(marshaller, supplyStorer)
	scp := newSupplyCorrectionProcessor(0, logsProc)

	correctionFromStorage, err := scp.getCorrectionInfo([]byte(key1))
	require.Nil(t, err)
	require.False(t, correctionFromStorage.WasApplied)

	correctionFromStorage, err = scp.getCorrectionInfo([]byte("key2"))
	require.Nil(t, err)
	require.True(t, correctionFromStorage.WasApplied)
}

func TestApplySupplyCorrectionEmpty(t *testing.T) {
	t.Parallel()

	supplyStorer := &storageStubs.StorerStub{}
	marshaller := &marshallerMock.MarshalizerMock{}
	logsProc := newLogsProcessor(marshaller, supplyStorer)
	scp := newSupplyCorrectionProcessor(0, logsProc)

	err := scp.applySupplyCorrections(nil)
	require.Nil(t, err)
}

func TestApplySupplyCorrectionMultipleEntries(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
	supplyStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			switch string(key) {
			case processedBlockKey:
				pb := &ProcessedBlockNonce{Nonce: 1500}
				pbBytes, _ := marshaller.Marshal(pb)
				return pbBytes, nil
			}
			return nil, storage.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			switch string(key) {
			case "TTT-0001":
				supply := &SupplyESDT{}
				err := marshaller.Unmarshal(supply, data)
				require.Nil(t, err)
				require.Equal(t, big.NewInt(100), supply.Supply)
				require.Equal(t, big.NewInt(100), supply.Minted)
			case "TTT-0002":
				supply := &SupplyESDT{}
				err := marshaller.Unmarshal(supply, data)
				require.Nil(t, err)
				require.Equal(t, big.NewInt(-10000), supply.Supply)
				require.Equal(t, big.NewInt(10000), supply.Burned)
			case "supplyCorrectionsc1", "supplyCorrectionsc2":
				correction := &Correction{}
				err := marshaller.Unmarshal(correction, data)
				require.Nil(t, err)
				require.True(t, correction.WasApplied)

			}
			return nil
		},
	}

	logsProc := newLogsProcessor(marshaller, supplyStorer)
	scp := newSupplyCorrectionProcessor(0, logsProc)

	correction := []config.SupplyCorrection{
		{
			ID:      "sc3",
			ShardID: 1,
		},
		{
			ID:         "sc1",
			ShardID:    0,
			BlockNonce: 1000,
			Token:      "TTT-0001",
			Value:      "100",
		},
		{
			ID:         "sc2",
			ShardID:    0,
			BlockNonce: 1200,
			Token:      "TTT-0002",
			Value:      "-10000",
		},
	}

	err := scp.applySupplyCorrections(correction)
	require.Nil(t, err)
}

func TestApplySupplyCorrectionSinglePositiveEntry(t *testing.T) {
	t.Parallel()

	marshaller := &marshallerMock.MarshalizerMock{}
	supplyStorer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			switch string(key) {
			case processedBlockKey:
				pb := &ProcessedBlockNonce{Nonce: 2000}
				pbBytes, _ := marshaller.Marshal(pb)
				return pbBytes, nil
			case "POS-0001":
				// Simulate existing supply
				supply := &SupplyESDT{
					Supply: big.NewInt(50),
					Minted: big.NewInt(50),
					Burned: big.NewInt(0),
				}
				return marshaller.Marshal(supply)
			}
			return nil, storage.ErrKeyNotFound
		},
		PutCalled: func(key, data []byte) error {
			switch string(key) {
			case "POS-0001":
				supply := &SupplyESDT{}
				err := marshaller.Unmarshal(supply, data)
				require.Nil(t, err)
				require.Equal(t, big.NewInt(150), supply.Supply)
				require.Equal(t, big.NewInt(150), supply.Minted)
			case "supplyCorrectionpos1":
				correction := &Correction{}
				err := marshaller.Unmarshal(correction, data)
				require.Nil(t, err)
				require.True(t, correction.WasApplied)
			}
			return nil
		},
	}

	logsProc := newLogsProcessor(marshaller, supplyStorer)
	scp := newSupplyCorrectionProcessor(0, logsProc)

	corrections := []config.SupplyCorrection{
		{
			ID:         "pos1",
			ShardID:    0,
			BlockNonce: 1000,
			Token:      "POS-0001",
			Value:      "100",
		},
	}

	err := scp.applySupplyCorrections(corrections)
	require.Nil(t, err)
}
