// +build !js

//go:generate rm -f armok_web.js.gen.go armok_web.js.map.gen.go index.html.gen.go asset_dev.gen.go asset_nodev.gen.go
//go:generate go get github.com/gopherjs/gopherjs
//go:generate gopherjs build --tags=appengine
//go:generate go run asset.go -wrap handle -var _ armok_web.js
//go:generate go run asset.go -wrap handle -var _ armok_web.js.map
//go:generate go run asset.go -wrap handle -var index index.html

package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

var flagAddr = flag.String("addr", ":8050", "address to listen for HTTP connections on")

func handle(a asset) asset { http.Handle("/"+a.Name, a); return a }

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			index.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

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
