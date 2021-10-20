package slash_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

func TestToProtoMultipleHeaderProposal_NilHeaders_ExpectError(t *testing.T) {
	proof := &slashMocks.MultipleHeaderProposalProofStub{
		GetHeadersCalled: func() []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{nil}
		},
	}
	res, err := slash.ToProtoMultipleHeaderProposal(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestToProtoMultipleHeaderProposal(t *testing.T) {
	h1 := &block.HeaderV2{Header: &block.Header{Nonce: 1, Round: 1}}
	h2 := &block.HeaderV2{Header: &block.Header{Nonce: 2, Round: 2}}

	interceptedHeader1 := slashMocks.CreateInterceptedHeaderData(h1)
	interceptedHeader2 := slashMocks.CreateInterceptedHeaderData(h2)

	proof := &slashMocks.MultipleHeaderProposalProofStub{
		GetHeadersCalled: func() []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{interceptedHeader1, interceptedHeader2}
		},
		GetLevelCalled: func() slash.ThreatLevel {
			return slash.High
		},
	}

	expectedRes := &coreSlash.MultipleHeaderProposalProof{
		Level:   coreSlash.High,
		Headers: &coreSlash.Headers{Headers: []*block.HeaderV2{h1, h2}},
	}
	res, err := slash.ToProtoMultipleHeaderProposal(proof)
	require.Nil(t, err)
	require.Equal(t, expectedRes, res)
}

func TestToProtoMultipleHeaderSign_NilHeaders_ExpectError(t *testing.T) {
	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{[]byte("address")}
		},
		GetHeadersCalled: func([]byte) []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{nil}
		},
	}
	res, err := slash.ToProtoMultipleHeaderSign(proof)

	require.Nil(t, res)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestToProtoMultipleHeaderSign(t *testing.T) {
	h1 := &block.HeaderV2{Header: &block.Header{Nonce: 1, Round: 1}}
	h2 := &block.HeaderV2{Header: &block.Header{Nonce: 2, Round: 2}}

	interceptedHeader1 := slashMocks.CreateInterceptedHeaderData(h1)
	interceptedHeader2 := slashMocks.CreateInterceptedHeaderData(h2)

	pk1 := []byte("pubKey1")
	pk2 := []byte("pubKey2")

	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetPubKeysCalled: func() [][]byte {
			return [][]byte{pk1, pk2}
		},
		GetHeadersCalled: func(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
			switch string(pubKey) {
			case string(pk1):
				return []*interceptedBlocks.InterceptedHeader{interceptedHeader1}
			case string(pk2):
				return []*interceptedBlocks.InterceptedHeader{interceptedHeader1, interceptedHeader2}
			default:
				return nil
			}
		},
		GetLevelCalled: func(pubKey []byte) slash.ThreatLevel {
			switch string(pubKey) {
			case string(pk1):
				return slash.Medium
			case string(pk2):
				return slash.High
			default:
				return slash.Low
			}
		},
	}

	expectedRes := &coreSlash.MultipleHeaderSigningProof{
		PubKeys: [][]byte{pk1, pk2},
		Levels: map[string]coreSlash.ThreatLevel{
			string(pk1): coreSlash.Medium,
			string(pk2): coreSlash.High,
		},
		Headers: map[string]*coreSlash.Headers{
			string(pk1): {Headers: []*block.HeaderV2{h1}},
			string(pk2): {Headers: []*block.HeaderV2{h1, h2}},
		},
	}
	res, err := slash.ToProtoMultipleHeaderSign(proof)
	require.Nil(t, err)
	require.Equal(t, expectedRes, res)
}
