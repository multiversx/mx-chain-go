package pagedTrieWalks

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
)

const (
	numPlainPairs          = 300
	numStateTokens         = 40
	numExtraFungibleTokens = 20
	numSFTs                = 3
	sftQuantity            = 5
	baseIssuingCost        = "1000"
)

// issuedTokens holds the tokens issued through the ESDT system smart contract
type issuedTokens struct {
	fungible      string   // with the local mint role
	extraFungible []string // without roles, they make the registry span pages
	nft           string   // one created, its metadata is on the user account
	sfts          []string // one created of each with a quantity, their metadata is on the system account
	nextNonce     uint64
}

func (tokens issuedTokens) all() []string {
	all := append([]string{tokens.fungible, tokens.nft}, tokens.extraFungible...)
	return append(all, tokens.sfts...)
}

func TestChainSimulator_PagedTrieWalks_AccountAndRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startSimulator(t, 1, func(cfg *config.Configs) {
		cfg.SystemSCConfig.ESDTSystemSCConfig.BaseIssuingCost = baseIssuingCost
	})
	defer cs.Close()

	err := cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)

	owner, err := cs.GenerateAndMintWalletAddress(0, big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(100)))
	require.Nil(t, err)
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	tokens := issueTokens(t, cs, owner)
	plainPairs := randomPairs(numPlainPairs, "plain", 24)
	statePairs := fungibleTokenPairs(t, numStateTokens)
	for key, value := range plainPairs {
		statePairs[key] = value
	}
	setPairs(t, cs, owner.Bech32, statePairs)

	keysPath := "/address/" + owner.Bech32 + "/keys"
	pageSizes := randomPageSizes(t, 60)

	t.Run("keys", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		samePagesForBothMethods(t, pageRequest{path: keysPath}, pageSizes, func(req pageRequest) (interface{}, []int) {
			return client.walkPairs(req, "pairs")
		})

		union, _ := client.walkPairs(pageRequest{method: http.MethodGet, path: keysPath}, "pairs")
		for key, value := range statePairs {
			require.Equal(t, value, union[key], "key %s", key)
		}
	})

	t.Run("account with keys", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		req := pageRequest{path: "/address/" + owner.Bech32, query: "withKeys=true"}
		walkAccount := func(req pageRequest) (interface{}, []int) {
			union := make(map[string]string)
			pageSizes := make([]int, 0)
			client.walk(req, func(p page) {
				account := struct {
					Pairs map[string]string `json:"pairs"`
				}{}
				decodeField(t, p, "account", &account)
				pageSizes = append(pageSizes, len(account.Pairs))
				for key, value := range account.Pairs {
					_, duplicate := union[key]
					require.False(t, duplicate, "key %s returned twice", key)
					union[key] = value
				}
			})
			return union, pageSizes
		}
		samePagesForBothMethods(t, req, pageSizes, walkAccount)

		union, _ := walkAccount(req)
		keys, _ := client.walkPairs(pageRequest{method: http.MethodGet, path: keysPath}, "pairs")
		require.Equal(t, keys, union)

		// the account without keys is not paged
		p := client.get("/address/"+owner.Bech32, "")
		require.Equal(t, http.StatusOK, p.status)
		require.Nil(t, p.iteratorState)
		account := struct {
			Pairs map[string]string `json:"pairs"`
		}{}
		decodeField(t, p, "account", &account)
		require.Empty(t, account.Pairs)
	})

	t.Run("esdt", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		req := pageRequest{path: "/address/" + owner.Bech32 + "/esdt"}
		samePagesForBothMethods(t, req, pageSizes, func(req pageRequest) (interface{}, []int) {
			return client.walkESDTs(req)
		})

		union, _ := client.walkESDTs(pageRequest{method: http.MethodGet, path: req.path})
		require.Len(t, union, numStateTokens+len(tokens.all()))
		for i := 0; i < numStateTokens; i++ {
			identifier := stateTokenIdentifier(i)
			require.Equal(t, fmt.Sprintf("%d", i+1), union[identifier].Balance, identifier)
			require.Equal(t, core.FungibleESDT, union[identifier].Type)
		}
		require.Equal(t, core.FungibleESDT, union[tokens.fungible].Type)
		nft := union[tokens.nft+"-01"]
		require.Equal(t, core.NonFungibleESDTv2, nft.Type)
		require.Equal(t, uint64(1), nft.Nonce)
		require.Equal(t, "name", nft.Name) // the metadata of a NonFungibleESDTv2 token is on the user account
		require.Equal(t, owner.Bech32, nft.Creator)
		for _, sftID := range tokens.sfts {
			sft := union[sftID+"-01"]
			require.Equal(t, core.SemiFungibleESDT, sft.Type, sftID)
			require.Equal(t, fmt.Sprintf("%d", sftQuantity), sft.Balance, sftID)
			require.Equal(t, "name", sft.Name, sftID) // read from the system account
			require.Equal(t, []byte("attributes"), sft.Attributes, sftID)
		}
	})

	t.Run("registry", func(t *testing.T) {
		client := newRestClient(t, cs, core.MetachainShardId)
		fungibleTokens := append([]string{tokens.fungible}, tokens.extraFungible...)
		expectedByPath := map[string][]string{
			"/network/esdts":                     tokens.all(),
			"/network/esdt/fungible-tokens":      fungibleTokens,
			"/network/esdt/semi-fungible-tokens": tokens.sfts,
			"/network/esdt/non-fungible-tokens":  {tokens.nft},
			"/address/" + owner.Bech32 + "/esdts-with-role/" + core.ESDTRoleNFTCreate:      append([]string{tokens.nft}, tokens.sfts...),
			"/address/" + owner.Bech32 + "/esdts-with-role/" + core.ESDTRoleLocalMint:      {tokens.fungible},
			"/address/" + owner.Bech32 + "/esdts-with-role/" + core.ESDTRoleNFTAddQuantity: tokens.sfts,
			"/address/" + owner.Bech32 + "/registered-nfts":                                append([]string{tokens.nft}, tokens.sfts...),
		}
		for path, expected := range expectedByPath {
			req := pageRequest{path: path}
			samePagesForBothMethods(t, req, []uint{1, 2}, func(req pageRequest) (interface{}, []int) {
				union, pageSizes := client.walkTokens(req, "tokens")
				return len(union), pageSizes
			})
			union, _ := client.walkTokens(pageRequest{method: http.MethodGet, path: path}, "tokens")
			require.ElementsMatch(t, expected, union, path)
		}

		rolesReq := pageRequest{path: "/address/" + owner.Bech32 + "/esdts/roles"}
		walkRoles := func(req pageRequest) (interface{}, []int) {
			union := make(map[string][]string)
			pageSizes := make([]int, 0)
			client.walk(req, func(p page) {
				roles := make(map[string][]string)
				decodeField(t, p, "roles", &roles)
				pageSizes = append(pageSizes, len(roles))
				for token, tokenRoles := range roles {
					_, duplicate := union[token]
					require.False(t, duplicate, "token %s returned twice", token)
					union[token] = tokenRoles
				}
			})
			return union, pageSizes
		}
		samePagesForBothMethods(t, rolesReq, []uint{1, 2}, walkRoles)
		union, _ := walkRoles(pageRequest{method: http.MethodGet, path: rolesReq.path})
		roles := union.(map[string][]string)
		require.Len(t, roles, 2+numSFTs)
		require.ElementsMatch(t, []string{core.ESDTRoleNFTCreate}, roles[tokens.nft])
		require.ElementsMatch(t, []string{core.ESDTRoleLocalMint}, roles[tokens.fungible])
		for _, sftID := range tokens.sfts {
			require.ElementsMatch(t, []string{core.ESDTRoleNFTCreate, core.ESDTRoleNFTAddQuantity}, roles[sftID])
		}
	})

	t.Run("legacy iterate-keys returns the pairs of keys", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		expected, _ := client.walkPairs(pageRequest{method: http.MethodGet, path: keysPath}, "pairs")

		union := make(map[string]string)
		var state [][]byte
		numPages := 0
		for {
			require.Less(t, numPages, maxPagesPerWalk)
			p := client.post("/address/iterate-keys", map[string]interface{}{"address": owner.Bech32, "numKeys": 7, "iteratorState": state})
			require.Equal(t, http.StatusOK, p.status, p.response.Error)
			numPages++
			pairs := make(map[string]string)
			decodeField(t, p, "pairs", &pairs)
			for key, value := range pairs {
				_, duplicate := union[key]
				require.False(t, duplicate, "key %s returned twice", key)
				union[key] = value
			}
			state = nil
			decodeField(t, p, "newIteratorState", &state)
			if len(state) == 0 {
				break
			}
		}
		require.Greater(t, numPages, 1)
		require.Equal(t, expected, union)
	})

	t.Run("continuation keeps the state of the first page while blocks and epochs advance", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		before, _ := client.walkPairs(pageRequest{method: http.MethodGet, path: keysPath}, "pairs")
		firstPage := client.requestPage(pageRequest{method: http.MethodGet, path: keysPath, numKeys: 5}, nil)
		require.Equal(t, http.StatusOK, firstPage.status)
		require.NotNil(t, firstPage.iteratorState)
		require.NotNil(t, firstPage.iteratorState.Epoch)
		epochOfFirstPage := *firstPage.iteratorState.Epoch
		require.Equal(t, cs.GetNodeHandler(0).GetProcessComponents().EpochStartTrigger().Epoch(), epochOfFirstPage)

		err = cs.GenerateBlocksUntilEpochIsReached(int32(epochOfFirstPage) + 1)
		require.Nil(t, err)
		newPairs := randomPairs(50, "later", 8)
		setPairs(t, cs, owner.Bech32, newPairs)

		union := make(map[string]string)
		decodeField(t, firstPage, "pairs", &union)
		rest, pageSizes := client.walkPairs(pageRequest{method: http.MethodPost, path: keysPath, numKeys: 7, start: firstPage.iteratorState}, "pairs")
		require.Greater(t, len(pageSizes), 1)
		for key, value := range rest {
			_, duplicate := union[key]
			require.False(t, duplicate, "key %s returned twice", key)
			union[key] = value
		}
		require.Equal(t, before, union)

		// the same state is reachable by a fresh request with the block root hash and the epoch of the first page
		historical, _ := client.walkPairs(pageRequest{
			method: http.MethodGet,
			path:   keysPath,
			query:  fmt.Sprintf("blockRootHash=%s&hintEpoch=%d", firstPage.blockInfo.RootHash, epochOfFirstPage),
		}, "pairs")
		require.Equal(t, before, historical)
		historical, _ = client.walkPairs(pageRequest{
			method: http.MethodPost,
			path:   keysPath,
			body:   map[string]interface{}{"blockRootHash": firstPage.blockInfo.RootHash, "hintEpoch": epochOfFirstPage},
		}, "pairs")
		require.Equal(t, before, historical)

		// while the current state holds the new keys as well
		current, _ := client.walkPairs(pageRequest{method: http.MethodGet, path: keysPath}, "pairs")
		require.Equal(t, len(before)+len(newPairs), len(current))
	})

	t.Run("continuation keeps the system account of the first page", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		path := "/address/" + owner.Bech32 + "/esdt"

		// the metadata of a SFT lives in the system account; a dynamic token can have it updated
		for _, sftID := range tokens.sfts {
			tx := vm.ChangeToDynamicTx(tokens.nextNonce, owner.Bytes, []byte(sftID))
			tokens.nextNonce++
			txResult, errTx := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
			require.Nil(t, errTx)
			require.Equal(t, "success", txResult.Status.String())
			vm.SetAddressEsdtRoles(t, cs, tokens.nextNonce, owner, []byte(sftID), [][]byte{[]byte(core.ESDTRoleNFTUpdate)})
			tokens.nextNonce++
		}
		err = cs.GenerateBlocks(2) // the roles reach the shard through smart contract results
		require.Nil(t, err)

		before, _ := client.walkESDTs(pageRequest{method: http.MethodGet, path: path})
		firstPage := client.requestPage(pageRequest{method: http.MethodGet, path: path, numKeys: 1}, nil)
		require.Equal(t, http.StatusOK, firstPage.status)
		require.NotNil(t, firstPage.iteratorState)
		firstTokens := make(map[string]groups.ESDTNFTTokenData)
		decodeField(t, firstPage, "esdts", &firstTokens)

		updated := ""
		for _, sftID := range tokens.sfts {
			if _, found := firstTokens[sftID+"-01"]; !found {
				updated = sftID
				break
			}
		}
		if updated == "" {
			t.Skip("every SFT was returned by the first page, nothing left to update")
		}
		newMetaData := txsFee.GetDefaultMetaData()
		newMetaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
		newMetaData.Name = []byte(hex.EncodeToString([]byte("name2")))
		newMetaData.Attributes = []byte(hex.EncodeToString([]byte("attributes2")))
		tx := vm.EsdtMetaDataUpdateTx([]byte(updated), newMetaData, tokens.nextNonce, owner.Bytes)
		tokens.nextNonce++
		txResult, errTx := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, errTx)
		require.Equal(t, "success", txResult.Status.String())

		rest, _ := client.walkESDTs(pageRequest{method: http.MethodGet, path: path, numKeys: 1, start: firstPage.iteratorState})
		for identifier, token := range rest {
			_, duplicate := firstTokens[identifier]
			require.False(t, duplicate, "token %s returned twice", identifier)
			firstTokens[identifier] = token
		}
		require.Equal(t, before, firstTokens)
		require.Equal(t, "name", firstTokens[updated+"-01"].Name)
		require.Equal(t, []byte("attributes"), firstTokens[updated+"-01"].Attributes)

		current, _ := client.walkESDTs(pageRequest{method: http.MethodGet, path: path})
		require.Equal(t, "name2", current[updated+"-01"].Name)
		require.Equal(t, []byte("attributes2"), current[updated+"-01"].Attributes)
	})

	t.Run("continuation after the data trie is emptied and the account removed", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		wallet := newWallet(t, cs, 0)
		walletPairs := randomPairs(60, "emptied", 8)
		setPairs(t, cs, wallet, walletPairs)
		path := "/address/" + wallet + "/keys"

		firstPage := client.requestPage(pageRequest{method: http.MethodGet, path: path, numKeys: 10}, nil)
		require.Equal(t, http.StatusOK, firstPage.status)
		require.NotNil(t, firstPage.iteratorState)

		deletions := make(map[string]string, len(walletPairs))
		for key := range walletPairs {
			deletions[key] = ""
		}
		setPairs(t, cs, wallet, deletions)

		firstPairs := make(map[string]string)
		decodeField(t, firstPage, "pairs", &firstPairs)
		union := make(map[string]string)
		for key, value := range firstPairs {
			union[key] = value
		}
		rest, _ := client.walkPairs(pageRequest{method: http.MethodGet, path: path, numKeys: 10, start: firstPage.iteratorState}, "pairs")
		for key, value := range rest {
			_, duplicate := union[key]
			require.False(t, duplicate, "key %s returned twice", key)
			union[key] = value
		}
		require.Equal(t, walletPairs, union)

		requireEmptyListing := func() {
			current := client.get(path, "")
			require.Equal(t, http.StatusOK, current.status, current.response.Error)
			require.Nil(t, current.iteratorState)
			pairs := make(map[string]string)
			decodeField(t, current, "pairs", &pairs)
			require.Empty(t, pairs)
		}
		requireEmptyListing()

		err = cs.RemoveAccounts([]string{wallet})
		require.Nil(t, err)
		err = cs.GenerateBlocks(1)
		require.Nil(t, err)
		rest, _ = client.walkPairs(pageRequest{method: http.MethodPost, path: path, numKeys: 10, start: firstPage.iteratorState}, "pairs")
		require.Equal(t, len(walletPairs)-len(firstPairs), len(rest))
		requireEmptyListing()
	})

	t.Run("invalid requests", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		firstPage := client.get(keysPath, "numKeys=2")
		require.Equal(t, http.StatusOK, firstPage.status)
		require.NotNil(t, firstPage.iteratorState)

		p := client.get(keysPath, "iteratorState=***")
		require.Equal(t, http.StatusBadRequest, p.status)
		require.Contains(t, p.response.Error, checkpoint.ErrInvalidCheckpoint.Error())

		p = client.get(keysPath, "iteratorState="+base64.RawURLEncoding.EncodeToString([]byte(`{"blockRootHash":"zz"}`)))
		require.Equal(t, http.StatusBadRequest, p.status)

		p = client.get(keysPath, "numKeys=-1")
		require.Equal(t, http.StatusBadRequest, p.status)

		p = client.do(http.MethodPost, keysPath, []byte(`{"numKeys":`))
		require.Equal(t, http.StatusBadRequest, p.status)

		p = client.post(keysPath, map[string]interface{}{"iteratorState": map[string]interface{}{"blockRootHash": "zz"}})
		require.Equal(t, http.StatusBadRequest, p.status)

		unknownNode := make([]byte, 32)
		_, _ = rand.Read(unknownNode)
		unavailable := *firstPage.iteratorState
		unavailable.IteratorState = [][]byte{unknownNode}
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			p = client.requestPage(pageRequest{method: method, path: keysPath}, &unavailable)
			require.Equal(t, http.StatusBadRequest, p.status, method)
			require.Contains(t, p.response.Error, common.ErrIteratorNodeUnavailable.Error())
		}

		unknownRoot := *firstPage.iteratorState
		unknownRoot.BlockRootHash = hex.EncodeToString(unknownNode)
		p = client.requestPage(pageRequest{method: http.MethodGet, path: keysPath}, &unknownRoot)
		require.Equal(t, http.StatusBadRequest, p.status)
		require.Contains(t, p.response.Error, common.ErrIteratorNodeUnavailable.Error())
	})
}

