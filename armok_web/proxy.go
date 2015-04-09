// +build !js

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/dfproto"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/websocket"
)

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
	Pad  [2]byte
	Size int32
}

const (
	cr_link_failure    = -3
	cr_needs_console   = -2
	cr_not_implemented = -1
	cr_failure         = 1
	cr_wrong_usage     = 2
	cr_not_found       = 3
)

type proxy_ctx struct {
	b []byte
	w io.Writer

	Conn *dfhack.Conn
}

func (ctx *proxy_ctx) ReadMessage(req proto.Message) error {
	return proto.Unmarshal(ctx.b, req)
}

func (ctx *proxy_ctx) writeHeader(id int16, size int32) error {
	return binary.Write(ctx.w, binary.LittleEndian, &rpcMessageHeader{
		ID:   id,
		Size: size,
	})
}

func (ctx *proxy_ctx) WriteText(text *dfproto.CoreTextNotification) error {
	b, err := proto.Marshal(text)
	if err != nil {
		return err
	}

	if err := ctx.writeHeader(rpcReplyText, int32(len(b))); err != nil {
		return err
	}

	n, err := ctx.w.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}

	return err
}

func (ctx *proxy_ctx) WriteError(code int32, message string) error {
	if message != "" {
		if err := ctx.WriteText(&dfproto.CoreTextNotification{Fragments: []*dfproto.CoreTextFragment{{
			Text:  proto.String(message),
			Color: dfproto.CoreTextFragment_COLOR_LIGHTRED.Enum(),
		}}}); err != nil {
			return err
		}
	}

	return ctx.writeHeader(rpcReplyFail, code)
}

func (ctx *proxy_ctx) WriteMessage(resp proto.Message) error {
	b, err := proto.Marshal(resp)
	if err != nil {
		return err
	}

	if len(b) > int(maxMessageSize) {
		return ctx.WriteError(cr_link_failure, fmt.Sprintf("reply too large: %d", len(b)))
	}

	if err := ctx.writeHeader(rpcReplyResult, int32(len(b))); err != nil {
		return err
	}

	n, err := ctx.w.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}

	return err
}

func (ctx *proxy_ctx) Respond(resp proto.Message, text []*dfproto.CoreTextNotification, err error) error {
	var errno int32
	switch err {
	case dfhack.ErrLinkFailure:
		errno = cr_link_failure
	case dfhack.ErrNeedsConsole:
		errno = cr_needs_console
	case dfhack.ErrNotImplemented:
		errno = cr_not_implemented
	case dfhack.ErrFailure:
		errno = cr_failure
	case dfhack.ErrWrongUsage:
		errno = cr_wrong_usage
	case dfhack.ErrNotFound:
		errno = cr_not_found
	case nil:
	default:
		return err
	}

	for _, t := range text {
		if err := ctx.WriteText(t); err != nil {
			return err
		}
	}

	if errno == 0 {
		return ctx.WriteMessage(resp)
	}
	return ctx.WriteError(errno, "")
}

var (
	pluginRemoteFortressReader = "RemoteFortressReader"

	CoreMessages    = make(map[string]int16)
	PluginMessages  = make(map[string]map[string]int16)
	AllowedMessages = []struct {
		Command string
		Plugin  *string
		In, Out string
		Handle  func(*proxy_ctx) error
	}{
		{
			Command: "BindMethod",
			In:      "dfproto.CoreBindRequest",
			Out:     "dfproto.CoreBindReply",
			Handle:  nil, // assigned below
		},
		{
			Command: "RunCommand",
			In:      "dfproto.CoreRunCommandRequest",
			Out:     "dfproto.EmptyMessage",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.CoreRunCommandRequest
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				return ctx.WriteError(cr_not_implemented, "")
			},
		},
		{
			Command: "GetVersion",
			In:      "dfproto.EmptyMessage",
			Out:     "dfproto.StringMessage",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := ctx.Conn.GetVersion()
				return ctx.Respond(&dfproto.StringMessage{
					Value: proto.String(resp),
				}, text, err)
			},
		},
		{
			Command: "GetDFVersion",
			In:      "dfproto.EmptyMessage",
			Out:     "dfproto.StringMessage",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := ctx.Conn.GetDFVersion()
				return ctx.Respond(&dfproto.StringMessage{
					Value: proto.String(resp),
				}, text, err)
			},
		},
		{
			Command: "GetWorldInfo",
			In:      "dfproto.EmptyMessage",
			Out:     "dfproto.GetWorldInfoOut",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := ctx.Conn.GetWorldInfo()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "ListEnums",
			In:      "dfproto.EmptyMessage",
			Out:     "dfproto.ListEnumsOut",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := ctx.Conn.ListEnums()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "ListJobSkills",
			In:      "dfproto.EmptyMessage",
			Out:     "dfproto.ListJobSkillsOut",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := ctx.Conn.ListJobSkills()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "ListMaterials",
			In:      "dfproto.ListMaterialsIn",
			Out:     "dfproto.ListMaterialsOut",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.ListMaterialsIn
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := ctx.Conn.ListMaterials(&req)
				return ctx.Respond(resp, text, err)
			},
		},
	}
)

