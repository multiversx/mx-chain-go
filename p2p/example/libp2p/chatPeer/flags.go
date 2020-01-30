package main

import (
	"flag"
	"strings"

	maddr "github.com/multiformats/go-multiaddr"
)

// A new type we need for writing a custom flag parser
type addrList []maddr.Multiaddr

// String will return the string representation of a multi addr
func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

// Set will update the value for a addrList
func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

// IPFS bootstrap nodes. Used to find other peers in the network.
var defaultBootstrapAddrStrings = make([]string, 0)

// StringsToAddrs will convert from string representations to a slice of Multiaddr
func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

// Config represents the struct which holds the settings
type Config struct {
	BootstrapPeers  addrList
	ListenAddresses addrList
}

// ParseFlags will check and parse the given configuration
func ParseFlags() (Config, error) {
	config := Config{}
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")
	flag.Parse()

	if len(config.BootstrapPeers) == 0 {
		bootstrapPeerAddrs, err := StringsToAddrs(defaultBootstrapAddrStrings)
		if err != nil {
			return config, err
		}
		config.BootstrapPeers = bootstrapPeerAddrs
	}

	return config, nil
}
