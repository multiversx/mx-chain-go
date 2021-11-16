package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

func TestBaseSlashingDetector_IsRoundRelevant_DifferentRelevantAndIrrelevantRounds(t *testing.T) {
	t.Parallel()

	round := uint64(100)
	bsd := baseSlashingDetector{roundHandler: &mock.RoundHandlerMock{
		RoundIndex: int64(round),
	}}

	tests := []struct {
		round    func() uint64
		relevant bool
	}{
		{
			round: func() uint64 {
				return round
			},
			relevant: true,
		},
		{
			round: func() uint64 {
				return round - MaxDeltaToCurrentRound + 1
			},
			relevant: true,
		},
		{
			round: func() uint64 {
				return round + MaxDeltaToCurrentRound - 1
			},
			relevant: true,
		},
		{
			round: func() uint64 {
				return round - MaxDeltaToCurrentRound
			},
			relevant: false,
		},
		{
			round: func() uint64 {
				return round + MaxDeltaToCurrentRound
			},
			relevant: false,
		},
		{
			round: func() uint64 {
				return round - MaxDeltaToCurrentRound - 1
			},
			relevant: false,
		},
		{
			round: func() uint64 {
				return round + MaxDeltaToCurrentRound + 1
			},
			relevant: false,
		},
	}

	for _, currTest := range tests {
		require.Equal(t, currTest.relevant, bsd.isRoundRelevant(currTest.round()))
	}
}

func TestBaseSlashingDetector_CheckAndGetHeader(t *testing.T) {
	header, err := checkAndGetHeader(nil)
	require.Nil(t, header)
	require.Equal(t, process.ErrNilInterceptedData, err)

	header, err = checkAndGetHeader(&testscommon.InterceptedDataStub{})
	require.Nil(t, header)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)

	header, err = checkAndGetHeader(&interceptedBlocks.InterceptedHeader{})
	require.Nil(t, header)
	require.Equal(t, process.ErrNilHeaderHandler, err)

	blockHeader := &block.Header{Round: 1}
	interceptedHeader := slashMocks.CreateInterceptedHeaderData(blockHeader)
	header, err = checkAndGetHeader(interceptedHeader)
	require.Equal(t, blockHeader, header)
	require.Nil(t, err)
}

func TestBaseSlashingDetector_CheckSlashLevelBasedOnHeadersCount_DifferentSlashLevelsAndTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() (coreSlash.ThreatLevel, slash.HeaderList)
		expectedErr error
	}{
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				return coreSlash.Zero, slash.HeaderList{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				return coreSlash.ThreatLevel(9999), slash.HeaderList{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				return coreSlash.Medium, slash.HeaderList{}
			},
			expectedErr: process.ErrNotEnoughHeadersProvided,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 3}}
				return coreSlash.Medium, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}}
				return coreSlash.High, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}}
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: nil,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 3}}
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		level, headers := currTest.args()

		err := checkThreatLevelBasedOnHeadersCount(headers, level)
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestBaseSlashingDetector_CheckProofType(t *testing.T) {
	err := checkProofType(nil, coreSlash.MultipleProposal)
	require.Equal(t, process.ErrNilProof, err)

	proof := &slashMocks.SlashingProofStub{
		GetTypeCalled: func() coreSlash.SlashingType {
			return coreSlash.MultipleProposal
		},
	}
	err = checkProofType(proof, coreSlash.MultipleSigning)
	require.Equal(t, process.ErrInvalidSlashType, err)

	err = checkProofType(proof, coreSlash.MultipleProposal)
	require.Nil(t, err)
}
