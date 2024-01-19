package headerCheck

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultChancesSelection = 1

func createHeaderSigVerifierArgs() *ArgsHeaderSigVerifier {
	return &ArgsHeaderSigVerifier{
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		NodesCoordinator:        &shardingMocks.NodesCoordinatorMock{},
		MultiSigContainer:       cryptoMocks.NewMultiSignerContainerMock(cryptoMocks.NewMultiSigner()),
		SingleSigVerifier:       &mock.SignerMock{},
		KeyGen:                  &mock.SingleSignKeyGenMock{},
		FallbackHeaderValidator: &testscommon.FallBackHeaderValidatorStub{},
		EnableEpochsHandler:     enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		HeadersPool:             &mock.HeadersCacherStub{},
	}
}

func TestNewHeaderSigVerifier_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	hdrSigVerifier, err := NewHeaderSigVerifier(nil)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewHeaderSigVerifier_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.Hasher = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilHasher, err)
}

func TestNewHeaderSigVerifier_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.KeyGen = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewHeaderSigVerifier_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.Marshalizer = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewHeaderSigVerifier_NilMultiSigShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(nil)
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewHeaderSigVerifier_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.NodesCoordinator = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewHeaderSigVerifier_NilSingleSigShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.SingleSigVerifier = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewHeaderSigVerifier_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.EnableEpochsHandler = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewHeaderSigVerifier_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, err)
	require.NotNil(t, hdrSigVerifier)
	require.False(t, hdrSigVerifier.IsInterfaceNil())
}

func TestHeaderSigVerifier_VerifySignatureNilPrevRandSeedShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Equal(t, nodesCoordinator.ErrNilRandomness, err)
}

func TestHeaderSigVerifier_VerifyRandSeedOk(t *testing.T) {
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
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Nil(t, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifyRandSeedShouldErrWhenVerificationFails(t *testing.T) {
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
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Equal(t, localError, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifyRandSeedAndLeaderSignatureNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Equal(t, nodesCoordinator.ErrNilRandomness, err)
}

func TestHeaderSigVerifier_VerifyRandSeedAndLeaderSignatureVerifyShouldErrWhenValidationFails(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	count := 0
	localErr := errors.New("err")

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			count++
			return localErr
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Equal(t, localErr, err)
	require.Equal(t, 1, count)
}

func TestHeaderSigVerifier_VerifyRandSeedAndLeaderSignatureVerifyLeaderSigShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	count := 0
	localErr := errors.New("err")
	leaderSig := []byte("signature")

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			count++
			if bytes.Equal(sig, leaderSig) {
				return localErr
			}
			return nil
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		LeaderSignature: leaderSig,
	}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Equal(t, localErr, err)
	require.Equal(t, 2, count)
}

func TestHeaderSigVerifier_VerifyRandSeedAndLeaderSignatureOk(t *testing.T) {
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
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Nil(t, err)
	require.Equal(t, 2, count)
}

func TestHeaderSigVerifier_VerifyLeaderSignatureNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyLeaderSignature(header)
	require.Equal(t, nodesCoordinator.ErrNilRandomness, err)
}

func TestHeaderSigVerifier_VerifyLeaderSignatureVerifyShouldErrWhenValidationFails(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	count := 0
	localErr := errors.New("err")

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			count++
			return localErr
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyLeaderSignature(header)
	require.Equal(t, localErr, err)
	require.Equal(t, 1, count)
}

func TestHeaderSigVerifier_VerifyLeaderSignatureVerifyLeaderSigShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	count := 0
	localErr := errors.New("err")
	leaderSig := []byte("signature")

	args.KeyGen = &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
			return &mock.SingleSignPublicKey{}, nil
		},
	}
	args.SingleSigVerifier = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			count++
			if bytes.Equal(sig, leaderSig) {
				return localErr
			}
			return nil
		},
	}

	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		LeaderSignature: leaderSig,
	}

	err := hdrSigVerifier.VerifyLeaderSignature(header)
	require.Equal(t, localErr, err)
	require.Equal(t, 1, count)
}

func TestHeaderSigVerifier_VerifyLeaderSignatureOk(t *testing.T) {
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
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{}

	err := hdrSigVerifier.VerifyLeaderSignature(header)
	require.Nil(t, err)
	require.Equal(t, 1, count)
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

func TestHeaderSigVerifier_VerifySignatureNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("1"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, nodesCoordinator.ErrNilRandomness, err)
}

func TestHeaderSigVerifier_VerifySignatureWrongSizeBitmapShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("11"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, ErrWrongSizeBitmap, err)
}

