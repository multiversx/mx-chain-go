package block

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/parsers"
)

const (
	egldBalanceFilePrefix           = "egld_balance"
	legacyDelegationStateFilePrefix = "legacy_delegation_state"
	directStakeFilePrefix           = "direct_stake"
	delegatedInfoPrefix             = "delegated_info"

	legacyDelegationAddress = "erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt"

	envNodePIInterface = "NODE_API_INTERFACE"
	envS3AccessKey     = "S3_ACCESS_LEY"
	envS3SecretKey     = "S3_SECRET_KEY"
	envS3URL           = "S3_URL"
	envBucketName      = "S3_BUCKET_NAME"

	stakingProvidersActivationEpoch = uint32(239)
)

type accountsDumper struct {
	marshaller       marshal.Marshalizer
	addressConverter core.PubkeyConverter
	restClient       *restClient
	sClient          *s3Client
}

func newAccountsDumper(marshaller marshal.Marshalizer, addressConverter core.PubkeyConverter) (*accountsDumper, error) {
	nodeApiInterface := os.Getenv(envNodePIInterface)
	if nodeApiInterface == "" {
		return nil, errors.New("NODE_API_INTERFACE env empty")
	}

	accessKey := os.Getenv(envS3AccessKey)
	if accessKey == "" {
		return nil, errors.New("S3_ACCESS_LEY env empty")
	}

	secretKey := os.Getenv(envS3SecretKey)
	if secretKey == "" {
		return nil, errors.New("S3_SECRET_KEY env empty")
	}

	url := os.Getenv(envS3URL)
	if url == "" {
		return nil, errors.New("S3_URL env empty")
	}

	bucketName := os.Getenv(envBucketName)
	if bucketName == "" {
		return nil, errors.New("S3_BUCKET_NAME env empty")
	}

	client, err := NewRestClient(nodeApiInterface)
	if err != nil {
		return nil, err
	}
	sClient, err := newS3Client(accessKey, secretKey, bucketName, url)
	if err != nil {
		return nil, err
	}

	return &accountsDumper{
		marshaller:       marshaller,
		addressConverter: addressConverter,
		restClient:       client,
		sClient:          sClient,
	}, nil
}

func (ad *accountsDumper) dumpBalancesInfo(shardID, epoch uint32, accountsAdapter state.AccountsAdapter) error {
	log.Warn("start dumping balances", "epoch", epoch)
	accountsMap, err := ad.getAllAccounts(accountsAdapter)
	if err != nil {
		return err
	}

	// save accounts
	accountsMapBytes, err := json.Marshal(accountsMap)
	if err != nil {
		return err
	}

	err = ad.writeInS3(prepareFileName(egldBalanceFilePrefix, shardID, epoch), accountsMapBytes)
	if err != nil {
		return err
	}

	// save legacy delegation state
	if shardID == 2 {
		legacyDelegationState, errL := ad.getLegacyDelegationState()
		if errL != nil {
			return errL
		}

		err = ad.writeInS3(prepareFileName(legacyDelegationStateFilePrefix, shardID, epoch), legacyDelegationState)
		if err != nil {
			return err
		}
	}

	// save direct stake
	if shardID == core.MetachainShardId {
		directStake, errL := ad.getDirectStakeInfo()
		if errL != nil {
			return errL
		}

		err = ad.writeInS3(prepareFileName(directStakeFilePrefix, shardID, epoch), directStake)
		if err != nil {
			return err
		}
	}

	// save delegated info
	// check also is active
	if shardID == core.MetachainShardId && epoch >= stakingProvidersActivationEpoch {
		delegatedInfo, errL := ad.getDelegatedInfo()
		if errL != nil {
			return errL
		}

		err = ad.writeInS3(prepareFileName(delegatedInfoPrefix, shardID, epoch), delegatedInfo)
		if err != nil {
			return err
		}
	}

	log.Warn("all balances was saved", "epoch", epoch)

	return nil
}

func (ad *accountsDumper) writeInS3(fileName string, fileBytes []byte) error {
	return ad.sClient.Put(fileBytes, fileName)
}

