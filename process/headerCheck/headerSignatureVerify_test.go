package headerCheck

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	testscommonStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
)

const defaultChancesSelection = 1

var expectedErr = errors.New("expected error")

func createHeaderSigVerifierArgs() *ArgsHeaderSigVerifier {
	v1, _ := nodesCoordinator.NewValidator([]byte("pubKey1"), 1, defaultChancesSelection)
	v2, _ := nodesCoordinator.NewValidator([]byte("pubKey2"), 1, defaultChancesSelection)
	return &ArgsHeaderSigVerifier{
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
		NodesCoordinator: &shardingMocks.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
				return v1, []nodesCoordinator.Validator{v1, v2}, nil
			},
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{"pubKey1", "pubKey2"}, nil
			},
		},
		MultiSigContainer: cryptoMocks.NewMultiSignerContainerMock(cryptoMocks.NewMultiSigner()),
		SingleSigVerifier: &mock.SignerMock{},
		KeyGen: &mock.SingleSignKeyGenMock{
			PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, err error) {
				return &mock.SingleSignPublicKey{}, nil
			},
		},
		FallbackHeaderValidator: &testscommon.FallBackHeaderValidatorStub{},
		EnableEpochsHandler:     enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		HeadersPool: &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return &dataBlock.Header{
					PrevRandSeed: []byte("prevRandSeed"),
				}, nil
			},
		},
		ProofsPool:     &dataRetrieverMocks.ProofsPoolMock{},
		StorageService: &genericMocks.ChainStorerMock{},
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

func TestNewHeaderSigVerifier_NilMultiSigContainerShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.MultiSigContainer = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilMultiSignerContainer, err)
}

func TestNewHeaderSigVerifier_NilFallbackHeaderValidatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.FallbackHeaderValidator = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilFallbackHeaderValidator, err)
}

func TestNewHeaderSigVerifier_NilHeadersPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.HeadersPool = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilHeadersDataPool, err)
}

func TestNewHeaderSigVerifier_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.ProofsPool = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewHeaderSigVerifier_NilStorageServiceShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	args.StorageService = nil
	hdrSigVerifier, err := NewHeaderSigVerifier(args)

	require.Nil(t, hdrSigVerifier)
	require.Equal(t, process.ErrNilStorageService, err)
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
	header := &dataBlock.Header{
		PrevRandSeed: nil,
		RandSeed:     []byte("rand seed"),
	}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Equal(t, process.ErrNilPrevRandSeed, err)
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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PrevRandSeed: []byte("prev rand seed"),
		RandSeed:     []byte("rand seed"),
	}

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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     []byte("randSeed"),
		PrevRandSeed: []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifyRandSeed(header)
	require.Equal(t, localError, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifyRandSeedAndLeaderSignatureNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     nil,
		PrevRandSeed: []byte("prev rand seed"),
	}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Equal(t, process.ErrNilRandSeed, err)
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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     []byte("randSeed"),
		PrevRandSeed: []byte("prevRandSeed"),
	}

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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:        []byte("randSeed"),
		PrevRandSeed:    []byte("prevRandSeed"),
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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     []byte("randSeed"),
		PrevRandSeed: []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifyRandSeedAndLeaderSignature(header)
	require.Nil(t, err)
	require.Equal(t, 2, count)
}

func TestHeaderSigVerifier_VerifyLeaderSignatureNilPrevRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     []byte("rand seed "),
		PrevRandSeed: nil,
	}

	err := hdrSigVerifier.VerifyLeaderSignature(header)
	require.Equal(t, process.ErrNilPrevRandSeed, err)
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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     []byte("randSeed"),
		PrevRandSeed: []byte("prevRandSeed"),
	}

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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:        []byte("randSeed"),
		PrevRandSeed:    []byte("prevRandSeed"),
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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		RandSeed:     []byte("randSeed"),
		PrevRandSeed: []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifyLeaderSignature(header)
	require.Nil(t, err)
	require.Equal(t, 1, count)
}