func TestHeaderSigVerifier_VerifySignatureNotEnoughSigsShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v, v, v, v, v}, nil
		},
	}
	args.NodesCoordinator = nc

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("A"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, ErrNotEnoughSignatures, err)
}

func TestHeaderSigVerifier_VerifySignatureOk(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc

	args.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(&cryptoMocks.MultisignerMock{
		VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
			wasCalled = true
			return nil
		}})

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("1"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Nil(t, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifySignatureNotEnoughSigsShouldErrWhenFallbackThresholdCouldNotBeApplied(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v, v, v, v, v}, nil
		},
	}
	fallbackHeaderValidator := &testscommon.FallBackHeaderValidatorStub{
		ShouldApplyFallbackValidationCalled: func(headerHandler data.HeaderHandler) bool {
			return false
		},
	}
	multiSigVerifier := &cryptoMocks.MultisignerMock{
		VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
			wasCalled = true
			return nil
		},
	}

	args.NodesCoordinator = nc
	args.FallbackHeaderValidator = fallbackHeaderValidator
	args.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigVerifier)

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.MetaBlock{
		PubKeysBitmap: []byte("C"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, ErrNotEnoughSignatures, err)
	require.False(t, wasCalled)
}

func TestHeaderSigVerifier_VerifySignatureOkWhenFallbackThresholdCouldBeApplied(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return []nodesCoordinator.Validator{v, v, v, v, v}, nil
		},
	}
	fallbackHeaderValidator := &testscommon.FallBackHeaderValidatorStub{
		ShouldApplyFallbackValidationCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
	}
	multiSigVerifier := &cryptoMocks.MultisignerMock{
		VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
			wasCalled = true
			return nil
		}}

	args.NodesCoordinator = nc
	args.FallbackHeaderValidator = fallbackHeaderValidator
	args.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(multiSigVerifier)

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.MetaBlock{
		PubKeysBitmap: []byte("C"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Nil(t, err)
	require.True(t, wasCalled)
}

func TestCheckHeaderHandler_VerifyPreviousBlockProof(t *testing.T) {
	t.Parallel()

	t.Run("flag enabled and no proof should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.ConsensusPropagationChangesFlag
			},
		}

		hdrSigVerifier, _ := NewHeaderSigVerifier(args)

		hdr := &testscommon.HeaderHandlerStub{
			GetPreviousAggregatedSignatureAndBitmapCalled: func() ([]byte, []byte) {
				return nil, nil
			},
		}
		err := hdrSigVerifier.VerifyPreviousBlockProof(hdr)
		assert.True(t, errors.Is(err, process.ErrInvalidHeader))
		assert.True(t, strings.Contains(err.Error(), "received header without proof after flag activation"))
	})
	t.Run("flag not enabled and proof should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		hdrSigVerifier, _ := NewHeaderSigVerifier(args)

		hdr := &testscommon.HeaderHandlerStub{
			GetPreviousAggregatedSignatureAndBitmapCalled: func() ([]byte, []byte) {
				return []byte("sig"), []byte("bitmap")
			},
		}
		err := hdrSigVerifier.VerifyPreviousBlockProof(hdr)
		assert.True(t, errors.Is(err, process.ErrInvalidHeader))
		assert.True(t, strings.Contains(err.Error(), "received header with proof before flag activation"))
	})
	t.Run("flag enabled and no leader signature should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.ConsensusPropagationChangesFlag
			},
		}

		hdrSigVerifier, _ := NewHeaderSigVerifier(args)

		hdr := &testscommon.HeaderHandlerStub{
			GetPreviousAggregatedSignatureAndBitmapCalled: func() ([]byte, []byte) {
				return []byte("sig"), []byte{0, 1, 1, 1}
			},
		}
		err := hdrSigVerifier.VerifyPreviousBlockProof(hdr)
		assert.True(t, errors.Is(err, process.ErrInvalidHeader))
		assert.True(t, strings.Contains(err.Error(), "received header without leader signature after flag activation"))
	})
	t.Run("should work, flag enabled with proof", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.ConsensusPropagationChangesFlag
			},
		}

		hdrSigVerifier, _ := NewHeaderSigVerifier(args)

		hdr := &testscommon.HeaderHandlerStub{
			GetPreviousAggregatedSignatureAndBitmapCalled: func() ([]byte, []byte) {
				return []byte("sig"), []byte{1, 1, 1, 1}
			},
		}
		err := hdrSigVerifier.VerifyPreviousBlockProof(hdr)
		assert.Nil(t, err)
	})
}
