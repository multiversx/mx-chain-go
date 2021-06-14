package headerCheck

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const wildcard = "*"
const keySize = 4

type headerIntegrityVerifier struct {
	referenceChainID []byte
	versions         []config.VersionByEpochs
	defaultVersion   string
	versionCache     storage.Cacher
}

// NewHeaderIntegrityVerifier returns a new instance of a structure capable of verifying the integrity of a provided header
func NewHeaderIntegrityVerifier(
	referenceChainID []byte,
	versionsByEpochs []config.VersionByEpochs,
	defaultVersion string,
	versionCache storage.Cacher,
) (*headerIntegrityVerifier, error) {

	if len(referenceChainID) == 0 {
		return nil, ErrInvalidReferenceChainID
	}
	if check.IfNil(versionCache) {
		return nil, fmt.Errorf("%w, in NewHeaderVersioningHandler", ErrNilCacher)
	}

	hdrIntVer := &headerIntegrityVerifier{
		referenceChainID: referenceChainID,
		defaultVersion:   defaultVersion,
		versionCache:     versionCache,
	}
	var err error
	hdrIntVer.versions, err = hdrIntVer.prepareVersions(versionsByEpochs)
	if err != nil {
		return nil, err
	}

	err = hdrIntVer.checkVersionLength([]byte(defaultVersion))
	if err != nil {
		return nil, err
	}

	return hdrIntVer, err
}

func (hdrIntVer *headerIntegrityVerifier) prepareVersions(versionsByEpochs []config.VersionByEpochs) ([]config.VersionByEpochs, error) {
	if len(versionsByEpochs) == 0 {
		return nil, ErrEmptyVersionsByEpochsList
	}

	sort.Slice(versionsByEpochs, func(i, j int) bool {
		return versionsByEpochs[i].StartEpoch < versionsByEpochs[j].StartEpoch
	})

	currentEpoch := uint32(0)
	for idx, ver := range versionsByEpochs {
		if idx == 0 && ver.StartEpoch != 0 {
			return nil, fmt.Errorf("%w first version should start on epoch 0", ErrInvalidVersionOnEpochValues)
		}

		if idx > 0 && currentEpoch >= ver.StartEpoch {
			return nil, fmt.Errorf("%w, StartEpoch is greater or equal to next epoch StartEpoch value, version %s",
				ErrInvalidVersionOnEpochValues, ver.Version)
		}
		currentEpoch = ver.StartEpoch

		if len(ver.Version) > core.MaxSoftwareVersionLengthInBytes {
			return nil, fmt.Errorf("%w for version %s",
				ErrInvalidVersionStringTooLong, ver.Version)
		}
	}

	return versionsByEpochs, nil
}

// GetVersion returns the version by providing the epoch
func (hdrIntVer *headerIntegrityVerifier) GetVersion(epoch uint32) string {
	ver := hdrIntVer.getMatchingVersion(epoch)
	if ver == wildcard {
		return hdrIntVer.defaultVersion
	}

	return ver
}

func (hdrIntVer *headerIntegrityVerifier) getMatchingVersion(epoch uint32) string {
	storedVersion, ok := hdrIntVer.getFromCache(epoch)
	if ok {
		return storedVersion
	}

	matchingVersion := hdrIntVer.versions[len(hdrIntVer.versions)-1].Version
	for idx := 0; idx < len(hdrIntVer.versions)-1; idx++ {
		crtVer := hdrIntVer.versions[idx]
		nextVer := hdrIntVer.versions[idx+1]
		if crtVer.StartEpoch <= epoch && epoch < nextVer.StartEpoch {
			hdrIntVer.setInCache(epoch, crtVer.Version)

			return crtVer.Version
		}
	}

	hdrIntVer.setInCache(epoch, matchingVersion)

	return matchingVersion
}

func (hdrIntVer *headerIntegrityVerifier) getFromCache(epoch uint32) (string, bool) {
	key := make([]byte, keySize)
	binary.BigEndian.PutUint32(key, epoch)

	obj, ok := hdrIntVer.versionCache.Get(key)
	if !ok {
		return "", false
	}

	str, ok := obj.(string)

	return str, ok
}

func (hdrIntVer *headerIntegrityVerifier) setInCache(epoch uint32, version string) {
	key := make([]byte, keySize)
	binary.BigEndian.PutUint32(key, epoch)

	_ = hdrIntVer.versionCache.Put(key, version, len(key)+len(version))
}

// Verify will check the header's fields such as the chain ID or the software version
func (hdrIntVer *headerIntegrityVerifier) Verify(hdr data.HeaderHandler) error {
	if len(hdr.GetReserved()) > 0 {
		return process.ErrReservedFieldNotSupportedYet
	}

	err := hdrIntVer.checkSoftwareVersion(hdr)
	if err != nil {
		return err
	}

	return hdrIntVer.checkChainID(hdr)
}

func (hdrIntVer *headerIntegrityVerifier) checkVersionLength(version []byte) error {
	if len(version) == 0 || len(version) > core.MaxSoftwareVersionLengthInBytes {
		return fmt.Errorf("%w when checking lenghts", ErrInvalidSoftwareVersion)
	}

	return nil
}

// checkSoftwareVersion returns nil if the software version has the correct length
func (hdrIntVer *headerIntegrityVerifier) checkSoftwareVersion(hdr data.HeaderHandler) error {
	err := hdrIntVer.checkVersionLength(hdr.GetSoftwareVersion())
	if err != nil {
		return err
	}

	version := hdrIntVer.getMatchingVersion(hdr.GetEpoch())
	if version == wildcard {
		return nil
	}

	if !bytes.Equal([]byte(version), hdr.GetSoftwareVersion()) {
		return fmt.Errorf("%w, got: %s, should have been %s",
			ErrSoftwareVersionMismatch, hex.EncodeToString(hdr.GetSoftwareVersion()),
			hex.EncodeToString([]byte(version)),
		)
	}

	return nil
}

// checkChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (hdrIntVer *headerIntegrityVerifier) checkChainID(hdr data.HeaderHandler) error {
	if !bytes.Equal(hdrIntVer.referenceChainID, hdr.GetChainID()) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			ErrInvalidChainID,
			hex.EncodeToString(hdrIntVer.referenceChainID),
			hex.EncodeToString(hdr.GetChainID()),
		)
	}

	return nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (hdrIntVer *headerIntegrityVerifier) IsInterfaceNil() bool {
	return hdrIntVer == nil
}
