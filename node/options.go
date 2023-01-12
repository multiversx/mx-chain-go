package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/p2p"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// WithBootstrapComponents sets up the Node bootstrap components
func WithBootstrapComponents(bootstrapComponents factory.BootstrapComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(bootstrapComponents) {
			return ErrNilBootstrapComponents
		}
		err := bootstrapComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.bootstrapComponents = bootstrapComponents
		n.closableComponents = append(n.closableComponents, bootstrapComponents)

		return nil
	}
}

// WithCoreComponents sets up the Node core components
func WithCoreComponents(coreComponents factory.CoreComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(coreComponents) {
			return ErrNilCoreComponents
		}
		err := coreComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.coreComponents = coreComponents
		n.closableComponents = append(n.closableComponents, coreComponents)
		return nil
	}
}

// WithStatusCoreComponents sets up the Node status core components
func WithStatusCoreComponents(statusCoreComponents factory.StatusCoreComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(statusCoreComponents) {
			return ErrNilStatusCoreComponents
		}
		err := statusCoreComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.statusCoreComponents = statusCoreComponents
		n.closableComponents = append(n.closableComponents, statusCoreComponents)
		return nil
	}
}

// WithCryptoComponents sets up Node crypto components
func WithCryptoComponents(cryptoComponents factory.CryptoComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(cryptoComponents) {
			return ErrNilCryptoComponents
		}
		err := cryptoComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.cryptoComponents = cryptoComponents
		n.closableComponents = append(n.closableComponents, cryptoComponents)
		return nil
	}
}

// WithDataComponents sets up the Node data components
func WithDataComponents(dataComponents factory.DataComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(dataComponents) {
			return ErrNilDataComponents
		}
		err := dataComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.dataComponents = dataComponents
		n.closableComponents = append(n.closableComponents, dataComponents)
		return nil
	}
}

// WithNetworkComponents sets up the Node network components
func WithNetworkComponents(networkComponents factory.NetworkComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(networkComponents) {
			return ErrNilNetworkComponents
		}
		err := networkComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.networkComponents = networkComponents
		n.closableComponents = append(n.closableComponents, networkComponents)
		return nil
	}
}

// WithProcessComponents sets up the Node process components
func WithProcessComponents(processComponents factory.ProcessComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(processComponents) {
			return ErrNilProcessComponents
		}
		err := processComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.processComponents = processComponents
		n.closableComponents = append(n.closableComponents, processComponents)
		return nil
	}
}

// WithStateComponents sets up the Node state components
func WithStateComponents(stateComponents factory.StateComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(stateComponents) {
			return ErrNilStateComponents
		}
		err := stateComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.stateComponents = stateComponents
		n.closableComponents = append(n.closableComponents, stateComponents)
		return nil
	}
}

// WithStatusComponents sets up the Node status components
func WithStatusComponents(statusComponents factory.StatusComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(statusComponents) {
			return ErrNilStatusComponents
		}
		err := statusComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.statusComponents = statusComponents
		n.closableComponents = append(n.closableComponents, statusComponents)
		return nil
	}
}

// WithHeartbeatV2Components sets up the Node heartbeatV2 components
func WithHeartbeatV2Components(heartbeatV2Components factory.HeartbeatV2ComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(heartbeatV2Components) {
			return ErrNilStatusComponents
		}
		err := heartbeatV2Components.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.heartbeatV2Components = heartbeatV2Components
		n.closableComponents = append(n.closableComponents, heartbeatV2Components)
		return nil
	}
}

// WithConsensusComponents sets up the Node consensus components
func WithConsensusComponents(consensusComponents factory.ConsensusComponentsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(consensusComponents) {
			return ErrNilStatusComponents
		}
		err := consensusComponents.CheckSubcomponents()
		if err != nil {
			return err
		}
		n.consensusComponents = consensusComponents
		n.closableComponents = append(n.closableComponents, consensusComponents)
		return nil
	}
}

// WithInitialNodesPubKeys sets up the initial nodes public key option for the Node
func WithInitialNodesPubKeys(pubKeys map[uint32][]string) Option {
	return func(n *Node) error {
		n.initialNodesPubkeys = pubKeys
		return nil
	}
}

