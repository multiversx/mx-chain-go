// +build linux

package vm

import (
	"os/user"
)

const libName = "libhera.so"

// WASMLibLocation defines the wasm library location
func WASMLibLocation() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	return usr.HomeDir + "/elrond-vm-binaries/" + libName
}
