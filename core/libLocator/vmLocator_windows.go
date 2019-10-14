// +build windows

package vm

import (
	"os/user"
)

const libName = "libhera.dll"

func WASMLibLocation() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	return usr.HomeDir + "\\elrond-vm-binaries\\" + libName
}
