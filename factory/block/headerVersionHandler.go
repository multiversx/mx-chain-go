package block

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

const wildcard = "*"

type headerVersionHandler struct {
	versions       []config.VersionByEpochs
	defaultVersion string
}

// NewHeaderVersionHandler returns a new instance of a structure capable of handling the header versions
func NewHeaderVersionHandler(
	versionsByEpochs []config.VersionByEpochs,
	defaultVersion string,
) (*headerVersionHandler, error) {
	hvh := &headerVersionHandler{
		defaultVersion: defaultVersion,
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

		if len(ver.Version) > common.MaxSoftwareVersionLengthInBytes {
			return nil, fmt.Errorf("%w for version %s",
				ErrInvalidVersionStringTooLong, ver.Version)
		}
	}

	return versionsByEpochs, nil
}

// GetVersion returns the version by providing the epoch
func (hvh *headerVersionHandler) GetVersion(epoch uint32, round uint64) string {
	ver := hvh.getMatchingVersionAndUpdateCache(epoch, round)
	if ver == wildcard {
		return hvh.defaultVersion
	}

	return ver
}

func (hvh *headerVersionHandler) getMatchingVersionAndUpdateCache(epoch uint32, round uint64) string {
	version := hvh.getVersionFromConfig(epoch, round)

	return version
}

func (hvh *headerVersionHandler) getVersionFromConfig(epoch uint32, round uint64) string {
	for idx := 0; idx < len(hvh.versions)-1; idx++ {
		crtVer := hvh.versions[idx]
		nextVer := hvh.versions[idx+1]

		if (crtVer.StartEpoch <= epoch && epoch < nextVer.StartEpoch) || round < nextVer.StartRound {
			return crtVer.Version
		}
	}

	return hvh.versions[len(hvh.versions)-1].Version
}

// Verify will check the header's fields such as the chain ID or the software version
func (hvh *headerVersionHandler) Verify(hdr data.HeaderHandler) error {
	if len(hdr.GetReserved()) > 0 {
		return process.ErrReservedFieldInvalid
	}

	return hvh.checkSoftwareVersion(hdr)
}

func (hvh *headerVersionHandler) checkVersionLength(version []byte) error {
	if len(version) == 0 || len(version) > common.MaxSoftwareVersionLengthInBytes {
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

	version := hvh.getMatchingVersionAndUpdateCache(hdr.GetEpoch(), hdr.GetRound())
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
