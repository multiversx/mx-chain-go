package parsing_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInitialSmartContract(owner string) *data.InitialSmartContract {
	return &data.InitialSmartContract{
		Owner:    owner,
		Filename: "dummy.wasm",
		VmType:   "00",
	}
}

func TestNewSmartContractsParser_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewSmartContractsParser(
		"./testdata/smartcontracts_ok.json",
		nil,
		&mock.KeyGeneratorStub{},
	)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewSmartContractsParser_NilKeyGeneratorShouldErr(t *testing.T) {
	t.Parallel()

	ap, err := parsing.NewSmartContractsParser(
		"./testdata/smartcontracts_ok.json",
		createMockHexPubkeyConverter(),
		nil,
	)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilKeyGenerator, err)
}

func TestNewSmartContractsParser_BadFilenameShouldEr(t *testing.T) {
	t.Parallel()

	scp, err := parsing.NewSmartContractsParser(
		"inexistent file",
		createMockHexPubkeyConverter(),
		&mock.KeyGeneratorStub{},
	)

	assert.True(t, check.IfNil(scp))
	assert.NotNil(t, err)
}

func TestNewSmartContractsParser_BadJsonShouldErr(t *testing.T) {
	t.Parallel()

	scp, err := parsing.NewSmartContractsParser(
		"testdata/smartcontracts_bad.json",
		createMockHexPubkeyConverter(),
		&mock.KeyGeneratorStub{},
	)

	assert.True(t, check.IfNil(scp))
	assert.NotNil(t, err)
}

func TestNewSmartContractsParser_ShouldWork(t *testing.T) {
	t.Parallel()

	scp, err := parsing.NewSmartContractsParser(
		"testdata/smartcontracts_ok.json",
		createMockHexPubkeyConverter(),
		&mock.KeyGeneratorStub{},
	)

	assert.False(t, check.IfNil(scp))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(scp.InitialSmartContracts()))
}

//------- process

func TestSmartContractsParser_ProcessEmptyOwnerAddressShouldErr(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	isc := createMockInitialSmartContract("0001")
	isc.Owner = ""
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc})

	err := scp.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyOwnerAddress))
}

func TestSmartContractsParser_ProcessInvalidOwnerAddressShouldErr(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	isc := createMockInitialSmartContract("0001")
	isc.Owner = "invalid address"
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc})

	err := scp.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidOwnerAddress))
}

func TestSmartContractsParser_ProcessInvalidOwnerPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	scp.SetKeyGenerator(&mock.KeyGeneratorStub{
		CheckPublicKeyValidCalled: func(b []byte) error {
			return expectedErr
		},
	})
	isc := createMockInitialSmartContract("0001")
	isc.Owner = "00"
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc})

	err := scp.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidPubKey))
}

func TestSmartContractsParser_ProcessEmptyVmTypeShouldErr(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	isc := createMockInitialSmartContract("0001")
	isc.VmType = ""
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc})

	err := scp.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyVmType))
}

func TestSmartContractsParser_ProcessInvalidVmTypeShouldErr(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	isc := createMockInitialSmartContract("0001")
	isc.VmType = "invalid VM type"
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc})

	err := scp.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidVmType))
}

func TestSmartContractsParser_FileErrorShouldErr(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	expectedErr := errors.New("expected error")
	scp.SetFileHandler(
		func(s string) error {
			return expectedErr
		},
	)
	isc := createMockInitialSmartContract("0001")
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc})

	err := scp.Process()

	assert.True(t, errors.Is(err, expectedErr))
}

//------- InitialSmartContractsSplitOnOwnersShards

func TestSmartContractsParser_InitialSmartContractsSplitOnOwnersShardsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	m, err := scp.InitialSmartContractsSplitOnOwnersShards(nil)

	assert.Nil(t, m)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestSmartContractsParser_InitialSmartContractsSplitOnOwnersShards(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	ics := []*data.InitialSmartContract{
		createMockInitialSmartContract("0001"),
		createMockInitialSmartContract("0002"),
		createMockInitialSmartContract("0000"),
		createMockInitialSmartContract("0101"),
	}

	scp.SetInitialSmartContracts(ics)
	err := scp.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	icsSplit, err := scp.InitialSmartContractsSplitOnOwnersShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(icsSplit))
	assert.Equal(t, 2, len(icsSplit[1]))
}

func TestSmartContractsParser_InitialSmartContractsNoNeedtoSplit(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	ics := []*data.InitialSmartContract{
		{
			Owner:    "0001",
			Filename: "dummy.wasm",
			VmType:   "00",
			Type:     genesis.DNSType,
		},
	}

	scp.SetInitialSmartContracts(ics)
	err := scp.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	icsSplit, err := scp.InitialSmartContractsSplitOnOwnersShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(icsSplit))
	for i := uint32(0); i < threeSharder.NumOfShards; i++ {
		assert.Equal(t, 1, len(icsSplit[i]))
	}
}

func TestSmartContractsParser_ProcessDNSTwiceShouldError(t *testing.T) {
	t.Parallel()

	scp := parsing.NewTestSmartContractsParser(createMockHexPubkeyConverter())
	isc1 := &data.InitialSmartContract{
		Owner:    "0001",
		Filename: "dummy.wasm",
		VmType:   "00",
		Type:     genesis.DNSType,
	}
	isc2 := &data.InitialSmartContract{
		Owner:    "0002",
		Filename: "dummy.wasm",
		VmType:   "00",
		Type:     genesis.DNSType,
	}
	scp.SetInitialSmartContracts([]*data.InitialSmartContract{isc1, isc2})

	err := scp.Process()

	assert.True(t, errors.Is(err, genesis.ErrTooManyDNSContracts))
}
