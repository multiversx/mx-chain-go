package containers

import (
	"errors"
)

// ErrCloseVMContainer signals an error that occurs when closing the VMContainer
var ErrCloseVMContainer = errors.New("cannot close all items")
