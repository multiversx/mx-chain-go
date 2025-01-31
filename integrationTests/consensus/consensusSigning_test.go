package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	"github.com/multiversx/mx-chain-go/integrationTests"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	logger "github.com/multiversx/mx-chain-logger-go"
)

func initNodesWithTestSigner(
	numMetaNodes,
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
) map[uint32][]*integrationTests.TestFullNode {

	fmt.Println("Step 1. Setup nodes...")

	equivalentProofsActivationEpoch := uint32(0)

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.EquivalentMessagesEnableEpoch = equivalentProofsActivationEpoch
	enableEpochsConfig.FixedOrderInConsensusEnableEpoch = equivalentProofsActivationEpoch

	nodes := integrationTests.CreateNodesWithTestFullNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		consensusType,
		1,
		enableEpochsConfig,
		false,
	)

	time.Sleep(p2pBootstrapDelay)

	for shardID := range nodes {
		if numInvalid < numNodes {
			for i := uint32(0); i < numInvalid; i++ {
				ii := numNodes - i - 1
				nodes[shardID][ii].MultiSigner.CreateSignatureShareCalled = func(privateKeyBytes, message []byte) ([]byte, error) {
					var invalidSigShare []byte
					if i%2 == 0 {
						// invalid sig share but with valid format
						invalidSigShare, _ = hex.DecodeString("2ee350b9a821e20df97ba487a80b0d0ffffca7da663185cf6a562edc7c2c71e3ca46ed71b31bccaf53c626b87f2b6e08")
					} else {
						// sig share with invalid size
						invalidSigShare = bytes.Repeat([]byte("a"), 3)
					}
					log.Warn("invalid sig share from ", "pk", getPkEncoded(nodes[shardID][ii].NodeKeys.MainKey.Pk), "sig", invalidSigShare)

					return invalidSigShare, nil
				}
			}
		}
	}

	return nodes
}

func TestConsensusWithInvalidSigners(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	logger.ToggleLoggerName(true)
	logger.SetLogLevel("*:TRACE,consensus:TRACE")

	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(1)
	roundTime := uint64(1000)

	nodes := initNodesWithTestSigner(numMetaNodes, numNodes, consensusSize, numInvalid, roundTime, blsConsensusType)

	defer func() {
		for shardID := range nodes {
			for _, n := range nodes[shardID] {
				_ = n.MainMessenger.Close()
				_ = n.FullArchiveMessenger.Close()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	for _, nodesList := range nodes {
		for _, n := range nodesList {
			statusComponents := integrationTests.GetDefaultStatusComponents()

			consensusArgs := consensusComp.ConsensusComponentsFactoryArgs{
				Config: config.Config{
					Consensus: config.ConsensusConfig{
						Type: blsConsensusType,
					},
					ValidatorPubkeyConverter: config.PubkeyConfig{
						Length:          96,
						Type:            "bls",
						SignatureLength: 48,
					},
					TrieSync: config.TrieSyncConfig{
						NumConcurrentTrieSyncers:  5,
						MaxHardCapForMissingNodes: 5,
						TrieSyncerVersion:         2,
						CheckNodesOnDisk:          false,
					},
					GeneralSettings: config.GeneralSettingsConfig{
						SyncProcessTimeInMillis: 6000,
					},
				},
				BootstrapRoundIndex:  0,
				CoreComponents:       n.Node.GetCoreComponents(),
				NetworkComponents:    n.Node.GetNetworkComponents(),
				CryptoComponents:     n.Node.GetCryptoComponents(),
				DataComponents:       n.Node.GetDataComponents(),
				ProcessComponents:    n.Node.GetProcessComponents(),
				StateComponents:      n.Node.GetStateComponents(),
				StatusComponents:     statusComponents,
				StatusCoreComponents: n.Node.GetStatusCoreComponents(),
				ScheduledProcessor:   &consensusMocks.ScheduledProcessorStub{},
				IsInImportMode:       n.Node.IsInImportMode(),
			}

			consensusFactory, err := consensusComp.NewConsensusComponentsFactory(consensusArgs)
			require.Nil(t, err)

			managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
			require.Nil(t, err)

			err = managedConsensusComponents.Create()
			require.Nil(t, err)
		}
	}

	fmt.Println("Wait for several rounds...")

	time.Sleep(15 * time.Second)

	fmt.Println("Checking shards...")

	expectedNonce := uint64(10)
	for _, nodesList := range nodes {
		for _, n := range nodesList {
			for i := 1; i < len(nodes); i++ {
				if check.IfNil(n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()) {
					assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
				} else {
					assert.GreaterOrEqual(t, n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce(), expectedNonce)
				}
			}
		}
	}
}
