package versioning

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestTxVersionChecker_IsSignedWithHashOptionsZeroShouldReturnFalse(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	tx := &transaction.Transaction{
		Options: 0,
		Version: minTxVersion,
	}
	tvc := NewTxVersionChecker(minTxVersion, tx)

	res := tvc.IsSignedWithHash()
	require.False(t, res)
}

func TestTxVersionChecker_IsSignedWithHash(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	tx := &transaction.Transaction{
		Options: 1 | maskSignedWithHash,
		Version: minTxVersion + 1,
	}
	tvc := NewTxVersionChecker(minTxVersion, tx)

	res := tvc.IsSignedWithHash()
	require.True(t, res)
}

func TestTxVersionChecker_CheckTxVersionShouldReturnErrorOptionsNotZero(t *testing.T) {
	minTxVersion := uint32(1)
	tx := &transaction.Transaction{
		Options: 1 | maskSignedWithHash,
		Version: minTxVersion,
	}

	tvc := NewTxVersionChecker(minTxVersion, tx)
	err := tvc.CheckTxVersion()
	require.Equal(t, core.ErrInvalidTransactionVersion, err)
}

func TestTxVersionChecker_CheckTxVersionShould(t *testing.T) {
	minTxVersion := uint32(1)
	tx := &transaction.Transaction{
		Options: 0,
		Version: minTxVersion,
	}

	tvc := NewTxVersionChecker(minTxVersion, tx)
	err := tvc.CheckTxVersion()
	require.Nil(t, err)
}
