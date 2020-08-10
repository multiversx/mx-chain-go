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

type headerVersioningHandler struct {
	referenceChainID []byte
	versions         []config.VersionByEpochs
	defaultVersion   string
	versionCache     storage.Cacher
}

// NewHeaderVersioningHandler returns a new instance of a structure capable of handling the versions of a header
// based on the provided epoch
func NewHeaderVersioningHandler(
	referenceChainID []byte,
	versionsByEpochs []config.VersionByEpochs,
	defaultVersion string,
	versionCache storage.Cacher,
) (*headerVersioningHandler, error) {

	if len(referenceChainID) == 0 {
		return nil, ErrInvalidReferenceChainID
	}
	if check.IfNil(versionCache) {
		return nil, fmt.Errorf("%w, in NewHeaderVersioningHandler", ErrNilCacher)
	}

	hvh := &headerVersioningHandler{
		referenceChainID: referenceChainID,
		defaultVersion:   defaultVersion,
		versionCache:     versionCache,
	}
	var err error
	hvh.versions, err = hvh.prepareVersions(versionsByEpochs)
	if err != nil {
		return nil, err
	}

	err = hvh.checkVersionLength([]byte(defaultVersion))
	if err != nil {
		return nil, err
	}

	return hvh, err
}

func (hvh *headerVersioningHandler) prepareVersions(versionsByEpochs []config.VersionByEpochs) ([]config.VersionByEpochs, error) {
	sort.Slice(versionsByEpochs, func(i, j int) bool {
		return versionsByEpochs[i].StartEpoch < versionsByEpochs[j].StartEpoch
	})

	currentEpoch := uint32(0)
	for _, ver := range versionsByEpochs {
		if ver.StartEpoch >= ver.EndEpoch {
			return nil, fmt.Errorf("%w, StartEpoch is greater or equal to EndEpoch, version %s",
				ErrInvalidVersionOnEpochValues, ver.Version)
		}
		if currentEpoch != ver.StartEpoch {
			return nil, fmt.Errorf("%w for version %s",
				ErrInvalidVersionFoundGaps, ver.Version)
		}
		if len(ver.Version) > core.MaxSoftwareVersionLengthInBytes {
			return nil, fmt.Errorf("%w for version %s",
				ErrInvalidVersionStringTooLong, ver.Version)
		}

		currentEpoch = ver.EndEpoch
	}

	return versionsByEpochs, nil
}

// GetVersion returns the version by providing the epoch
func (hvh *headerVersioningHandler) GetVersion(epoch uint32) string {
	ver := hvh.getMatchingVersion(epoch)
	if ver == wildcard {
		return hvh.defaultVersion
	}

	return ver
}

// GetVersion returns the version by providing the epoch
func (hvh *headerVersioningHandler) getMatchingVersion(epoch uint32) string {
	storedVersion, ok := hvh.getFromCache(epoch)
	if ok {
		return storedVersion
	}

	for _, ver := range hvh.versions {
		if ver.StartEpoch <= epoch && epoch < ver.EndEpoch {
			hvh.setInCache(epoch, ver.Version)

			return ver.Version
		}
	}

	log.Debug("headerVersioningHandler.GetVersion version not found",
		"epoch", epoch, "default version", hvh.defaultVersion)

	hvh.setInCache(epoch, hvh.defaultVersion)

	return hvh.defaultVersion
}

func (hvh *headerVersioningHandler) getFromCache(epoch uint32) (string, bool) {
	key := make([]byte, keySize)
	binary.BigEndian.PutUint32(key, epoch)

	obj, ok := hvh.versionCache.Get(key)
	if !ok {
		return "", false
	}

	str, ok := obj.(string)

	return str, ok
}

func (hvh *headerVersioningHandler) setInCache(epoch uint32, version string) {
	key := make([]byte, keySize)
	binary.BigEndian.PutUint32(key, epoch)

	_ = hvh.versionCache.Put(key, version, len(key)+len(version))
}

// Verify will check the header's fields such as the chain ID or the software version
func (hvh *headerVersioningHandler) Verify(hdr data.HeaderHandler) error {
	if len(hdr.GetReserved()) > 0 {
		return process.ErrReservedFieldNotSupportedYet
	}

	err := hvh.checkSoftwareVersion(hdr)
	if err != nil {
		return err
	}

	return hvh.checkChainID(hdr)
}

func (hvh *headerVersioningHandler) checkVersionLength(version []byte) error {
	if len(version) == 0 || len(version) > core.MaxSoftwareVersionLengthInBytes {
		return fmt.Errorf("%w when checking lenghts", ErrInvalidSoftwareVersion)
	}

	return nil
}

// checkSoftwareVersion returns nil if the software version has the correct length
func (hvh *headerVersioningHandler) checkSoftwareVersion(hdr data.HeaderHandler) error {
	err := hvh.checkVersionLength(hdr.GetSoftwareVersion())
	if err != nil {
		return err
	}

	version := hvh.getMatchingVersion(hdr.GetEpoch())
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
func (hvh *headerVersioningHandler) checkChainID(hdr data.HeaderHandler) error {
	if !bytes.Equal(hvh.referenceChainID, hdr.GetChainID()) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			ErrInvalidChainID,
			hex.EncodeToString(hvh.referenceChainID),
			hex.EncodeToString(hdr.GetChainID()),
		)
	}

	return nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (hvh *headerVersioningHandler) IsInterfaceNil() bool {
	return hvh == nil
}