// TODO --- dump legacy delegation state -- SHARD 2 // check if the vm query exists
func (ad *accountsDumper) getLegacyDelegationState() ([]byte, error) {
	endpoint := fmt.Sprintf("/address/%s/keys", legacyDelegationAddress)
	genericResponse := &shared.GenericAPIResponse{}
	err := ad.restClient.CallGetRestEndPoint(endpoint, genericResponse)
	if err != nil {
		return nil, err
	}
	if genericResponse.Error != "" {
		return nil, fmt.Errorf("cannot get delegation legacy state error: %s", genericResponse.Error)
	}

	stateBytes, err := json.Marshal(genericResponse.Data)
	if err != nil {
		return nil, err
	}

	return stateBytes, nil
}

// TODO --- call endpoint /direct-staked-info -- dump result in a file --- META
func (ad *accountsDumper) getDirectStakeInfo() ([]byte, error) {
	endpoint := "/network/direct-staked-info"
	genericResponse := &shared.GenericAPIResponse{}
	err := ad.restClient.CallGetRestEndPoint(endpoint, genericResponse)
	if err != nil {
		return nil, err
	}
	if genericResponse.Error != "" {
		return nil, fmt.Errorf("cannot get direct stake error: %s", genericResponse.Error)
	}

	directStakeBytes, err := json.Marshal(genericResponse.Data)
	if err != nil {
		return nil, err
	}

	return directStakeBytes, nil
}

// TODO --- call endpoint /delegated-info -- dump result in a file --- META
func (ad *accountsDumper) getDelegatedInfo() ([]byte, error) {
	endpoint := "/network/delegated-info"
	genericResponse := &shared.GenericAPIResponse{}
	err := ad.restClient.CallGetRestEndPoint(endpoint, genericResponse)
	if err != nil {
		return nil, err
	}
	if genericResponse.Error != "" {
		return nil, fmt.Errorf("cannot get delegated info error: %s", genericResponse.Error)
	}

	delegatedInfo, err := json.Marshal(genericResponse.Data)
	if err != nil {
		return nil, err
	}

	return delegatedInfo, nil
}

func (ad *accountsDumper) getAllAccounts(accountsAdapter state.AccountsAdapter) (map[string]*alteredAccount.AlteredAccount, error) {
	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	rootHash, err := accountsAdapter.RootHash()
	if err != nil {
		return nil, err
	}
	err = accountsAdapter.GetAllLeaves(leavesChannels, context.Background(), rootHash, parsers.NewMainTrieLeafParser())
	if err != nil {
		return map[string]*alteredAccount.AlteredAccount{}, err
	}

	accountsMap := make(map[string]*alteredAccount.AlteredAccount)
	for leaf := range leavesChannels.LeavesChan {
		userAccount, errUnmarshal := ad.unmarshalUserAccount(leaf.Value())
		if errUnmarshal != nil {
			log.Debug("cannot unmarshal user account. it may be a code leaf", "error", errUnmarshal)
			continue
		}

		encodedAddress, errEncode := ad.addressConverter.Encode(userAccount.GetAddress())
		if errEncode != nil {
			return map[string]*alteredAccount.AlteredAccount{}, errEncode
		}

		accountsMap[encodedAddress] = &alteredAccount.AlteredAccount{
			Address: encodedAddress,
			Balance: userAccount.GetBalance().String(),
			Nonce:   userAccount.GetNonce(),
		}
	}

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return map[string]*alteredAccount.AlteredAccount{}, err
	}

	return accountsMap, nil
}

func (ad *accountsDumper) unmarshalUserAccount(userAccountsBytes []byte) (*accounts.UserAccountData, error) {
	userAccount := &accounts.UserAccountData{}
	err := ad.marshaller.Unmarshal(userAccount, userAccountsBytes)
	if err != nil {
		return nil, err
	}

	return userAccount, nil
}

func prepareFileName(prefix string, shardID, epoch uint32) string {
	return fmt.Sprintf("%s_%d_%d", prefix, shardID, epoch)
}
