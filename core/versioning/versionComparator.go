package versioning

import (
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
)

const numComponents = 3

type versionComparator struct {
	version string
	major   string
	minor   string
	release string
}

// NewVersionComparator returns a new version comparator instance
func NewVersionComparator(providedVersion string) (*versionComparator, error) {
	vc := &versionComparator{
		version: providedVersion,
	}

	var err error
	vc.major, vc.minor, vc.release, err = vc.splitVersionComponents(providedVersion)
	if err != nil {
		return nil, err
	}

	return vc, nil
}

func (vc *versionComparator) splitVersionComponents(version string) (string, string, string, error) {
	components := strings.Split(version, ".")
	if len(components) != numComponents {
		return "", "", "", fmt.Errorf("%w, expected %d, got %d",
			core.ErrVersionNumComponents,
			numComponents,
			len(components),
		)
	}

	return components[0], components[1], components[2], nil
}

// Check compares if the provided value is compatible with the stored values. The comparison is done on all 3 components:
// major, minor and release. A wildcard in any of the components will accept any string
func (vc *versionComparator) Check(version string) error {
	major, minor, release, err := vc.splitVersionComponents(version)
	if err != nil {
		return err
	}

	if major != vc.major && vc.major != "*" {
		return fmt.Errorf("%w, expected version %s, got %s", core.ErrMajorVersionMismatch, vc.version, version)
	}
	if minor != vc.minor && vc.minor != "*" {
		return fmt.Errorf("%w, expected version %s, got %s", core.ErrMinorVersionMismatch, vc.version, version)
	}
	if release != vc.release && vc.release != "*" {
		return fmt.Errorf("%w, expected version %s, got %s", core.ErrReleaseVersionMismatch, vc.version, version)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vc *versionComparator) IsInterfaceNil() bool {
	return vc == nil
}
