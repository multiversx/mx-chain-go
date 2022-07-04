package keysManagement

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
)

var log = logger.GetOrCreate("keysManagement")

type peerInfo struct {
	pid                core.PeerID
	p2pPrivateKeyBytes []byte
	privateKey         crypto.PrivateKey
	machineID          string
}

type virtualPeersHolder struct {
	mut                  sync.RWMutex
	data                 map[string]*peerInfo
	keyGenerator         crypto.KeyGenerator
	p2pIdentityGenerator P2PIdentityGenerator
}

// ArgsVirtualPeersHolder represents the argument for the virtual peers holder
type ArgsVirtualPeersHolder struct {
	KeyGenerator         crypto.KeyGenerator
	P2PIdentityGenerator P2PIdentityGenerator
}

// NewVirtualPeersHolder creates a new instance of a virtual peers holder
func NewVirtualPeersHolder(args ArgsVirtualPeersHolder) (*virtualPeersHolder, error) {
	err := checkVirtualPeersHolderArgs(args)
	if err != nil {
		return nil, err
	}

	return &virtualPeersHolder{
		data:                 make(map[string]*peerInfo),
		keyGenerator:         args.KeyGenerator,
		p2pIdentityGenerator: args.P2PIdentityGenerator,
	}, nil
}

func checkVirtualPeersHolderArgs(args ArgsVirtualPeersHolder) error {
	if check.IfNil(args.KeyGenerator) {
		return errNilKeyGenerator
	}
	if check.IfNil(args.P2PIdentityGenerator) {
		return errNilP2PIdentityGenerator
	}

	return nil
}

// AddVirtualPeer will try to add a new virtual peer providing the private key bytes.
// It errors if the generated public key is already contained by the struct
// It will auto-generate some fields like the machineID and pid
func (holder *virtualPeersHolder) AddVirtualPeer(privateKeyBytes []byte) error {
	privateKey, err := holder.keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("%w for provided bytes %s", err, hex.EncodeToString(privateKeyBytes))
	}

	publicKey := privateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return fmt.Errorf("%w for provided bytes %s", err, hex.EncodeToString(privateKeyBytes))
	}

	p2pPrivateKeyBytes, pid, err := holder.p2pIdentityGenerator.CreateRandomP2PIdentity()
	if err != nil {
		return err
	}

	pInfo := &peerInfo{
		pid:                pid,
		p2pPrivateKeyBytes: p2pPrivateKeyBytes,
		privateKey:         privateKey,
		machineID:          generateRandomMachineID(),
	}

	holder.mut.Lock()
	defer holder.mut.Unlock()

	_, found := holder.data[string(publicKeyBytes)]
	if found {
		return fmt.Errorf("%w for provided bytes %s and generated public key %s",
			errDuplicatedKey, hex.EncodeToString(privateKeyBytes), hex.EncodeToString(publicKeyBytes))
	}

	holder.data[string(publicKeyBytes)] = pInfo

	log.Debug("added new key definition",
		"hex public key", hex.EncodeToString(publicKeyBytes),
		"pid", pid.Pretty(),
		"machine ID", pInfo.machineID)

	return nil
}

func (holder *virtualPeersHolder) getPeerInfo(pkBytes []byte) *peerInfo {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	return holder.data[string(pkBytes)]
}

func generateRandomMachineID() string {
	buff := make([]byte, common.MaxMachineIDLen/2)
	_, _ = rand.Read(buff)

	return hex.EncodeToString(buff)
}

// GetPrivateKey returns the associated private key with the provided public key bytes. Errors if the key is not found
func (holder *virtualPeersHolder) GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return nil, fmt.Errorf("%w in GetPrivateKey for public key %s",
			errMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.privateKey, nil
}

// GetP2PIdentity returns the associated p2p identity with the provided public key bytes: the private key and the peer ID
func (holder *virtualPeersHolder) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return nil, "", fmt.Errorf("%w in GetP2PIdentity for public key %s",
			errMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.p2pPrivateKeyBytes, pInfo.pid, nil
}

// GetMachineID returns the associated machine ID with the provided public key bytes
func (holder *virtualPeersHolder) GetMachineID(pkBytes []byte) (string, error) {
	pInfo := holder.getPeerInfo(pkBytes)
	if pInfo == nil {
		return "", fmt.Errorf("%w in GetMachineID for public key %s",
			errMissingPublicKeyDefinition, hex.EncodeToString(pkBytes))
	}

	return pInfo.machineID, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *virtualPeersHolder) IsInterfaceNil() bool {
	return holder == nil
}
