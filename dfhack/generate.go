//go:generate git submodule init
//go:generate git submodule update
//go:generate go get github.com/golang/protobuf/protoc-gen-go
//go:generate go generate ./RemoteFortressReader
//go:generate go generate ./dfproto

package dfhack
