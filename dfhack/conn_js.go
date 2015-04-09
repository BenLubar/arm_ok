// +build js

package dfhack

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
)

func Connect() (*Conn, error) {
	return Dial(js.Global.Get("location").Get("host").String())
}

func Dial(addr string) (*Conn, error) {
	sock, err := websocket.Dial("ws://" + addr + "/ws")
	if err != nil {
		return nil, err
	}

	return (&Conn{sock: sock}).init()
}