func init() {
	AllowedMessages[0].Handle = func(ctx *proxy_ctx) error {
		var req dfproto.CoreBindRequest
		if err := ctx.ReadMessage(&req); err != nil {
			return err
		}

		var id int16
		var ok bool
		if req.Plugin == nil {
			id, ok = CoreMessages[req.GetMethod()]
		} else {
			id, ok = PluginMessages[req.GetPlugin()][req.GetMethod()]
		}

		if !ok {
			return ctx.WriteError(cr_failure, fmt.Sprintf("RPC method not found: %s::%s\n", req.GetPlugin(), req.GetMethod()))
		}

		msg := AllowedMessages[id]

		if msg.In != req.GetInputMsg() || msg.Out != req.GetOutputMsg() {
			return ctx.WriteError(cr_failure, fmt.Sprintf("Requested wrong signature for RPC method: %s::%s (%q -> %q, %q -> %q)\n", req.GetPlugin(), req.GetMethod(), req.GetInputMsg(), msg.In, req.GetOutputMsg(), msg.Out))
		}

		return ctx.WriteMessage(&dfproto.CoreBindReply{
			AssignedId: proto.Int32(int32(id)),
		})
	}

	for i, msg := range AllowedMessages {
		if msg.Plugin == nil {
			CoreMessages[msg.Command] = int16(i)
		} else {
			m, ok := PluginMessages[*msg.Plugin]
			if !ok {
				m = make(map[string]int16)
				PluginMessages[*msg.Plugin] = m
			}
			m[msg.Command] = int16(i)
		}
	}
}

func proxy(in *websocket.Conn) {
	addr := in.Request().RemoteAddr
	defer in.Close()

	var handshake rpcHandshakeHeader
	err := binary.Read(in, binary.LittleEndian, &handshake)
	if err != nil {
		log.Println(addr, "reading handshake:", err)
		return
	}
	if handshake.Magic != rpcMagicRequest || handshake.Version != rpcVersion {
		log.Println(addr, "invalid handshake")
		return
	}
	err = binary.Write(in, binary.LittleEndian, &rpcHandshakeHeader{
		Magic:   rpcMagicResponse,
		Version: rpcVersion,
	})
	if err != nil {
		log.Println(addr, "writing handshake:", err)
		return
	}

	log.Println(addr, "connect")

	out, err := dfhack.Connect()
	if err != nil {
		log.Println(addr, "remote connect:", err)
		return
	}
	defer out.Close()

	var buf bytes.Buffer
	ctx := &proxy_ctx{
		w:    in,
		Conn: out,
	}

	for {
		var header rpcMessageHeader
		err = binary.Read(in, binary.LittleEndian, &header)
		if err != nil {
			log.Println(addr, "reading header:", err)
			return
		}
		if header.ID == rpcRequestQuit {
			log.Println(addr, "disconnect")
			return
		}
		if header.Size < 0 || header.Size > maxMessageSize {
			log.Println(addr, "invalid received size:", header.Size)
			return
		}

		buf.Reset()
		n, err := io.CopyN(&buf, in, int64(header.Size))
		if err != nil {
			log.Println(addr, "reading data:", n, "/", header.Size, err)
			return
		}
		ctx.b = buf.Bytes()

		if header.ID < 0 || header.ID >= int16(len(AllowedMessages)) {
			err = ctx.WriteError(cr_not_found, fmt.Sprintf("RPC call of invalid id %d\n", header.ID))
		} else {
			err = AllowedMessages[header.ID].Handle(ctx)
		}
		if err != nil {
			log.Println(addr, "writing response:", err)
			return
		}
	}
}

func init() {
	http.Handle("/ws", websocket.Handler(proxy))
}
