package dfhack

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/BenLubar/arm_ok/dfhack/dfproto"
	"github.com/golang/protobuf/proto"
)

var (
	ErrInvalidHandshake = errors.New("dfhack: invalid handshake")
	ErrMessageTooLarge  = errors.New("dfhack: message too large")
	ErrInvalidError     = errors.New("dfhack: error code unknown")

	ErrLinkFailure    = errors.New("dfhack: CR_LINK_FAILURE: RPC call failed due to I/O or protocol error")
	ErrNeedsConsole   = errors.New("dfhack: CR_NEEDS_CONSOLE: attempt to call interactive command without console")
	ErrNotImplemented = errors.New("dfhack: CR_NOT_IMPLEMENTED: command not implemented, or plugin not loaded")
	ErrFailure        = errors.New("dfhack: CR_FAILURE: failure")
	ErrWrongUsage     = errors.New("dfhack: CR_WRONG_USAGE: wrong arguments or ui state")
	ErrNotFound       = errors.New("dfhack: CR_NOT_FOUND: target object not found (for RPC mainly)")
)
var knownErrors = map[int32]error{
	-3: ErrLinkFailure,
	-2: ErrNeedsConsole,
	-1: ErrNotImplemented,
	// CR_SUCCESS is treated as unknown
	1: ErrFailure,
	2: ErrWrongUsage,
	3: ErrNotFound,
}

type Conn struct {
	sock   io.ReadWriteCloser
	bound  map[[3]string]int16
	plugin map[string]map[[3]string]int16
	mtx    sync.Mutex
}

var (
	rpcMagicRequest  = [8]byte{'D', 'F', 'H', 'a', 'c', 'k', '?', '\n'}
	rpcMagicResponse = [8]byte{'D', 'F', 'H', 'a', 'c', 'k', '!', '\n'}
)

const rpcVersion int32 = 1

type rpcHandshakeHeader struct {
	Magic   [8]byte
	Version int32
}

const maxMessageSize int32 = 8 * 1048576

const (
	rpcReplyResult int16 = -1
	rpcReplyFail   int16 = -2
	rpcReplyText   int16 = -3
	rpcRequestQuit int16 = -4
)

type rpcMessageHeader struct {
	ID   int16
	Size int32
}

// 1. Handshake
//
//   Client initiates connection by sending the handshake
//   request header. The server responds with the response
//   magic. Currently both versions must be 1.
//
func (c *Conn) init() (self *Conn, err error) {
	self = c
	defer func() {
		if err != nil {
			// clean up if we fail to initialize.
			c.Close()
			self = nil
		}
	}()

	c.bound = make(map[[3]string]int16)
	c.plugin = make(map[string]map[[3]string]int16)

	err = binary.Write(c.sock, binary.LittleEndian, &rpcHandshakeHeader{
		Magic:   rpcMagicRequest,
		Version: rpcVersion,
	})
	if err != nil {
		return
	}

	var response rpcHandshakeHeader
	err = binary.Read(c.sock, binary.LittleEndian, &response)
	if err != nil {
		return
	}

	if response.Magic != rpcMagicResponse || response.Version != rpcVersion {
		err = ErrInvalidHandshake
	}

	return
}

// 2. Interaction
//
//   Requests are done by exchanging messages between the
//   client and the server. Messages consist of a serialized
//   protobuf message preceeded by RPCMessageHeader. The size
//   field specifies the length of the protobuf part.
//
//   NOTE: As a special exception, RPC_REPLY_FAIL uses the size
//         field to hold the error code directly.
//
//   Every callable function is assigned a non-negative id by
//   the server. Id 0 is reserved for BindMethod, which can be
//   used to request any other id by function name. Id 1 is
//   RunCommand, used to call console commands remotely.
//
//   The client initiates every call by sending a message with
//   appropriate function id and input arguments. The server
//   responds with zero or more RPC_REPLY_TEXT:CoreTextNotification
//   messages, followed by RPC_REPLY_RESULT containing the output
//   of the function if it succeeded, or RPC_REPLY_FAIL with the
//   error code if it did not.
//
func (c *Conn) roundTrip(id int16, req, resp proto.Message) ([]*dfproto.CoreTextNotification, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	if len(b) > int(maxMessageSize) {
		return nil, ErrMessageTooLarge
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	err = binary.Write(c.sock, binary.LittleEndian, &rpcMessageHeader{
		ID:   id,
		Size: int32(len(b)),
	})
	if err != nil {
		return nil, err
	}

	n, err := c.sock.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	if err != nil {
		return nil, err
	}

	var text []*dfproto.CoreTextNotification

	for {
		var header rpcMessageHeader
		err = binary.Read(c.sock, binary.LittleEndian, &header)
		if err != nil {
			return text, err
		}

		switch header.ID {
		case rpcReplyResult:
			b := make([]byte, header.Size)
			_, err = io.ReadFull(c.sock, b)
			if err != nil {
				return text, err
			}

			return text, proto.Unmarshal(b, resp)

		case rpcReplyFail:
			if err, ok := knownErrors[header.Size]; ok {
				return text, err
			}
			return text, ErrInvalidError

		case rpcReplyText:
			var message dfproto.CoreTextNotification
			b := make([]byte, header.Size)
			_, err = io.ReadFull(c.sock, b)
			if err != nil {
				return text, err
			}

			err = proto.Unmarshal(b, &message)
			if err != nil {
				return text, err
			}
			text = append(text, &message)
		}
	}
}

// 3. Disconnect
//
//   The client terminates the connection by sending an
//   RPC_REQUEST_QUIT header with zero size and immediately
//   closing the socket.
//
func (c *Conn) Close() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	_ = binary.Write(c.sock, binary.LittleEndian, &rpcMessageHeader{
		ID:   rpcRequestQuit,
		Size: 0,
	})

	return c.sock.Close()
}