func TestHeaderSigVerifier_VerifySignatureNilBitmapShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: nil,
		RandSeed:      []byte("randSeed"),
		PrevRandSeed:  []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestHeaderSigVerifier_VerifySignatureBlockProposerSigMissingShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("0"),
		RandSeed:      []byte("randSeed"),
		PrevRandSeed:  []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, process.ErrBlockProposerSignatureMissing, err)
}

func TestHeaderSigVerifier_VerifySignatureNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PrevRandSeed:  nil,
		PubKeysBitmap: []byte("1"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, process.ErrNilPrevRandSeed, err)
}

func TestHeaderSigVerifier_VerifySignatureWrongSizeBitmapShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
		},
	}
	args.NodesCoordinator = nc

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("11"),
		RandSeed:      []byte("randSeed"),
		PrevRandSeed:  []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, common.ErrWrongSizeBitmap, err)
}

func TestHeaderSigVerifier_VerifySignatureNotEnoughSigsShouldErr(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v, v, v, v, v}, nil
		},
	}
	args.NodesCoordinator = nc

	hdrSigVerifier, _ := NewHeaderSigVerifier(args)
	header := &dataBlock.Header{
		PubKeysBitmap: []byte("A"),
		RandSeed:      []byte("randSeed"),
		PrevRandSeed:  []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, common.ErrNotEnoughSignatures, err)
}

func TestHeaderSigVerifier_VerifySignatureOk(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v}, nil
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
		PrevRandSeed:  []byte("prevRandSeed"),
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
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v, v, v, v, v}, nil
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
		PrevRandSeed:  []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Equal(t, common.ErrNotEnoughSignatures, err)
	require.False(t, wasCalled)
}

func TestHeaderSigVerifier_VerifySignatureOkWhenFallbackThresholdCouldBeApplied(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createHeaderSigVerifierArgs()
	pkAddr := []byte("aaa00000000000000000000000000000")
	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			v, _ := nodesCoordinator.NewValidator(pkAddr, 1, defaultChancesSelection)
			return v, []nodesCoordinator.Validator{v, v, v, v, v}, nil
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
		PubKeysBitmap: []byte{15},
		PrevRandSeed:  []byte("prevRandSeed"),
	}

	err := hdrSigVerifier.VerifySignature(header)
	require.Nil(t, err)
	require.True(t, wasCalled)
}

func TestHeaderSigVerifier_VerifySignatureWithEquivalentProofsActivated(t *testing.T) {
	wasCalled := false
	args := createHeaderSigVerifierArgs()
	numValidatorsConsensusBeforeActivation := 7
	numValidatorsConsensusAfterActivation := 10
	eligibleListSize := numValidatorsConsensusAfterActivation
	eligibleValidatorsKeys := make([]string, eligibleListSize)
	eligibleValidators := make([]nodesCoordinator.Validator, eligibleListSize)
	activationEpoch := uint32(1)

	for i := 0; i < eligibleListSize; i++ {
		eligibleValidatorsKeys[i] = "pubKey" + strconv.Itoa(i)
		eligibleValidators[i], _ = nodesCoordinator.NewValidator([]byte(eligibleValidatorsKeys[i]), 1, defaultChancesSelection)
	}

	nc := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validators []nodesCoordinator.Validator, err error) {
			if epoch < activationEpoch {
				return eligibleValidators[0], eligibleValidators[:numValidatorsConsensusBeforeActivation], nil
			}
			return eligibleValidators[0], eligibleValidators, nil
		},
		GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
			return eligibleValidatorsKeys, nil
		},
	}

	t.Run("check transition block", func(t *testing.T) {
		enableEpochs := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
		args.EnableEpochsHandler = enableEpochs
		enableEpochs.IsFlagEnabledInEpochCalled = func(flag core.EnableEpochFlag, epoch uint32) bool {
			return epoch >= activationEpoch
		}
		enableEpochs.GetActivationEpochCalled = func(flag core.EnableEpochFlag) uint32 {
			return activationEpoch
		}

		args.NodesCoordinator = nc
		args.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(&cryptoMocks.MultisignerMock{
			VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
				wasCalled = true
				return nil
			}})
		hdrSigVerifier, _ := NewHeaderSigVerifier(args)
		header := &dataBlock.HeaderV2{
			Header: &dataBlock.Header{
				ShardID:            0,
				PrevRandSeed:       []byte("prevRandSeed"),
				PubKeysBitmap:      nil,
				Signature:          nil,
				Epoch:              1,
				EpochStartMetaHash: []byte("epoch start meta hash"), // to make this the epoch start block in the shard

			},
		}

		err := hdrSigVerifier.VerifySignature(header)
		require.Nil(t, err)
		require.False(t, wasCalled)

		// check current block proof
		err = hdrSigVerifier.VerifyHeaderProof(&dataBlock.HeaderProof{
			PubKeysBitmap:       []byte{0xff}, // bitmap should still have the old format
			AggregatedSignature: []byte("aggregated signature"),
			HeaderHash:          []byte("hash"),
			HeaderEpoch:         1,
			IsStartOfEpoch:      true,
		})
		require.Nil(t, err)
	})
}

