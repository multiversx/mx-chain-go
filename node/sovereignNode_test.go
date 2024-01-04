package node_test

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignNode(t *testing.T) {
	n, err := node.NewNode()
	require.Nil(t, err)
	require.NotNil(t, n)

	sn, err := node.NewSovereignNode(n)
	require.Nil(t, err)
	require.NotNil(t, sn)
}

func TestNewSovereignNode_NilNodeShouldError(t *testing.T) {
	t.Parallel()

	_, err := node.NewSovereignNode(nil)
	assert.NotNil(t, err)
}

func TestSovereignNode_GetAllIssuedESDTs(t *testing.T) {
	t.Parallel()

	acc := createAcc([]byte("newaddress"))
	esdtToken := []byte("TCK-RANDOM")
	sftToken := []byte("SFT-RANDOM")
	nftToken := []byte("NFT-RANDOM")

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT)}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.SaveKeyValue(esdtToken, marshalledData)

	sftData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("semi fungible"), TokenType: []byte(core.SemiFungibleESDT)}
	sftMarshalledData, _ := getMarshalizer().Marshal(sftData)
	_ = acc.SaveKeyValue(sftToken, sftMarshalledData)

	nftData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("non fungible"), TokenType: []byte(core.NonFungibleESDT)}
	nftMarshalledData, _ := getMarshalizer().Marshal(nftData)
	_ = acc.SaveKeyValue(nftToken, nftMarshalledData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)
	nftSuffix := append(nftToken, acc.AddressBytes()...)
	sftSuffix := append(sftToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf

					trieLeaf, _ = tlp.ParseLeaf(sftToken, append(sftMarshalledData, sftSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf

					trieLeaf, _ = tlp.ParseLeaf(nftToken, append(nftMarshalledData, nftSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}

	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: core.MetachainShardId,
	}

	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)
	sn, _ := node.NewSovereignNode(n)

	value, err := sn.GetAllIssuedESDTs(core.FungibleESDT, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, string(esdtToken), value[0])

	value, err = sn.GetAllIssuedESDTs(core.SemiFungibleESDT, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, string(sftToken), value[0])

	value, err = sn.GetAllIssuedESDTs(core.NonFungibleESDT, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, string(nftToken), value[0])

	value, err = sn.GetAllIssuedESDTs("", context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 3, len(value))
}

func TestSovereignNode_GetNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)
	esdtToken := []byte("TCK-RANDOM")

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.SemiFungibleESDT), OwnerAddress: addrBytes}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.SaveKeyValue(esdtToken, marshalledData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		},
	)

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	dataComponents := getDefaultDataComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: core.MetachainShardId,
	}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)
	sn, _ := node.NewSovereignNode(n)

	tokenResult, _, err := sn.GetNFTTokenIDsRegisteredByAddress(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, len(tokenResult))
	require.Equal(t, string(esdtToken), tokenResult[0])
}

func TestSovereignNode_GetESDTsWithRole(t *testing.T) {
	t.Parallel()

	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)
	esdtToken := []byte("TCK-RANDOM")

	specialRoles := []*systemSmartContracts.ESDTRoles{
		{
			Address: addrBytes,
			Roles:   [][]byte{[]byte(core.ESDTRoleNFTAddQuantity), []byte(core.ESDTRoleLocalMint)},
		},
	}

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT), SpecialRoles: specialRoles}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)
	_ = acc.SaveKeyValue(esdtToken, marshalledData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	dataComponents := getDefaultDataComponents()
	stateComponents := getDefaultStateComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: core.MetachainShardId,
	}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)
	sn, _ := node.NewSovereignNode(n)

	tokenResult, _, err := sn.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleNFTAddQuantity, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, len(tokenResult))
	require.Equal(t, string(esdtToken), tokenResult[0])

	tokenResult, _, err = sn.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleLocalMint, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, len(tokenResult))
	require.Equal(t, string(esdtToken), tokenResult[0])

	tokenResult, _, err = sn.GetESDTsWithRole(testscommon.TestAddressAlice, core.ESDTRoleNFTCreate, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Len(t, tokenResult, 0)
}

func TestSovereignNode_GetESDTsRoles(t *testing.T) {
	t.Parallel()

	addrBytes := testscommon.TestPubKeyAlice
	acc := createAcc(addrBytes)
	esdtToken := []byte("TCK-RANDOM")

	specialRoles := []*systemSmartContracts.ESDTRoles{
		{
			Address: addrBytes,
			Roles:   [][]byte{[]byte(core.ESDTRoleNFTAddQuantity), []byte(core.ESDTRoleLocalMint)},
		},
	}

	esdtData := &systemSmartContracts.ESDTDataV2{TokenName: []byte("fungible"), TokenType: []byte(core.FungibleESDT), SpecialRoles: specialRoles}
	marshalledData, _ := getMarshalizer().Marshal(esdtData)

	esdtSuffix := append(esdtToken, acc.AddressBytes()...)

	acc.SetDataTrie(
		&trieMock.TrieStub{
			GetAllLeavesOnChannelCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.KeyBuilder, tlp common.TrieLeafParser) error {
				go func() {
					trieLeaf, _ := tlp.ParseLeaf(esdtToken, append(marshalledData, esdtSuffix...), core.NotSpecified)
					leavesChannels.LeavesChan <- trieLeaf
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
				}()

				return nil
			},
			RootCalled: func() ([]byte, error) {
				return nil, nil
			},
		})

	accDB := &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return nil
		},
	}
	accDB.GetAccountWithBlockInfoCalled = func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
		return acc, nil, nil
	}
	coreComponents := getDefaultCoreComponents()
	stateComponents := getDefaultStateComponents()
	dataComponents := getDefaultDataComponents()
	args := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      accDB,
		CurrentStateAccountsWrapper:    accDB,
		HistoricalStateAccountsWrapper: accDB,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(args)
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{
		SelfShardId: core.MetachainShardId,
	}
	n, _ := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithProcessComponents(processComponents),
	)
	sn, _ := node.NewSovereignNode(n)

	tokenResult, _, err := sn.GetESDTsRoles(testscommon.TestAddressAlice, api.AccountQueryOptions{}, context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string][]string{
		string(esdtToken): {core.ESDTRoleNFTAddQuantity, core.ESDTRoleLocalMint},
	}, tokenResult)
}
