package libp2p

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func getPort(port string, handler func(int) error) (int, error) {
	val, err := strconv.Atoi(port)
	if err == nil {
		if val < 0 {
			return 0, fmt.Errorf("%w, %d does not represent a positive value for port", p2p.ErrInvalidPortValue, val)
		}

		return val, nil
	}

	ports := strings.Split(port, "-")
	if len(ports) != 2 {
		return 0, fmt.Errorf("%w, provided port string `%s` is not in the correct format, expected `start-end`", p2p.ErrInvalidPortsRangeString, port)
	}

	startPort, err := strconv.Atoi(ports[0])
	if err != nil {
		return 0, p2p.ErrInvalidStartingPortValue
	}

	endPort, err := strconv.Atoi(ports[1])
	if err != nil {
		return 0, p2p.ErrInvalidEndingPortValue
	}

	if startPort < minRangePortValue {
		return 0, fmt.Errorf("%w, provided starting port should be >= %d", p2p.ErrInvalidValue, minRangePortValue)
	}
	if endPort < startPort {
		return 0, p2p.ErrEndPortIsSmallerThanStartPort
	}

	return choosePort(startPort, endPort, handler)
}

func choosePort(startPort int, endPort int, handler func(int) error) (int, error) {
	log.Info("generating random free port",
		"range", fmt.Sprintf("%d-%d", startPort, endPort),
	)

	ports := make([]int, 0, endPort-startPort+1)
	for i := startPort; i <= endPort; i++ {
		ports = append(ports, i)
	}

	ports = random.FisherYatesShuffle(ports, &random.ConcurrentSafeIntRandomizer{})
	for _, p := range ports {
		err := handler(p)
		if err != nil {
			log.Trace("opening port error",
				"port", p, "error", err)
			continue
		}

		return p, nil
	}

	return 0, fmt.Errorf("%w, range %d-%d", p2p.ErrNoFreePortInRange, startPort, endPort)
}

func checkFreePort(port int) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	_ = l.Close()

	return nil
}
