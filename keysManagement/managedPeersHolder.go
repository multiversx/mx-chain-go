package keysManagement

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/redundancy/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("keysManagement")

type managedPeersHolder struct {
	mut                         sync.RWMutex
	defaultPeerInfoCurrentIndex int
	providedIdentities          map[string]*peerInfo
	data                        map[string]*peerInfo
	pids                        map[core.PeerID]struct{}
	keyGenerator                crypto.KeyGenerator
	p2pKeyGenerator             crypto.KeyGenerator
	isMainMachine               bool
	maxRoundsOfInactivity       int
	defaultName                 string
	defaultIdentity             string
	p2pKeyConverter             p2p.P2PKeyConverter
}

// ArgsManagedPeersHolder represents the argument for the managed peers holder
type ArgsManagedPeersHolder struct {
	KeyGenerator          crypto.KeyGenerator
	P2PKeyGenerator       crypto.KeyGenerator
	MaxRoundsOfInactivity int
	PrefsConfig           config.Preferences
	P2PKeyConverter       p2p.P2PKeyConverter
}

// NewManagedPeersHolder creates a new instance of a managed peers holder
func NewManagedPeersHolder(args ArgsManagedPeersHolder) (*managedPeersHolder, error) {
	err := checkManagedPeersHolderArgs(args)
	if err != nil {
		return nil, err
	}

	handler := common.NewRedundancyHandler()

	holder := &managedPeersHolder{
		defaultPeerInfoCurrentIndex: 0,
		pids:                        make(map[core.PeerID]struct{}),
		keyGenerator:                args.KeyGenerator,
		p2pKeyGenerator:             args.P2PKeyGenerator,
		isMainMachine:               !handler.IsRedundancyNode(args.MaxRoundsOfInactivity),
		maxRoundsOfInactivity:       args.MaxRoundsOfInactivity,
		defaultName:                 args.PrefsConfig.Preferences.NodeDisplayName,
		defaultIdentity:             args.PrefsConfig.Preferences.Identity,
		p2pKeyConverter:             args.P2PKeyConverter,
		data:                        make(map[string]*peerInfo),
	}

	holder.providedIdentities, err = holder.createProvidedIdentitiesMap(args.PrefsConfig.NamedIdentity)
	if err != nil {
		return nil, err
	}

	return holder, nil
}

func checkManagedPeersHolderArgs(args ArgsManagedPeersHolder) error {
	if check.IfNil(args.KeyGenerator) {
		return fmt.Errorf("%w for args.KeyGenerator", ErrNilKeyGenerator)
	}
	if check.IfNil(args.P2PKeyGenerator) {
		return fmt.Errorf("%w for args.P2PKeyGenerator", ErrNilKeyGenerator)
	}
	err := common.CheckMaxRoundsOfInactivity(args.MaxRoundsOfInactivity)
	if err != nil {
		return err
	}
	if check.IfNil(args.P2PKeyConverter) {
		return fmt.Errorf("%w for args.P2PKeyConverter", ErrNilP2PKeyConverter)
	}

	return nil
}

func (holder *managedPeersHolder) createProvidedIdentitiesMap(namedIdentities []config.NamedIdentity) (map[string]*peerInfo, error) {
	dataMap := make(map[string]*peerInfo)

	for _, identity := range namedIdentities {
		index := 0
		for _, blsKey := range identity.BLSKeys {
			bls, err := hex.DecodeString(blsKey)
			if err != nil {
				return nil, fmt.Errorf("%w for key %s", ErrInvalidKey, blsKey)
			}
			if len(bls) == 0 {
				continue
			}

			blsStr := string(bls)
			dataMap[blsStr] = &peerInfo{
				machineID:    generateRandomMachineID(),
				nodeName:     generateNodeName(identity.NodeName, index),
				nodeIdentity: identity.Identity,
			}
			index++
		}
	}

	return dataMap, nil
}

func generateNodeName(providedNodeName string, index int) string {
	if len(providedNodeName) == 0 {
		return ""
	}

	return fmt.Sprintf("%s-%02d", providedNodeName, index)
}

// AddManagedPeer will try to add a new managed peer providing the private key bytes.
// It errors if the generated public key is already contained by the struct
// It will auto-generate some fields like the machineID and pid
func (holder *managedPeersHolder) AddManagedPeer(privateKeyBytes []byte) error {
	privateKey, err := holder.keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("%w for provided bytes %s", err, hex.EncodeToString(privateKeyBytes))
	}

	publicKey := privateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return fmt.Errorf("%w for provided bytes %s", err, hex.EncodeToString(privateKeyBytes))
	}

	p2pPrivateKey, p2pPublicKey := holder.p2pKeyGenerator.GeneratePair()

	p2pPrivateKeyBytes, err := p2pPrivateKey.ToByteArray()
	if err != nil {
		return err
	}

	pid, err := holder.p2pKeyConverter.ConvertPublicKeyToPeerID(p2pPublicKey)
	if err != nil {
		return err
	}

	holder.mut.Lock()
	defer holder.mut.Unlock()

	pInfo, found := holder.data[string(publicKeyBytes)]
	if found && len(pInfo.pid.Bytes()) != 0 {
		return fmt.Errorf("%w for provided bytes %s and generated public key %s",
			ErrDuplicatedKey, hex.EncodeToString(privateKeyBytes), hex.EncodeToString(publicKeyBytes))
	}

	pInfo, found = holder.providedIdentities[string(publicKeyBytes)]
	if !found {
		pInfo = &peerInfo{
			handler:      common.NewRedundancyHandler(),
			machineID:    generateRandomMachineID(),
			nodeName:     generateNodeName(holder.defaultName, holder.defaultPeerInfoCurrentIndex),
			nodeIdentity: holder.defaultIdentity,
		}
		holder.defaultPeerInfoCurrentIndex++
	}

	pInfo.pid = pid
	pInfo.p2pPrivateKeyBytes = p2pPrivateKeyBytes
	pInfo.privateKey = privateKey
	holder.data[string(publicKeyBytes)] = pInfo
	holder.pids[pid] = struct{}{}

	log.Debug("added new key definition",
		"hex public key", hex.EncodeToString(publicKeyBytes),
		"pid", pid.Pretty(),
		"machine ID", pInfo.machineID,
		"name", pInfo.nodeName,
		"identity", pInfo.nodeIdentity)

	return nil
}

