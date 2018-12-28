package groupSelectors_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/stretchr/testify/assert"
)

//------- NewIndexHashedGroupSelector

func TestNewIndexHashedGroupSelector_NilHasherShouldErr(t *testing.T) {
	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(1, nil)

	assert.Nil(t, ihgs)
	assert.Equal(t, consensus.ErrNilHasher, err)
}

func TestNewIndexHashedGroupSelector_InvalidConsensusSizeShouldErr(t *testing.T) {
	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(0, mock.HasherMock{})

	assert.Nil(t, ihgs)
	assert.Equal(t, consensus.ErrInvalidConsensusSize, err)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(1, mock.HasherMock{})

	assert.NotNil(t, ihgs)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelector_LoadEligibleListNilListShouldErr(t *testing.T) {
	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(10, mock.HasherMock{})

	assert.Equal(t, consensus.ErrNilInputSlice, ihgs.LoadEligibleList(nil))
}

func TestIndexHashedGroupSelector_OkValShouldWork(t *testing.T) {
	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(10, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(*big.NewInt(1), 2, []byte("aaa")),
		mock.NewValidatorMock(*big.NewInt(2), 3, []byte("bbb")),
	}

	err := ihgs.LoadEligibleList(list)
	assert.Nil(t, err)
	assert.Equal(t, list, ihgs.EligibleList())
}

//------- ComputeValidatorsGroup

func TestIndexHashedGroupSelector_ComputeValidatorsGroupWrongSizeShouldErr(t *testing.T) {
	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(10, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(*big.NewInt(1), 2, []byte("aaa")),
		mock.NewValidatorMock(*big.NewInt(2), 3, []byte("bbb")),
	}

	_ = ihgs.LoadEligibleList(list)

	list, err := ihgs.ComputeValidatorsGroup([]byte("aaa"))

	assert.Nil(t, list)
	assert.Equal(t, consensus.ErrSmallEligibleListSize, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(2, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(*big.NewInt(1), 2, []byte("aaa")),
		mock.NewValidatorMock(*big.NewInt(2), 3, []byte("bbb")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup(nil)

	assert.Nil(t, list2)
	assert.Equal(t, consensus.ErrNilRandomness, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroup1ValidatorShouldReturnSame(t *testing.T) {
	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(1, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(*big.NewInt(1), 2, []byte("aaa")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte("aaa"))

	assert.Nil(t, err)
	assert.Equal(t, list, list2)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest2Validators(t *testing.T) {
	hasher := &mock.HasherStub{}

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(2, hasher)

	list := []consensus.Validator{
		mock.NewValidatorMock(*big.NewInt(1), 2, []byte("aaa")),
		mock.NewValidatorMock(*big.NewInt(2), 3, []byte("bbb")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte("random"))

	assert.Nil(t, err)
	assert.Equal(t, list, list2)
}
