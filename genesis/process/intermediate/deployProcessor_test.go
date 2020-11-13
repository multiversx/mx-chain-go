package intermediate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func createMockDeployArg() ArgDeployProcessor {
	return ArgDeployProcessor{
		Executor:       &mock.TxExecutionProcessorStub{},
		PubkeyConv:     mock.NewPubkeyConverterMock(32),
		BlockchainHook: &mock.BlockChainHookHandlerMock{},
		QueryService:   &mock.QueryServiceStub{},
	}
}

func TestNewDeployProcessor_NilExecutorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockDeployArg()
	arg.Executor = nil
	dp, err := NewDeployProcessor(arg)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilTxExecutionProcessor, err)
}

func TestNewDeployProcessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockDeployArg()
	arg.PubkeyConv = nil
	dp, err := NewDeployProcessor(arg)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewDeployProcessor_NilBlockchainHookShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockDeployArg()
	arg.BlockchainHook = nil
	dp, err := NewDeployProcessor(arg)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, process.ErrNilBlockChainHook, err)
}

func TestNewDeployProcessor_NilQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockDeployArg()
	arg.QueryService = nil
	dp, err := NewDeployProcessor(arg)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilQueryService, err)
}

func TestNewDeployProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockDeployArg()
	dp, err := NewDeployProcessor(arg)

	assert.False(t, check.IfNil(dp))
	assert.Nil(t, err)
}

//------- Deploy

func TestDeployProcessor_DeployGetCodeFailsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockDeployArg()
	dp, _ := NewDeployProcessor(arg)
	expectedErr := fmt.Errorf("expected error")
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return "", expectedErr
	}

	scAddresses, err := dp.Deploy(&data.InitialSmartContract{})

	assert.Nil(t, scAddresses)
	assert.Equal(t, expectedErr, err)
}

func TestDeployProcessor_DeployGetNonceFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	arg := createMockDeployArg()
	arg.Executor = &mock.TxExecutionProcessorStub{
		GetNonceCalled: func(senderBytes []byte) (uint64, error) {
			return 0, expectedErr
		},
	}
	dp, _ := NewDeployProcessor(arg)
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return "", nil
	}

	scAddresses, err := dp.Deploy(&data.InitialSmartContract{})

	assert.Nil(t, scAddresses)
	assert.Equal(t, expectedErr, err)
}

func TestDeployProcessor_DeployNewAddressFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	arg := createMockDeployArg()
	arg.BlockchainHook = &mock.BlockChainHookHandlerMock{
		NewAddressCalled: func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	dp, _ := NewDeployProcessor(arg)
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return "", nil
	}

	scAddresses, err := dp.Deploy(&data.InitialSmartContract{})

	assert.Nil(t, scAddresses)
	assert.Equal(t, expectedErr, err)
}

func TestDeployProcessor_DeployShouldWork(t *testing.T) {
	t.Parallel()

	testNonce := uint64(4453)
	testSender := []byte("sender")
	executeCalled := false
	testCode := "code"
	vmType := "0500"
	version := "1.0.0"
	accountExists := false
	arg := createMockDeployArg()
	arg.Executor = &mock.TxExecutionProcessorStub{
		GetNonceCalled: func(senderBytes []byte) (uint64, error) {
			if bytes.Equal(senderBytes, testSender) {
				return testNonce, nil
			}
			assert.Fail(t, "wrong sender")

			return 0, nil
		},
		ExecuteTransactionCalled: func(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error {
			if nonce != testNonce {
				assert.Fail(t, "nonce mismatch")
			}
			if !bytes.Equal(sndAddr, testSender) {
				assert.Fail(t, "sender mismatch")
			}
			if !bytes.Equal(rcvAddress, make([]byte, arg.PubkeyConv.Len())) {
				assert.Fail(t, "receiver mismatch")
			}
			if value.Cmp(zero) != 0 {
				assert.Fail(t, "value should have been 0")
			}
			expectedCode := fmt.Sprintf("%s@%s@0100", testCode, vmType)
			if string(data) != expectedCode {
				assert.Fail(t, "code mismatch")
			}

			executeCalled = true
			return nil
		},
		AccountExistsCalled: func(address []byte) bool {
			result := accountExists
			accountExists = true

			return result
		},
	}
	var scResulting []byte
	arg.BlockchainHook = &mock.BlockChainHookHandlerMock{
		NewAddressCalled: func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
			buff := fmt.Sprintf("%s_%d_%s", string(creatorAddress), creatorNonce, hex.EncodeToString(vmType))
			scResulting = []byte(buff)

			return []byte(buff), nil
		},
	}
	arg.QueryService = &mock.QueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{
				ReturnData: [][]byte{[]byte(version)},
			}, nil
		},
	}
	dp, _ := NewDeployProcessor(arg)
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return testCode, nil
	}

	sc := &data.InitialSmartContract{
		VmType:  vmType,
		Version: version,
	}
	sc.SetOwnerBytes(testSender)

	scAddresses, err := dp.Deploy(sc)

	assert.Nil(t, err)
	assert.True(t, executeCalled)
	assert.Equal(t, 1, len(scAddresses))
	assert.Equal(t, scResulting, scAddresses[0])
}

//------- getSCCodeAsHex

func TestDeployProcessor_GetSCCodeAsHexShouldWork(t *testing.T) {
	t.Parallel()

	value, err := getSCCodeAsHex("inexistent file")
	assert.Equal(t, "", value)
	assert.NotNil(t, err)

	value, err = getSCCodeAsHex("./testdata/dummy.txt")
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString([]byte("test string")), value)
}
