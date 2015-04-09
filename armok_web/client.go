// +build js

package main

import (
	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	conn, err := dfhack.Connect()
	if err != nil {
		panic(err)
	}

	js.Global.Set("dfhack", conn)
}
