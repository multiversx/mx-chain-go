package headerCheck

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func createHeaderSigVerifierArgs() *ArgsHeaderSigVerifier {
	return &ArgsHeaderSigVerifier{
		Marshalizer:       &mock.MarshalizerMock{},
		Hasher:            &mock.HasherMock{},
		NodesCoordinator:  &mock.NodesCoordinatorMock{},
		MultiSigVerifier:  mock.NewMultiSigner(),
		SingleSigVerifier: &mock.SignerMock{},
		KeyGen:            &mock.SingleSignKeyGenMock{},
	}
}

func TestHeaderSigVerifier_VerifySignatureNilPrevRandSeedShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Equal(t, sharding.ErrNilRandomness, err)
}

func TestHeaderSigVerifier_VerifyRandSeed(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	wasCalled := false

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			wasCalled = true
			return nil
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {
			v, _ := sharding.NewValidator(big.NewInt(0), 1, pkAddr, pkAddr)
			return []sharding.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nodesCoordinator
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Nil(t, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifyRandSeedShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	wasCalled := false
	localError := errors.New("err")

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			wasCalled = true
			return localError
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {
			v, _ := sharding.NewValidator(big.NewInt(0), 1, pkAddr, pkAddr)
			return []sharding.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nodesCoordinator
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Equal(t, localError, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifyRandSeedAndLeaderSignature(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	count := 0

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			count++
			return nil
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nodesCoordinator := &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validators []sharding.Validator, err error) {
			v, _ := sharding.NewValidator(big.NewInt(0), 1, pkAddr, pkAddr)
			return []sharding.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nodesCoordinator
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Nil(t, err)
	require.Equal(t, 2, count)
}

func TestHeaderSigVerifier_VerifySignatureNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestHeaderSigVerifier_VerifySignatureBlockProposerSigMissingShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("0"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, process.ErrBlockProposerSignatureMissing, err)
}
