package utils

import (
	"net"
	"strconv"
)

// Ports are probed on the IPv4 wildcard address because that is where etcd,
// the relay server and the HTTP API bind. Probing localhost only would miss
// ports that are already in use on another interface.
const portProbeNetwork = "tcp4"
const portProbeAddr = "0.0.0.0:0"

// ReservedPort is a free TCP port that stays bound until Release is called, so
// no other process can be handed the same port in the meantime. Test setups
// that do slow work between allocating a port and binding it (starting etcd,
// preparing a database) should hold a reservation and release it just before
// the component binds, keeping the unbound window to microseconds.
type ReservedPort struct {
	listener net.Listener
	port     int
}

// Port returns the reserved port number. The port stays bound until Release.
func (r *ReservedPort) Port() int {
	return r.port
}

// Release closes the reservation so the port can be bound. Safe to call more
// than once.
func (r *ReservedPort) Release() {
	if r.listener != nil {
		r.listener.Close()
		r.listener = nil
	}
}

// ReservePorts asks the kernel for n distinct free TCP ports and keeps them
// bound until each reservation is released.
func ReservePorts(n int) ([]*ReservedPort, error) {
	reservations := make([]*ReservedPort, 0, n)

	for i := 0; i < n; i++ {
		l, err := net.Listen(portProbeNetwork, portProbeAddr)
		if err != nil {
			ReleasePorts(reservations)
			return nil, err
		}
		reservations = append(reservations, &ReservedPort{listener: l, port: l.Addr().(*net.TCPAddr).Port})
	}

	return reservations, nil
}

// ReservePortsOrPanic is a convenience wrapper for test setup code without
// error returns.
func ReservePortsOrPanic(n int) []*ReservedPort {
	reservations, err := ReservePorts(n)
	if err != nil {
		panic("failed to reserve free ports: " + err.Error() + " (n=" + strconv.Itoa(n) + ")")
	}
	return reservations
}

// ReleasePorts releases every reservation in the slice.
func ReleasePorts(reservations []*ReservedPort) {
	for _, r := range reservations {
		r.Release()
	}
}
