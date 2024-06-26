package keysManagement

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
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

const (
	redundancyReasonForOneKey       = "multikey node stepped in with one key"
	redundancyReasonForMultipleKeys = "multikey node stepped in with %d keys"
)

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

	holder := &managedPeersHolder{
		defaultPeerInfoCurrentIndex: 0,
		pids:                        make(map[core.PeerID]struct{}),
		keyGenerator:                args.KeyGenerator,
		p2pKeyGenerator:             args.P2PKeyGenerator,
		isMainMachine:               common.IsMainNode(args.MaxRoundsOfInactivity),
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
			machineID:    generateRandomMachineID(),
			nodeName:     generateNodeName(holder.defaultName, holder.defaultPeerInfoCurrentIndex),
			nodeIdentity: holder.defaultIdentity,
		}
		holder.defaultPeerInfoCurrentIndex++
	}

	pInfo.handler = common.NewRedundancyHandler()
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

// GetManagedKeysByCurrentNode returns all keys that should act as validator(main or backup that took over) and will be managed by this node
func (holder *managedPeersHolder) GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	allManagedKeys := make(map[string]crypto.PrivateKey)
	for pk, pInfo := range holder.data {
		shouldAddToMap := pInfo.shouldActAsValidator(holder.maxRoundsOfInactivity)
		if !shouldAddToMap {
			continue
		}

		allManagedKeys[pk] = pInfo.privateKey
	}

	return allManagedKeys
}

// GetLoadedKeysByCurrentNode returns all keys that were loaded and will be managed by this node
func (holder *managedPeersHolder) GetLoadedKeysByCurrentNode() [][]byte {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	allLoadedKeys := make([][]byte, 0, len(holder.data))
	for pk := range holder.data {
		allLoadedKeys = append(allLoadedKeys, []byte(pk))
	}

	sort.Slice(allLoadedKeys, func(i, j int) bool {
		return string(allLoadedKeys[i]) < string(allLoadedKeys[j])
	})

	return allLoadedKeys
}

// IsKeyManagedByCurrentNode returns true if the key is managed by the current node
func (holder *managedPeersHolder) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return false
	}

	return pInfo.shouldActAsValidator(holder.maxRoundsOfInactivity)
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

// GetRedundancyStepInReason returns the reason if the current node stepped in as a redundancy node
// Returns empty string if the current node is the main multikey machine, the machine is not running in multikey mode
// or the machine is acting as a backup but the main machine is acting accordingly
func (holder *managedPeersHolder) GetRedundancyStepInReason() string {
	if holder.isMainMachine {
		return ""
	}

	numManagedKeys := len(holder.GetManagedKeysByCurrentNode())
	if numManagedKeys == 0 {
		return ""
	}

	if numManagedKeys == 1 {
		return redundancyReasonForOneKey
	}

	return fmt.Sprintf(redundancyReasonForMultipleKeys, numManagedKeys)
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *managedPeersHolder) IsInterfaceNil() bool {
	return holder == nil
}
