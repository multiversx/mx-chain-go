package intermediate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewDeployProcessor_NilExecutorShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDeployProcessor(
		nil,
		mock.NewPubkeyConverterMock(32),
		&mock.BlockChainHookHandlerMock{},
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilTxExecutionProcessor, err)
}

func TestNewDeployProcessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{},
		nil,
		&mock.BlockChainHookHandlerMock{},
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewDeployProcessor_NilBlockchainHookShouldErr(t *testing.T) {
	t.Parallel()

	dp, err := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{},
		mock.NewPubkeyConverterMock(32),
		nil,
	)

	assert.True(t, check.IfNil(dp))
	assert.Equal(t, process.ErrNilBlockChainHook, err)
}

func TestNewDeployProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	dp, err := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{},
		mock.NewPubkeyConverterMock(32),
		&mock.BlockChainHookHandlerMock{},
	)

	assert.False(t, check.IfNil(dp))
	assert.Nil(t, err)
}

//------- Deploy

func TestDeployProcessor_DeployGetCodeFailsShouldErr(t *testing.T) {
	t.Parallel()

	dp, _ := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{},
		mock.NewPubkeyConverterMock(0),
		&mock.BlockChainHookHandlerMock{},
	)
	expectedErr := fmt.Errorf("expected error")
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return "", expectedErr
	}

	err := dp.Deploy(&data.InitialSmartContract{})

	assert.Equal(t, expectedErr, err)
}

func TestDeployProcessor_DeployGetNonceFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	dp, _ := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{
			GetNonceCalled: func(senderBytes []byte) (uint64, error) {
				return 0, expectedErr
			},
		},
		mock.NewPubkeyConverterMock(0),
		&mock.BlockChainHookHandlerMock{},
	)
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return "", nil
	}

	err := dp.Deploy(&data.InitialSmartContract{})

	assert.Equal(t, expectedErr, err)
}

func TestDeployProcessor_DeployNewAddressFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("expected error")
	dp, _ := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{},
		mock.NewPubkeyConverterMock(0),
		&mock.BlockChainHookHandlerMock{
			NewAddressCalled: func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
				return nil, expectedErr
			},
		},
	)
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return "", nil
	}

	err := dp.Deploy(&data.InitialSmartContract{})

	assert.Equal(t, expectedErr, err)
}

func TestDeployProcessor_DeployShouldWork(t *testing.T) {
	t.Parallel()

	testNonce := uint64(4453)
	testSender := []byte("sender")
	lenAddress := 32
	executeCalled := false
	testCode := "code"
	vmType := "0500"
	accountExists := false
	dp, _ := NewDeployProcessor(
		&mock.TxExecutionProcessorStub{
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
				if !bytes.Equal(rcvAddress, make([]byte, lenAddress)) {
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
		},
		mock.NewPubkeyConverterMock(lenAddress),
		&mock.BlockChainHookHandlerMock{
			NewAddressCalled: func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
				buff := fmt.Sprintf("%s_%d_%s", string(creatorAddress), creatorNonce, hex.EncodeToString(vmType))

				return []byte(buff), nil
			},
		},
	)
	dp.getScCodeAsHex = func(filename string) (string, error) {
		return testCode, nil
	}

	sc := &data.InitialSmartContract{
		VmType: vmType,
	}
	sc.SetOwnerBytes(testSender)

	err := dp.Deploy(sc)

	assert.Nil(t, err)
	assert.True(t, executeCalled)
}

//------- getSCCodeAsHex

func TestDeployProcessor_GetSCCodeAsHexShouldWork(t *testing.T) {
	t.Parallel()

	dp := &deployProcessor{}

	value, err := dp.getSCCodeAsHex("inexistent file")
	assert.Equal(t, "", value)
	assert.NotNil(t, err)

	value, err = dp.getSCCodeAsHex("./testdata/dummy.txt")
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString([]byte("test string")), value)
}