func (holder *managedPeersHolder) getPeerInfo(pkBytes []byte) *peerInfo {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	return holder.data[string(pkBytes)]
}

func generateRandomMachineID() string {
	buff := make([]byte, core.MaxMachineIDLen/2)
	_, _ = rand.Read(buff)

	return hex.EncodeToString(buff)
}

// GetPrivateKey returns the associated private key with the provided public key bytes. Errors if the key is not found
func (holder *managedPeersHolder) GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return nil, fmt.Errorf("%w in GetPrivateKey for public key %s",
			ErrMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.privateKey, nil
}

// GetP2PIdentity returns the associated p2p identity with the provided public key bytes: the private key and the peer ID
func (holder *managedPeersHolder) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return nil, "", fmt.Errorf("%w in GetP2PIdentity for public key %s",
			ErrMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.p2pPrivateKeyBytes, pInfo.pid, nil
}

// GetMachineID returns the associated machine ID with the provided public key bytes
func (holder *managedPeersHolder) GetMachineID(pkBytes []byte) (string, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return "", fmt.Errorf("%w in GetMachineID for public key %s",
			ErrMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.machineID, nil
}

// GetNameAndIdentity returns the associated name and identity with the provided public key bytes
func (holder *managedPeersHolder) GetNameAndIdentity(pkBytes []byte) (string, string, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return "", "", fmt.Errorf("%w in GetNameAndIdentity for public key %s",
			ErrMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.nodeName, pInfo.nodeIdentity, nil
}

// IncrementRoundsWithoutReceivedMessages increments the number of rounds without received messages on a provided public key
func (holder *managedPeersHolder) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
	if holder.isMainMachine {
		return
	}

	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return
	}

	pInfo.incrementRoundsWithoutReceivedMessages()
}

// ResetRoundsWithoutReceivedMessages resets the number of rounds without received messages on a provided public key
func (holder *managedPeersHolder) ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID) {
	if holder.isMainMachine {
		return
	}

	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return
	}
	if bytes.Equal(pInfo.pid.Bytes(), pid.Bytes()) {
		return
	}

	pInfo.resetRoundsWithoutReceivedMessages()
}

// GetManagedKeysByCurrentNode returns all keys that will be managed by this node
func (holder *managedPeersHolder) GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	allManagedKeys := make(map[string]crypto.PrivateKey)
	for pk, pInfo := range holder.data {
		isSlaveAndMainFailed := !holder.isMainMachine && !pInfo.isNodeActiveOnMainMachine(holder.maxRoundsOfInactivity)
		shouldAddToMap := holder.isMainMachine || isSlaveAndMainFailed
		if !shouldAddToMap {
			continue
		}

		allManagedKeys[pk] = pInfo.privateKey
	}

	return allManagedKeys
}

// IsKeyManagedByCurrentNode returns true if the key is managed by the current node
func (holder *managedPeersHolder) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return false
	}

	if holder.isMainMachine {
		return true
	}

	return !pInfo.isNodeActiveOnMainMachine(holder.maxRoundsOfInactivity)
}

// IsKeyRegistered returns true if the key is registered (not necessarily managed by the current node)
func (holder *managedPeersHolder) IsKeyRegistered(pkBytes []byte) bool {
	pInfo := holder.getPeerInfo(pkBytes)
	return pInfo != nil
}

// IsPidManagedByCurrentNode returns true if the peer id is managed by the current node
func (holder *managedPeersHolder) IsPidManagedByCurrentNode(pid core.PeerID) bool {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	_, found := holder.pids[pid]

	return found
}

// IsKeyValidator returns true if the key validator status was set to true
func (holder *managedPeersHolder) IsKeyValidator(pkBytes []byte) bool {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return false
	}

	return pInfo.isNodeValidator()
}

// SetValidatorState sets the provided validator status for the key
func (holder *managedPeersHolder) SetValidatorState(pkBytes []byte, state bool) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return
	}

	pInfo.setNodeValidator(state)
}

// GetNextPeerAuthenticationTime returns the next time the key should try to send peer authentication again
func (holder *managedPeersHolder) GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return time.Now(), fmt.Errorf("%w in GetNextPeerAuthenticationTime for public key %s",
			ErrMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.getNextPeerAuthenticationTime(), nil
}

// SetNextPeerAuthenticationTime sets the next time the key should try to send peer authentication
func (holder *managedPeersHolder) SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return
	}

	pInfo.setNextPeerAuthenticationTime(nextTime)
}

// IsMultiKeyMode returns true if the node has at least one managed key, regardless it was set as a main machine or a backup machine
func (holder *managedPeersHolder) IsMultiKeyMode() bool {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	return len(holder.data) > 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *managedPeersHolder) IsInterfaceNil() bool {
	return holder == nil
}
