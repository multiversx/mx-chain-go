package factory

import (
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/state"
)

func createArgs() ArgsEpochStartInterceptorContainer {
	pubKeyConv := &testscommon.PubkeyConverterMock{}
	return ArgsEpochStartInterceptorContainer{
		CoreComponents: &mock.CoreComponentsMock{
			AddrPubKeyConv: pubKeyConv,
		},
		CryptoComponents:        &mock.CryptoComponentsMock{},
		Config:                  config.Config{},
		ShardCoordinator:        &mock.ShardCoordinatorStub{},
		MainMessenger:           &processMock.TopicHandlerStub{},
		FullArchiveMessenger:    &processMock.TopicHandlerStub{},
		DataPool:                &dataRetriever.PoolsHolderStub{},
		WhiteListHandler:        &testscommon.WhiteListHandlerStub{},
		WhiteListerVerifiedTxs:  &testscommon.WhiteListHandlerStub{},
		AddressPubkeyConv:       pubKeyConv,
		NonceConverter:          &testscommon.Uint64ByteSliceConverterStub{},
		ChainID:                 []byte{1},
		ArgumentsParser:         &testscommon.ArgumentParserMock{},
		HeaderIntegrityVerifier: &processMock.HeaderIntegrityVerifierStub{},
		RequestHandler:          &testscommon.RequestHandlerStub{},
		SignaturesHandler:       &processMock.SignaturesHandlerStub{},
		NodeOperationMode:       "normal",
		AccountFactory:          &state.AccountsFactoryStub{},
	}
}

func TestCreateEpochStartContainerFactoryArgs(t *testing.T) {
	args := createArgs()

	accMock := &state.AccountWrapMock{
		Address: []byte("addr"),
	}
	accFactory := &state.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return accMock, nil
		},
	}

	args.AccountFactory = accFactory
	createdArgs, err := CreateEpochStartContainerFactoryArgs(args)
	require.Nil(t, err)
	require.NotNil(t, createdArgs)

	acc, err := createdArgs.Accounts.LoadAccount([]byte("addr"))
	require.Nil(t, err)
	require.Equal(t, accMock, acc)
}
