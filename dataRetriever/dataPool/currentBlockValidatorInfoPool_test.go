package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCurrentBlockValidatorInfoPool_AddGetCleanTx(t *testing.T) {
	t.Parallel()

	validatorInfoHash := []byte("hash")
	validatorInfo := &state.ShardValidatorInfo{}
	currentValidatorInfoPool := NewCurrentBlockValidatorInfoPool()
	require.False(t, currentValidatorInfoPool.IsInterfaceNil())

	currentValidatorInfoPool.AddValidatorInfo(validatorInfoHash, validatorInfo)
	currentValidatorInfoPool.AddValidatorInfo(validatorInfoHash, nil)

	validatorInfoFromPool, err := currentValidatorInfoPool.GetValidatorInfo([]byte("wrong hash"))
	require.Nil(t, validatorInfoFromPool)
	require.Equal(t, dataRetriever.ErrValidatorInfoNotFoundInBlockPool, err)

	validatorInfoFromPool, err = currentValidatorInfoPool.GetValidatorInfo(validatorInfoHash)
	require.Nil(t, err)
	require.Equal(t, validatorInfo, validatorInfoFromPool)

	currentValidatorInfoPool.Clean()
	validatorInfoFromPool, err = currentValidatorInfoPool.GetValidatorInfo(validatorInfoHash)
	require.Nil(t, validatorInfoFromPool)
	require.Equal(t, dataRetriever.ErrValidatorInfoNotFoundInBlockPool, err)
}