func getFilledHeader() data.HeaderHandler {
	return &dataBlock.Header{
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		PubKeysBitmap:   []byte{0xFF},
		LeaderSignature: []byte("leader signature"),
	}
}

func TestHeaderSigVerifier_VerifyHeaderProof(t *testing.T) {
	t.Parallel()

	t.Run("nil proof should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AndromedaFlag)
		hdrSigVerifier, err := NewHeaderSigVerifier(args)
		require.NoError(t, err)

		err = hdrSigVerifier.VerifyHeaderProof(nil)
		require.Equal(t, process.ErrNilHeaderProof, err)
	})
	t.Run("flag not active should error", func(t *testing.T) {
		t.Parallel()

		hdrSigVerifier, err := NewHeaderSigVerifier(createHeaderSigVerifierArgs())
		require.NoError(t, err)

		err = hdrSigVerifier.VerifyHeaderProof(&dataBlock.HeaderProof{
			PubKeysBitmap: []byte{3},
		})
		require.True(t, errors.Is(err, process.ErrFlagNotActive))
		require.True(t, strings.Contains(err.Error(), string(common.AndromedaFlag)))
	})
	t.Run("GetMultiSigner error should error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		args.MultiSigContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				cnt++
				if cnt > 1 {
					return nil, expectedErr
				}
				return &cryptoMocks.MultiSignerStub{}, nil
			},
		}
		hdrSigVerifier, err := NewHeaderSigVerifier(args)
		require.NoError(t, err)

		err = hdrSigVerifier.VerifyHeaderProof(&dataBlock.HeaderProof{})
		require.Equal(t, expectedErr, err)
	})
	t.Run("getConsensusSignersForEquivalentProofs error should error", func(t *testing.T) {
		t.Parallel()

		headerHash := []byte("header hash")
		wasVerifyAggregatedSigCalled := false
		args := createHeaderSigVerifierArgs()
		args.HeadersPool = &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return getFilledHeader(), nil
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		args.MultiSigContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return &cryptoMocks.MultiSignerStub{
					VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
						wasVerifyAggregatedSigCalled = true
						return nil
					},
				}, nil
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return nil, expectedErr
			},
		}
		hdrSigVerifier, err := NewHeaderSigVerifier(args)
		require.NoError(t, err)

		err = hdrSigVerifier.VerifyHeaderProof(&dataBlock.HeaderProof{
			PubKeysBitmap:       []byte{0x3},
			AggregatedSignature: make([]byte, 10),
			HeaderHash:          headerHash,
		})
		require.Equal(t, expectedErr, err)
		require.False(t, wasVerifyAggregatedSigCalled)
	})
	t.Run("should try multiple times to get header if not available", func(t *testing.T) {
		t.Parallel()

		headerHash := []byte("header hash")
		wasVerifyAggregatedSigCalled := false
		args := createHeaderSigVerifierArgs()

		args.StorageService = &testscommonStorage.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &testscommonStorage.StorerStub{
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("not found")
					},
				}, nil
			},
		}

		numCalls := 0
		args.HeadersPool = &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if numCalls < 2 {
					numCalls++
					return nil, errors.New("not found")
				}

				return getFilledHeader(), nil
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		args.MultiSigContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return &cryptoMocks.MultiSignerStub{
					VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
						wasVerifyAggregatedSigCalled = true
						return nil
					},
				}, nil
			},
		}
		hdrSigVerifier, err := NewHeaderSigVerifier(args)
		require.NoError(t, err)

		err = hdrSigVerifier.VerifyHeaderProof(&dataBlock.HeaderProof{
			PubKeysBitmap:       []byte{0x3},
			AggregatedSignature: make([]byte, 10),
			HeaderHash:          headerHash,
			IsStartOfEpoch:      true,
		})
		require.NoError(t, err)
		require.True(t, wasVerifyAggregatedSigCalled)

		require.Equal(t, 2, numCalls)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		headerHash := []byte("header hash")
		wasVerifyAggregatedSigCalled := false
		args := createHeaderSigVerifierArgs()
		args.HeadersPool = &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return getFilledHeader(), nil
			},
		}
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		args.MultiSigContainer = &cryptoMocks.MultiSignerContainerStub{
			GetMultiSignerCalled: func(epoch uint32) (crypto.MultiSigner, error) {
				return &cryptoMocks.MultiSignerStub{
					VerifyAggregatedSigCalled: func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
						wasVerifyAggregatedSigCalled = true
						return nil
					},
				}, nil
			},
		}
		hdrSigVerifier, err := NewHeaderSigVerifier(args)
		require.NoError(t, err)

		err = hdrSigVerifier.VerifyHeaderProof(&dataBlock.HeaderProof{
			PubKeysBitmap:       []byte{0x3},
			AggregatedSignature: make([]byte, 10),
			HeaderHash:          headerHash,
		})
		require.NoError(t, err)
		require.True(t, wasVerifyAggregatedSigCalled)
	})
}

