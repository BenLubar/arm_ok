// +build js

package main

import (
	"fmt"

	"github.com/BenLubar/arm_ok/dfhack"
)

func main() {
	conn, err := dfhack.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	version, _, err := conn.GetDFVersion()
	if err != nil {
		panic(err)
	}
	fmt.Println(version)
}
