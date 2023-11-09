package parsing

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

// smartContractParser hold data for initial smart contracts
type smartContractParser struct {
	initialSmartContracts []*data.InitialSmartContract
	pubkeyConverter       core.PubkeyConverter
	checkForFileHandler   func(filename string) error
	keyGenerator          crypto.KeyGenerator
}

// NewSmartContractsParser creates a new decoded smart contracts genesis structure from json config file
func NewSmartContractsParser(
	genesisFilePath string,
	pubkeyConverter core.PubkeyConverter,
	keyGenerator crypto.KeyGenerator,
) (*smartContractParser, error) {

	if check.IfNil(pubkeyConverter) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if check.IfNil(keyGenerator) {
		return nil, genesis.ErrNilKeyGenerator
	}

	initialSmartContracts := make([]*data.InitialSmartContract, 0)
	err := core.LoadJsonFile(&initialSmartContracts, genesisFilePath)
	if err != nil {
		return nil, err
	}

	scp := &smartContractParser{
		initialSmartContracts: initialSmartContracts,
		pubkeyConverter:       pubkeyConverter,
		keyGenerator:          keyGenerator,
	}
	scp.checkForFileHandler = scp.checkForFile

	err = scp.process()
	if err != nil {
		return nil, err
	}

	return scp, nil
}

func (scp *smartContractParser) process() error {
	numDNSContract := 0

	for _, initialSmartContract := range scp.initialSmartContracts {
		err := scp.parseElement(initialSmartContract)
		if err != nil {
			return err
		}

		err = scp.checkForFileHandler(initialSmartContract.Filename)
		if err != nil {
			return err
		}

		if initialSmartContract.Type == genesis.DNSType {
			numDNSContract++
		}
	}

	if numDNSContract > 1 {
		return genesis.ErrTooManyDNSContracts
	}

	return nil
}

func (scp *smartContractParser) parseElement(initialSmartContract *data.InitialSmartContract) error {
	if len(initialSmartContract.Owner) == 0 {
		return genesis.ErrEmptyOwnerAddress
	}
	ownerBytes, err := scp.pubkeyConverter.Decode(initialSmartContract.Owner)
	if err != nil {
		return fmt.Errorf("%w for `%s`",
			genesis.ErrInvalidOwnerAddress, initialSmartContract.Owner)
	}

	err = scp.keyGenerator.CheckPublicKeyValid(ownerBytes)
	if err != nil {
		return fmt.Errorf("%w for owner `%s`, error: %s",
			genesis.ErrInvalidPubKey,
			initialSmartContract.Owner,
			err.Error(),
		)
	}

	initialSmartContract.SetOwnerBytes(ownerBytes)

	if len(initialSmartContract.VmType) == 0 {
		return fmt.Errorf("%w for  %s",
			genesis.ErrEmptyVmType, initialSmartContract.Owner)
	}

	vmTypeBytes, err := hex.DecodeString(initialSmartContract.VmType)
	if err != nil {
		return fmt.Errorf("%w for provided %s, error: %s",
			genesis.ErrInvalidVmType, initialSmartContract.VmType, err.Error())
	}

	initialSmartContract.SetVmTypeBytes(vmTypeBytes)

	return nil
}

func (scp *smartContractParser) checkForFile(filename string) error {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w for the file %s", err, filename)
	}

	if info.IsDir() {
		return fmt.Errorf("%w for the file %s", genesis.ErrFilenameIsDirectory, filename)
	}

	return nil
}

// InitialSmartContracts return the initial smart contracts contained by this parser
func (scp *smartContractParser) InitialSmartContracts() []genesis.InitialSmartContractHandler {
	smartContracts := make([]genesis.InitialSmartContractHandler, len(scp.initialSmartContracts))

	for idx, isc := range scp.initialSmartContracts {
		smartContracts[idx] = isc
	}

	return smartContracts
}

// InitialSmartContractsSplitOnOwnersShards returns the initial smart contracts split by the owner shards
func (scp *smartContractParser) InitialSmartContractsSplitOnOwnersShards(
	shardCoordinator sharding.Coordinator,
) (map[uint32][]genesis.InitialSmartContractHandler, error) {

	if check.IfNil(shardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}

	var smartContracts = make(map[uint32][]genesis.InitialSmartContractHandler)
	for _, isc := range scp.initialSmartContracts {
		if isc.Type == genesis.DNSType {
			for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
				smartContracts[i] = append(smartContracts[i], isc)
			}
			continue
		}

		shardID := shardCoordinator.ComputeId(isc.OwnerBytes())
		smartContracts[shardID] = append(smartContracts[shardID], isc)
	}

	return smartContracts, nil
}

// GetDeployedSCAddresses will return the deployed smart contract addresses as a map for a specific type
func (scp *smartContractParser) GetDeployedSCAddresses(scType string) (map[string]struct{}, error) {
	mapAddresses := make(map[string]struct{})
	for _, sc := range scp.initialSmartContracts {
		if sc.GetType() != scType {
			continue
		}

		if len(sc.AddressesBytes()) == 0 {
			return nil, genesis.ErrSmartContractWasNotDeployed
		}

		for _, address := range sc.AddressesBytes() {
			mapAddresses[string(address)] = struct{}{}
		}
	}

	return mapAddresses, nil
}

// IsInterfaceNil returns if underlying object is true
func (scp *smartContractParser) IsInterfaceNil() bool {
	return scp == nil
}