func TestHeaderSigVerifier_getConsensusSignersForEquivalentProofs(t *testing.T) {
	t.Parallel()

	t.Run("nil proof should error", func(t *testing.T) {
		t.Parallel()

		hdrSigVerifier, _ := NewHeaderSigVerifier(createHeaderSigVerifierArgs())
		require.NotNil(t, hdrSigVerifier)

		signers, err := hdrSigVerifier.getConsensusSignersForEquivalentProofs(nil)
		require.Nil(t, signers)
		require.Equal(t, process.ErrNilHeaderProof, err)
	})
	t.Run("flag not active should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
		}
		hdrSigVerifier, _ := NewHeaderSigVerifier(args)
		require.NotNil(t, hdrSigVerifier)

		signers, err := hdrSigVerifier.getConsensusSignersForEquivalentProofs(&dataBlock.HeaderProof{})
		require.Nil(t, signers)
		require.Equal(t, process.ErrUnexpectedHeaderProof, err)
	})
	t.Run("nodesCoordinator error should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return nil, expectedErr
			},
		}
		hdrSigVerifier, _ := NewHeaderSigVerifier(args)
		require.NotNil(t, hdrSigVerifier)

		signers, err := hdrSigVerifier.getConsensusSignersForEquivalentProofs(&dataBlock.HeaderProof{
			IsStartOfEpoch: true, // for coverage only
			HeaderEpoch:    1,
		})
		require.Nil(t, signers)
		require.Equal(t, expectedErr, err)
	})
	t.Run("invalid consensus bitmap error should error", func(t *testing.T) {
		t.Parallel()

		args := createHeaderSigVerifierArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{"pk1", "pk2", "pk3", "pk4", "pk5", "pk6", "pk7", "pk8"}, nil
			},
		}
		hdrSigVerifier, _ := NewHeaderSigVerifier(args)
		require.NotNil(t, hdrSigVerifier)

		signers, err := hdrSigVerifier.getConsensusSignersForEquivalentProofs(&dataBlock.HeaderProof{
			PubKeysBitmap: []byte{1, 1},
		})
		require.Nil(t, signers)
		require.Equal(t, common.ErrWrongSizeBitmap, err)
	})
}
