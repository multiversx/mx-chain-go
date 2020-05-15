package versioning

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewVersionComparator_WrongVersionShouldErr(t *testing.T) {
	t.Parallel()

	vc, err := NewVersionComparator("not a valid version")

	assert.True(t, errors.Is(err, core.ErrVersionNumComponents))
	assert.True(t, check.IfNil(vc))
}

func TestNewVersionComparator_GoodVersionShouldWork(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "*"
	vc, err := NewVersionComparator(major + "." + minor + "." + release)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(vc))

	assert.Equal(t, major, vc.major)
	assert.Equal(t, minor, vc.minor)
	assert.Equal(t, release, vc.release)
}

//------- Check

func TestNewVersionComparator_CheckNotAversionShouldErr(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "*"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("not a version")

	assert.True(t, errors.Is(err, core.ErrVersionNumComponents))
}

func TestNewVersionComparator_MajorMismatchShouldErr(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "c"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("aa.b.c")

	assert.True(t, errors.Is(err, core.ErrMajorVersionMismatch))
}

func TestNewVersionComparator_MinorMismatchShouldErr(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "c"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.bb.c")

	assert.True(t, errors.Is(err, core.ErrMinorVersionMismatch))
}

func TestNewVersionComparator_ReleaseMismatchShouldErr(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "c"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.b.cc")

	assert.True(t, errors.Is(err, core.ErrReleaseVersionMismatch))
}

func TestNewVersionComparator_ShouldWork(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "c"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.b.c")

	assert.Nil(t, err)
}

func TestNewVersionComparator_WildCardShouldWork(t *testing.T) {
	t.Parallel()

	major := "*"
	minor := "*"
	release := "*"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.b.c")

	assert.Nil(t, err)
}

func TestNewVersionComparator_WildCardMajorShouldWork(t *testing.T) {
	t.Parallel()

	major := "*"
	minor := "b"
	release := "c"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.b.c")

	assert.Nil(t, err)
}

func TestNewVersionComparator_WildCardMinorShouldWork(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "*"
	release := "c"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.b.c")

	assert.Nil(t, err)
}

func TestNewVersionComparator_WildCardReleaseShouldWork(t *testing.T) {
	t.Parallel()

	major := "a"
	minor := "b"
	release := "*"
	vc, _ := NewVersionComparator(major + "." + minor + "." + release)

	err := vc.Check("a.b.c")

	assert.Nil(t, err)
}
