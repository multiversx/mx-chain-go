package block

import (
	"bytes"
	"encoding/binary"
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

type headerVersionHandler struct {
	versions       []config.VersionByEpochs
	defaultVersion string
	versionCache   storage.Cacher
}

// NewHeaderVersionHandler returns a new instance of a structure capable of handling the header versions
func NewHeaderVersionHandler(
	versionsByEpochs []config.VersionByEpochs,
	defaultVersion string,
	versionCache storage.Cacher,
) (*headerVersionHandler, error) {
	if check.IfNil(versionCache) {
		return nil, fmt.Errorf("%w, in NewHeaderVersionHandler", ErrNilCacher)
	}

	hvh := &headerVersionHandler{
		defaultVersion: defaultVersion,
		versionCache:   versionCache,
	}

	var err error
	err = hvh.checkVersionLength([]byte(defaultVersion))
	if err != nil {
		return nil, err
	}

	hvh.versions, err = hvh.prepareVersions(versionsByEpochs)
	if err != nil {
		return nil, err
	}

	return hvh, err
}

func (hvh *headerVersionHandler) prepareVersions(versionsByEpochs []config.VersionByEpochs) ([]config.VersionByEpochs, error) {
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
func (hvh *headerVersionHandler) GetVersion(epoch uint32) string {
	ver := hvh.getMatchingVersionAndUpdateCache(epoch)
	if ver == wildcard {
		return hvh.defaultVersion
	}

	return ver
}

// getMatchingVersion returns the version by providing the epoch
func (hvh *headerVersionHandler) getMatchingVersionAndUpdateCache(epoch uint32) string {
	version, ok := hvh.getFromCache(epoch)
	if ok {
		return version
	}

	version = hvh.getVersionFromConfig(epoch)
	hvh.setInCache(epoch, version)

	return version
}

func (hvh *headerVersionHandler) getVersionFromConfig(epoch uint32) string {
	for idx := 0; idx < len(hvh.versions)-1; idx++ {
		crtVer := hvh.versions[idx]
		nextVer := hvh.versions[idx+1]
		if crtVer.StartEpoch <= epoch && epoch < nextVer.StartEpoch {
			return crtVer.Version
		}
	}

	return hvh.versions[len(hvh.versions)-1].Version
}

func (hvh *headerVersionHandler) getFromCache(epoch uint32) (string, bool) {
	key := make([]byte, keySize)
	binary.BigEndian.PutUint32(key, epoch)

	obj, ok := hvh.versionCache.Get(key)
	if !ok {
		return "", false
	}

	str, ok := obj.(string)

	return str, ok
}

func (hvh *headerVersionHandler) setInCache(epoch uint32, version string) {
	key := make([]byte, keySize)
	binary.BigEndian.PutUint32(key, epoch)

	_ = hvh.versionCache.Put(key, version, len(key)+len(version))
}

// Verify will check the header's fields such as the chain ID or the software version
func (hvh *headerVersionHandler) Verify(hdr data.HeaderHandler) error {
	if len(hdr.GetReserved()) > 0 {
		return process.ErrReservedFieldNotSupportedYet
	}

	return hvh.checkSoftwareVersion(hdr)
}

func (hvh *headerVersionHandler) checkVersionLength(version []byte) error {
	if len(version) == 0 || len(version) > core.MaxSoftwareVersionLengthInBytes {
		return fmt.Errorf("%w when checking lenghts", ErrInvalidSoftwareVersion)
	}

	return nil
}

// checkSoftwareVersion returns nil if the software version has the correct length
func (hvh *headerVersionHandler) checkSoftwareVersion(hdr data.HeaderHandler) error {
	err := hvh.checkVersionLength(hdr.GetSoftwareVersion())
	if err != nil {
		return err
	}

	version := hvh.getMatchingVersionAndUpdateCache(hdr.GetEpoch())
	if version == wildcard {
		return nil
	}

	if !bytes.Equal([]byte(version), hdr.GetSoftwareVersion()) {
		return fmt.Errorf("%w, got: %s, should have been %s",
			ErrSoftwareVersionMismatch, string(hdr.GetSoftwareVersion()),
			version,
		)
	}

	return hdr.ValidateHeaderVersion()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (hvh *headerVersionHandler) IsInterfaceNil() bool {
	return hvh == nil
}
