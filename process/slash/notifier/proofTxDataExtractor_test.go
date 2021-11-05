package notifier_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash/notifier"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

func TestNewProofTxDataExtractor_InvalidMarshaller(t *testing.T) {
	t.Parallel()

	extractor, err := notifier.NewProofTxDataExtractor(nil)
	require.Nil(t, extractor)
	require.Equal(t, process.ErrNilMarshalizer, err)
}

func TestProofTxDataExtractor_GetProofTxData_InvalidProof_ExpectError(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	proof := &slashMocks.SlashingProofStub{}
	res, err := extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrInvalidProof, err)
}

func TestProofTxDataExtractor_GetProofTxData_InvalidMarshaller_ExpectError(t *testing.T) {
	t.Parallel()

	errMarshaller := errors.New("marshaller error")
	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, errMarshaller
		},
	}

	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	var proof coreSlash.SlashingProofHandler

	proof = &slashMocks.MultipleHeaderProposalProofStub{}
	res, err := extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, errMarshaller, err)

	proof = &slashMocks.MultipleHeaderSigningProofStub{}
	res, err = extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, errMarshaller, err)
}

func TestProofTxDataExtractor_GetProofTxData_MultipleProposalProof_NotEnoughHeaders_ExpectError(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, nil
		},
	}
	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	proof := &slashMocks.MultipleHeaderProposalProofStub{}
	res, err := extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrNotEnoughHeadersProvided, err)
}

func TestProofTxDataExtractor_GetProofTxData_MultipleSigningProof_NotEnoughPubKeys_ExpectError(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, nil
		},
	}
	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	proof := &slashMocks.MultipleHeaderSigningProofStub{}
	res, err := extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrNotEnoughPubKeysProvided, err)
}

func TestProofTxDataExtractor_GetProofTxData_MultipleSigningProof_NotEnoughHeaders_ExpectError(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, nil
		},
	}
	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{[]byte("pubKey")}
		},
	}
	res, err := extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrNotEnoughHeadersProvided, err)
}

func TestProofTxDataExtractor_GetProofTxData_MultipleSigningProof_NilHeader_ExpectError(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, nil
		},
	}
	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{[]byte("pubKey")}
		},
		GetHeadersCalled: func([]byte) []data.HeaderHandler {
			return []data.HeaderHandler{nil}
		},
	}
	res, err := extractor.GetProofTxData(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestProofTxDataExtractor_GetProofTxData_MultipleProposalProof(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return []byte("abc"), nil
		},
	}
	header := &block.Header{Round: 2, ShardID: 1}
	proof := &slashMocks.MultipleHeaderProposalProofStub{
		GetHeadersCalled: func() []data.HeaderHandler {
			return []data.HeaderHandler{header}
		},
	}

	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	txData, _ := extractor.GetProofTxData(proof)
	expectedTxData := &notifier.ProofTxData{
		Round:     header.Round,
		ShardID:   header.ShardID,
		Bytes:     []byte("abc"),
		SlashType: coreSlash.MultipleProposal,
	}
	require.Equal(t, expectedTxData, txData)
}

func TestProofTxDataExtractor_GetProofTxData_MultipleSigningProof(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return []byte("abc"), nil
		},
	}
	header := &block.Header{Round: 2, ShardID: 1}
	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetHeadersCalled: func([]byte) []data.HeaderHandler {
			return []data.HeaderHandler{header}
		},
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{[]byte("pubKey")}
		},
	}

	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	txData, _ := extractor.GetProofTxData(proof)
	expectedTxData := &notifier.ProofTxData{
		Round:     header.Round,
		ShardID:   header.ShardID,
		Bytes:     []byte("abc"),
		SlashType: coreSlash.MultipleSigning,
	}
	require.Equal(t, expectedTxData, txData)
}

func TestProofTxDataExtractor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshalizerMock{}
	extractor, _ := notifier.NewProofTxDataExtractor(marshaller)

	require.False(t, extractor.IsInterfaceNil())
}
