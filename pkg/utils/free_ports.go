package utils

import (
	"net"
	"strconv"
)

// FreePorts asks the kernel for n distinct free TCP ports. The listeners are
// closed before returning, so there is a small window in which another process
// could grab a port, but ports are not reused immediately in practice. Used by
// tests to avoid fixed port numbers so packages can run in parallel.
func FreePorts(n int) ([]int, error) {
	listeners := make([]net.Listener, 0, n)
	ports := make([]int, 0, n)

	for i := 0; i < n; i++ {
		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			for _, open := range listeners {
				open.Close()
			}
			return nil, err
		}
		listeners = append(listeners, l)
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}

	for _, l := range listeners {
		l.Close()
	}

	return ports, nil
}

// FreePort returns a single free TCP port.
func FreePort() (int, error) {
	ports, err := FreePorts(1)
	if err != nil {
		return 0, err
	}
	return ports[0], nil
}

// FreePortsOrPanic is a convenience wrapper for test setup code without error
// returns.
func FreePortsOrPanic(n int) []int {
	ports, err := FreePorts(n)
	if err != nil {
		panic("failed to allocate free ports: " + err.Error() + " (n=" + strconv.Itoa(n) + ")")
	}
	return ports
}

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
		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			for _, r := range reservations {
				r.Release()
			}
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
