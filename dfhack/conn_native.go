// +build !js

package dfhack

import (
	"net"
	"os"
	"strconv"
)

func Connect() (*Conn, error) {
	port, err := strconv.Atoi(os.Getenv("DFHACK_PORT"))
	if err != nil {
		port = 5000
	}

	return Dial(net.JoinHostPort("localhost", strconv.Itoa(port)))
}

func Dial(addr string) (*Conn, error) {
	sock, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return (&Conn{sock: sock}).init()
}
