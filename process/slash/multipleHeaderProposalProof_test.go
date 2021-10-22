package slash_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMultipleProposalProof(t *testing.T) {
	tests := []struct {
		args        func() *slash.SlashingResult
		expectedErr error
	}{
		{
			args: func() *slash.SlashingResult {
				return nil
			},
			expectedErr: process.ErrNilSlashResult,
		},
		{
			args: func() *slash.SlashingResult {
				return &slash.SlashingResult{SlashingLevel: slash.Medium, Headers: nil}
			},
			expectedErr: process.ErrNilHeaderHandler,
		},
		{
			args: func() *slash.SlashingResult {
				return &slash.SlashingResult{SlashingLevel: slash.Medium, Headers: slash.HeaderInfoList{}}
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := slash.NewMultipleProposalProof(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleProposalProof_GetHeaders_GetLevel_GetType(t *testing.T) {
	h1 := &slash.HeaderInfo{
		Header: &testscommon.HeaderHandlerStub{TimestampField: 1},
		Hash:   []byte("h1"),
	}
	h2 := &slash.HeaderInfo{
		Header: &testscommon.HeaderHandlerStub{TimestampField: 2},
		Hash:   []byte("h2"),
	}

	slashRes := &slash.SlashingResult{
		SlashingLevel: slash.Medium,
		Headers:       slash.HeaderInfoList{h1, h2}}

	proof, err := slash.NewMultipleProposalProof(slashRes)
	require.Nil(t, err)

	require.Equal(t, slash.MultipleProposal, proof.GetType())
	require.Equal(t, slash.Medium, proof.GetLevel())

	require.Len(t, proof.GetHeaders(), 2)
	require.Contains(t, proof.GetHeaders(), h1)
	require.Contains(t, proof.GetHeaders(), h2)
}
