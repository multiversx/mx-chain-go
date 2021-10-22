package slash_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMultipleSigningProof(t *testing.T) {
	tests := []struct {
		args        func() map[string]slash.SlashingResult
		expectedErr error
	}{
		{
			args: func() map[string]slash.SlashingResult {
				return nil
			},
			expectedErr: process.ErrNilSlashResult,
		},
		{
			args: func() map[string]slash.SlashingResult {
				return make(map[string]slash.SlashingResult)
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := slash.NewMultipleSigningProof(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleSigningProof_GetHeaders_GetLevel_GetType(t *testing.T) {
	h1 := &slash.HeaderInfo{
		Header: &testscommon.HeaderHandlerStub{TimestampField: 1},
		Hash:   []byte("h1"),
	}
	h2 := &slash.HeaderInfo{
		Header: &testscommon.HeaderHandlerStub{TimestampField: 2},
		Hash:   []byte("h2"),
	}

	slashRes1 := slash.SlashingResult{
		SlashingLevel: slash.Medium,
		Headers:       slash.HeaderInfoList{h1},
	}
	slashRes2 := slash.SlashingResult{
		SlashingLevel: slash.High,
		Headers:       slash.HeaderInfoList{h2},
	}
	slashRes := map[string]slash.SlashingResult{
		"pubKey1": slashRes1,
		"pubKey2": slashRes2,
	}

	proof, err := slash.NewMultipleSigningProof(slashRes)
	require.Nil(t, err)

	require.Equal(t, slash.MultipleSigning, proof.GetType())
	require.Equal(t, slash.Medium, proof.GetLevel([]byte("pubKey1")))
	require.Equal(t, slash.High, proof.GetLevel([]byte("pubKey2")))
	require.Equal(t, slash.Low, proof.GetLevel([]byte("pubKey3")))

	require.Len(t, proof.GetHeaders([]byte("pubKey1")), 1)
	require.Len(t, proof.GetHeaders([]byte("pubKey2")), 1)
	require.Len(t, proof.GetHeaders([]byte("pubKey3")), 0)

	require.Contains(t, proof.GetHeaders([]byte("pubKey1")), h1)
	require.Contains(t, proof.GetHeaders([]byte("pubKey2")), h2)
	require.Nil(t, proof.GetHeaders([]byte("pubKey3")))
}
