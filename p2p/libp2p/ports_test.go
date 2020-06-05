package libp2p

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func TestGetPort_InvalidStringShouldErr(t *testing.T) {
	t.Parallel()

	port, err := getPort("NaN", checkFreePort)

	assert.Equal(t, 0, port)
	assert.True(t, errors.Is(err, p2p.ErrInvalidPortsRangeString))
}

func TestGetPort_InvalidPortNumberShouldErr(t *testing.T) {
	t.Parallel()

	port, err := getPort("-1", checkFreePort)
	assert.Equal(t, 0, port)
	assert.True(t, errors.Is(err, p2p.ErrInvalidPortValue))
}

func TestGetPort_SinglePortShouldWork(t *testing.T) {
	t.Parallel()

	port, err := getPort("0", checkFreePort)
	assert.Equal(t, 0, port)
	assert.Nil(t, err)

	p := 3638
	port, err = getPort(fmt.Sprintf("%d", p), checkFreePort)
	assert.Equal(t, p, port)
	assert.Nil(t, err)
}

func TestCheckFreePort_InvalidStartingPortShouldErr(t *testing.T) {
	t.Parallel()

	port, err := getPort("NaN-10000", checkFreePort)
	assert.Equal(t, 0, port)
	assert.Equal(t, p2p.ErrInvalidStartingPortValue, err)

	port, err = getPort("1024-10000", checkFreePort)
	assert.Equal(t, 0, port)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestCheckFreePort_InvalidEndingPortShouldErr(t *testing.T) {
	t.Parallel()

	port, err := getPort("10000-NaN", checkFreePort)
	assert.Equal(t, 0, port)
	assert.Equal(t, p2p.ErrInvalidEndingPortValue, err)
}

func TestGetPort_EndPortLargerThanSendPort(t *testing.T) {
	t.Parallel()

	port, err := getPort("10000-9999", checkFreePort)
	assert.Equal(t, 0, port)
	assert.Equal(t, p2p.ErrEndPortIsSmallerThanStartPort, err)
}

func TestGetPort_RangeOfOneShouldWork(t *testing.T) {
	t.Parallel()

	port := 5000
	numCall := 0
	handler := func(p int) error {
		if p != port {
			assert.Fail(t, fmt.Sprintf("should have been %d", port))
		}
		numCall++
		return nil
	}

	result, err := getPort(fmt.Sprintf("%d-%d", port, port), handler)
	assert.Nil(t, err)
	assert.Equal(t, port, result)
}

func TestGetPort_RangeOccupiedShouldErrorShouldWork(t *testing.T) {
	t.Parallel()

	portStart := 5000
	portEnd := 10000
	portsTried := make(map[int]struct{})
	expectedErr := errors.New("expected error")
	handler := func(p int) error {
		portsTried[p] = struct{}{}
		return expectedErr
	}

	result, err := getPort(fmt.Sprintf("%d-%d", portStart, portEnd), handler)

	assert.True(t, errors.Is(err, p2p.ErrNoFreePortInRange))
	assert.Equal(t, portEnd-portStart+1, len(portsTried))
	assert.Equal(t, 0, result)
}

func TestCheckFreePort_PortZeroAlwaysWorks(t *testing.T) {
	err := checkFreePort(0)

	assert.Nil(t, err)
}

func TestCheckFreePort_InvalidPortShouldErr(t *testing.T) {
	err := checkFreePort(-1)

	assert.NotNil(t, err)
}

func TestCheckFreePort_OccupiedPortShouldErr(t *testing.T) {
	//1. get a free port from OS, open a TCP listner
	//2. get the allocated port
	//3. test if that port is occupied
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	port := l.Addr().(*net.TCPAddr).Port

	fmt.Printf("testing port %d\n", port)
	err = checkFreePort(port)
	assert.NotNil(t, err)

	_ = l.Close()
}
