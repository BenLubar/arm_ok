// +build !js

package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

var flagAddr = flag.String("addr", ":8050", "address to listen for HTTP connections on")

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *flagAddr)
	if err != nil {
		log.Fatalln("listening failed:", err)
	}
	defer l.Close()

	log.Printf("listening on http://%v/", l.Addr())

	log.Fatalln(http.Serve(l, nil))
}
