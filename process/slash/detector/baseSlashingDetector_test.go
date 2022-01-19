package detector

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process"
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
	tests := []struct {
		interceptedData process.InterceptedData
		expectedHeader  data.HeaderHandler
		expectedError   error
	}{
		{
			interceptedData: nil,
			expectedHeader:  nil,
			expectedError:   process.ErrNilInterceptedData,
		},
		{
			interceptedData: &testscommon.InterceptedDataStub{},
			expectedHeader:  nil,
			expectedError:   process.ErrCannotCastInterceptedDataToHeader,
		},
		{
			interceptedData: &testscommon.InterceptedHeaderStub{
				HeaderHandlerCalled: func() data.HeaderHandler {
					return nil
				},
				InterceptedDataStub: testscommon.InterceptedDataStub{
					HashCalled: func() []byte {
						return []byte("hash")
					},
				},
			},
			expectedHeader: nil,
			expectedError:  process.ErrNilHeaderHandler,
		},
		{
			interceptedData: &testscommon.InterceptedHeaderStub{
				HeaderHandlerCalled: func() data.HeaderHandler {
					return &testscommon.HeaderHandlerStub{}
				},
				InterceptedDataStub: testscommon.InterceptedDataStub{
					HashCalled: func() []byte {
						return nil
					},
				},
			},
			expectedHeader: nil,
			expectedError:  data.ErrNilHash,
		},
		{
			interceptedData: slashMocks.CreateInterceptedHeaderData(&block.Header{Round: 1}),
			expectedHeader:  &block.Header{Round: 1},
			expectedError:   nil,
		},
	}

	for _, test := range tests {
		header, err := getCheckedHeader(test.interceptedData)
		require.Equal(t, test.expectedHeader, header)
		require.Equal(t, test.expectedError, err)
	}
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