// WithRoundDuration sets up the round duration option for the Node
func WithRoundDuration(roundDuration uint64) Option {
	return func(n *Node) error {
		if roundDuration == 0 {
			return ErrZeroRoundDurationNotSupported
		}
		n.roundDuration = roundDuration
		return nil
	}
}

// WithConsensusGroupSize sets up the consensus group size option for the Node
func WithConsensusGroupSize(consensusGroupSize int) Option {
	return func(n *Node) error {
		if consensusGroupSize < 1 {
			return ErrNegativeOrZeroConsensusGroupSize
		}
		log.Info("consensus group", "size", consensusGroupSize)
		n.consensusGroupSize = consensusGroupSize
		return nil
	}
}

// WithGenesisTime sets up the genesis time option for the Node
func WithGenesisTime(genesisTime time.Time) Option {
	return func(n *Node) error {
		n.genesisTime = genesisTime
		return nil
	}
}

// WithConsensusType sets up the consensus type option for the Node
func WithConsensusType(consensusType string) Option {
	return func(n *Node) error {
		n.consensusType = consensusType
		return nil
	}
}

// WithBootstrapRoundIndex sets up a bootstrapRoundIndex option for the Node
func WithBootstrapRoundIndex(bootstrapRoundIndex uint64) Option {
	return func(n *Node) error {
		n.bootstrapRoundIndex = bootstrapRoundIndex
		return nil
	}
}

// WithPeerDenialEvaluator sets up a peer denial evaluator for the Node
func WithPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) Option {
	return func(n *Node) error {
		if check.IfNil(handler) {
			return fmt.Errorf("%w for WithPeerDenialEvaluator", ErrNilPeerDenialEvaluator)
		}
		n.peerDenialEvaluator = handler
		return nil
	}
}

// WithRequestedItemsHandler sets up a requested items handler for the Node
func WithRequestedItemsHandler(requestedItemsHandler dataRetriever.RequestedItemsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(requestedItemsHandler) {
			return ErrNilRequestedItemsHandler
		}
		n.requestedItemsHandler = requestedItemsHandler
		return nil
	}
}

// WithAddressSignatureSize sets up an addressSignatureSize option for the Node
func WithAddressSignatureSize(signatureSize int) Option {
	return func(n *Node) error {
		n.addressSignatureSize = signatureSize
		emptyByteSlice := bytes.Repeat([]byte{0}, signatureSize)
		hexEncodedEmptyByteSlice := hex.EncodeToString(emptyByteSlice)
		n.addressSignatureHexSize = len(hexEncodedEmptyByteSlice)

		return nil
	}
}

// WithValidatorSignatureSize sets up a validatorSignatureSize option for the Node
func WithValidatorSignatureSize(signatureSize int) Option {
	return func(n *Node) error {
		n.validatorSignatureSize = signatureSize
		return nil
	}
}

// WithPublicKeySize sets up a publicKeySize option for the Node
func WithPublicKeySize(publicKeySize int) Option {
	return func(n *Node) error {
		n.publicKeySize = publicKeySize
		return nil
	}
}

// WithNodeStopChannel sets up the channel which will handle closing the node
func WithNodeStopChannel(channel chan endProcess.ArgEndProcess) Option {
	return func(n *Node) error {
		if channel == nil {
			return ErrNilNodeStopChannel
		}
		n.chanStopNodeProcess = channel

		return nil
	}
}

// WithEnableSignTxWithHashEpoch sets up enableSignTxWithHashEpoch for the node
func WithEnableSignTxWithHashEpoch(enableSignTxWithHashEpoch uint32) Option {
	return func(n *Node) error {
		n.enableSignTxWithHashEpoch = enableSignTxWithHashEpoch
		return nil
	}
}

// WithImportMode sets up the flag if the node is running in import mode
func WithImportMode(importMode bool) Option {
	return func(n *Node) error {
		n.isInImportMode = importMode
		return nil
	}
}

// WithESDTNFTStorageHandler sets the esdt nft storage handler
func WithESDTNFTStorageHandler(storageHandler vmcommon.ESDTNFTStorageHandler) Option {
	return func(node *Node) error {
		if check.IfNil(storageHandler) {
			return ErrNilESDTNFTStorageHandler
		}

		node.esdtStorageHandler = storageHandler
		return nil
	}
}