// walkESDTs pages the esdt endpoint and returns the union of the tokens and the number of tokens of every page
func (rc *restClient) walkESDTs(req pageRequest) (map[string]groups.ESDTNFTTokenData, []int) {
	union := make(map[string]groups.ESDTNFTTokenData)
	pageSizes := make([]int, 0)
	rc.walk(req, func(p page) {
		esdts := make(map[string]groups.ESDTNFTTokenData)
		decodeField(rc.t, p, "esdts", &esdts)
		pageSizes = append(pageSizes, len(esdts))
		for identifier, token := range esdts {
			_, duplicate := union[identifier]
			require.False(rc.t, duplicate, "token %s returned twice", identifier)
			union[identifier] = token
		}
	})

	return union, pageSizes
}

func issueTokens(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, owner dtos.WalletAddress) issuedTokens {
	nonce := uint64(0)
	tokens := issuedTokens{}

	tokens.fungible = sendIssueTx(t, cs, vm.IssueTx(nonce, owner.Bytes, []byte("FUNTICKER"), baseIssuingCost))
	nonce++
	vm.SetAddressEsdtRoles(t, cs, nonce, owner, []byte(tokens.fungible), [][]byte{[]byte(core.ESDTRoleLocalMint)})
	nonce++

	for i := 0; i < numExtraFungibleTokens; i++ {
		ticker := []byte(fmt.Sprintf("EXTRA%02d", i))
		tokens.extraFungible = append(tokens.extraFungible, sendIssueTx(t, cs, vm.IssueTx(nonce, owner.Bytes, ticker, baseIssuingCost)))
		nonce++
	}

	tokens.nft = sendIssueTx(t, cs, vm.IssueNonFungibleTx(nonce, owner.Bytes, []byte("NFTTICKER"), baseIssuingCost))
	nonce++
	vm.SetAddressEsdtRoles(t, cs, nonce, owner, []byte(tokens.nft), [][]byte{[]byte(core.ESDTRoleNFTCreate)})
	nonce++

	sftRoles := [][]byte{[]byte(core.ESDTRoleNFTCreate), []byte(core.ESDTRoleNFTAddQuantity)}
	for i := 0; i < numSFTs; i++ {
		ticker := []byte(fmt.Sprintf("SFTTICKER%d", i))
		sftID := sendIssueTx(t, cs, vm.IssueSemiFungibleTx(nonce, owner.Bytes, ticker, baseIssuingCost))
		nonce++
		vm.SetAddressEsdtRoles(t, cs, nonce, owner, []byte(sftID), sftRoles)
		nonce++
		tokens.sfts = append(tokens.sfts, sftID)
	}

	metaData := txsFee.GetDefaultMetaData()
	metaData.Nonce = []byte(hex.EncodeToString(big.NewInt(1).Bytes()))
	creations := []struct {
		token    string
		quantity int64
	}{{tokens.nft, 1}}
	for _, sftID := range tokens.sfts {
		creations = append(creations, struct {
			token    string
			quantity int64
		}{sftID, sftQuantity})
	}
	for _, creation := range creations {
		tx := vm.EsdtNftCreateTx(nonce, owner.Bytes, []byte(creation.token), metaData, creation.quantity)
		nonce++
		txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
		require.Nil(t, err)
		require.Equal(t, "success", txResult.Status.String())
	}

	tokens.nextNonce = nonce
	return tokens
}

func sendIssueTx(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, tx *transaction.Transaction) string {
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, vm.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, "success", txResult.Status.String())

	return string(txResult.Logs.Events[0].Topics[0])
}

func stateTokenIdentifier(index int) string {
	return fmt.Sprintf("STATE%02d-abcdef", index)
}

// fungibleTokenPairs makes the data trie entries of fungible tokens, the token with index i holds i+1 units
func fungibleTokenPairs(t *testing.T, numTokens int) map[string]string {
	marshaller := &marshal.GogoProtoMarshalizer{}
	pairs := make(map[string]string, numTokens)
	for i := 0; i < numTokens; i++ {
		token := &esdt.ESDigitalToken{Value: big.NewInt(int64(i + 1)), Type: uint32(core.Fungible)}
		value, err := marshaller.Marshal(token)
		require.Nil(t, err)
		key := []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier + stateTokenIdentifier(i))
		pairs[hex.EncodeToString(key)] = hex.EncodeToString(value)
	}

	return pairs
}
