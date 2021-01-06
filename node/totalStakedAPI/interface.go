package totalStakedAPI

import "math/big"

// TotalStakedValueHandler defines the behavior of a component able to return total staked value
type TotalStakedValueHandler interface {
	GetTotalStakedValue() (*big.Int, error)
	IsInterfaceNil() bool
}
