package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
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

	hashes map[[3]int32]uint32
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
	if ok, err1 := ctx.RespondPartial(text, err); ok {
		return ctx.WriteMessage(resp)
	} else {
		return err1
	}
}

func (ctx *proxy_ctx) RespondPartial(text []*dfproto.CoreTextNotification, err error) (bool, error) {
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
		return false, err
	}

	for _, t := range text {
		if err := ctx.WriteText(t); err != nil {
			return false, err
		}
	}

	if errno == 0 {
		return true, nil
	}
	return false, ctx.WriteError(errno, "")
}

type MapBlock struct {
	Block *RemoteFortressReader.MapBlock
	Hash  uint32
}

var (
	pluginRemoteFortressReader = "RemoteFortressReader"

	mapBlocks = make(map[[3]int32]MapBlock)
	mapLock   sync.Mutex
	mapOnce   sync.Once

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

				return ctx.WriteMessage(RemoteVersion)
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

				return ctx.WriteMessage(RemoteDFVersion)
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

				resp, text, err := Remote.GetWorldInfo()
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

				resp, text, err := Remote.ListEnums()
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

				resp, text, err := Remote.ListJobSkills()
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

				resp, text, err := Remote.ListMaterials(&req)
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "GetGrowthList",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "RemoteFortressReader.MaterialList",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetGrowthList()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "GetMaterialList",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "RemoteFortressReader.MaterialList",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetMaterialList()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "GetTiletypeList",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "RemoteFortressReader.TiletypeList",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetTiletypeList()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "ResetMapHashes",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "dfproto.EmptyMessage",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				ctx.hashes = make(map[[3]int32]uint32)

				return ctx.WriteMessage(&dfproto.EmptyMessage{})
			},
		},
		{
			Command: "GetBlockList",
			Plugin:  &pluginRemoteFortressReader,
			In:      "RemoteFortressReader.BlockRequest",
			Out:     "RemoteFortressReader.BlockList",
			Handle: func(ctx *proxy_ctx) error {
				var req RemoteFortressReader.BlockRequest
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				mapLock.Lock()
				defer mapLock.Unlock()

				limit := req.BlocksNeeded
				req.BlocksNeeded = nil
				resp, text, err := Remote.GetBlockList(&req)
				if ok, err1 := ctx.RespondPartial(text, err); !ok {
					return err1
				}
				req.BlocksNeeded = limit

				for _, block := range resp.MapBlocks {
					pos := [3]int32{block.GetMapX() / 16, block.GetMapY() / 16, block.GetMapZ()}
					if old, ok := mapBlocks[pos]; ok {
						if len(block.Magma) == 0 {
							block.Magma = old.Block.Magma
						}
						if len(block.Water) == 0 {
							block.Water = old.Block.Water
						}
						if len(block.Tiles) == 0 {
							block.Tiles = old.Block.Tiles
						}
						if len(block.Materials) == 0 {
							block.Materials = old.Block.Materials
						}
						if len(block.LayerMaterials) == 0 {
							block.LayerMaterials = old.Block.LayerMaterials
						}
						if len(block.VeinMaterials) == 0 {
							block.VeinMaterials = old.Block.VeinMaterials
						}
						if len(block.BaseMaterials) == 0 {
							block.BaseMaterials = old.Block.BaseMaterials
						}
					}
					b, err := proto.Marshal(block)
					if err != nil {
						// should never happen
						panic(err)
					}
					mapBlocks[pos] = MapBlock{
						Block: block,
						Hash:  adler32.Checksum(b),
					}
				}

				resp1 := &RemoteFortressReader.BlockList{
					MapX: resp.MapX,
					MapY: resp.MapY,
				}

				min_x, max_x := req.GetMinX(), req.GetMaxX()
				min_y, max_y := req.GetMinY(), req.GetMaxY()
				min_z, max_z := req.GetMinZ(), req.GetMaxZ()
				center_x := (min_x + max_x) / 2
				center_y := (min_y + max_y) / 2
				number_of_points := ((max_x - center_x + 1) * 2) * ((max_y - center_y + 1) * 2)
				var blocks_needed, blocks_sent int32
				if req.BlocksNeeded != nil {
					blocks_needed = req.GetBlocksNeeded()
				} else {
					blocks_needed = number_of_points * (max_z - min_z)
				}
				for zz := max_z - 1; zz >= min_z; zz-- {
					// (di, dj) is a vector - direction in which we move right now
					var di, dj int32 = 1, 0
					// length of current segment
					var segment_length, segment_passed int32 = 1, 0
					// current position (i, j) and how much of current segment we passed
					var i, j, k int32 = center_x, center_y, 0
					for k = 0; k < number_of_points; k++ {
						if blocks_sent >= blocks_needed {
							break
						}
						if i >= min_x && i < max_x && j >= min_y && j < max_y {
							pos := [3]int32{i, j, zz}
							if block, ok := mapBlocks[pos]; ok {
								// if ctx.hashes is nil, the client hasn't called ResetMapHashes yet, so whether the block gets sent or not is undefined. We take the easy route of only needing one conditional.
								if hash, ok := ctx.hashes[pos]; ctx.hashes != nil && (!ok || hash != block.Hash) {
									resp1.MapBlocks = append(resp1.MapBlocks, block.Block)
									ctx.hashes[pos] = block.Hash
									blocks_sent++
								}
							}
						}

						// make a step, add 'direction' vector (di, dj) to current position (i, j)
						i += di
						j += dj
						segment_passed++

						if segment_passed == segment_length {
							// done with current segment
							segment_passed = 0

							// 'rotate' directions
							di, dj = -dj, di

							// increase segment length if necessary
							if dj == 0 {
								segment_length++
							}
						}
					}
				}

				return ctx.WriteMessage(resp1)
			},
		},
		{
			Command: "GetPlantList",
			Plugin:  &pluginRemoteFortressReader,
			In:      "RemoteFortressReader.BlockRequest",
			Out:     "RemoteFortressReader.PlantList",
			Handle: func(ctx *proxy_ctx) error {
				var req RemoteFortressReader.BlockRequest
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetPlantList(&req)
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "GetUnitList",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "RemoteFortressReader.UnitList",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetUnitList()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "GetViewInfo",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "RemoteFortressReader.ViewInfo",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetViewInfo()
				return ctx.Respond(resp, text, err)
			},
		},
		{
			Command: "GetMapInfo",
			Plugin:  &pluginRemoteFortressReader,
			In:      "dfproto.EmptyMessage",
			Out:     "RemoteFortressReader.MapInfo",
			Handle: func(ctx *proxy_ctx) error {
				var req dfproto.EmptyMessage
				if err := ctx.ReadMessage(&req); err != nil {
					return err
				}

				resp, text, err := Remote.GetMapInfo()
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

var remoteOnce sync.Once
var Remote *dfhack.Conn
var RemoteVersion *dfproto.StringMessage
var RemoteDFVersion *dfproto.StringMessage

func remote() {
	var err error
	Remote, err = dfhack.Connect()
	if err != nil {
		log.Panicln("remote connect:", err)
	}

	version, _, err := Remote.GetVersion()
	if err != nil {
		log.Panicln("remote GetVersion:", err)
	}
	RemoteVersion = &dfproto.StringMessage{Value: proto.String(version)}

	version, _, err = Remote.GetDFVersion()
	if err != nil {
		log.Panicln("remote GetDFVersion:", err)
	}
	RemoteDFVersion = &dfproto.StringMessage{Value: proto.String(version)}

	_, err = Remote.ResetMapHashes()
	if err != nil {
		log.Panicln("remote ResetMapHashes:", err)
	}
}

func proxy(in *websocket.Conn) {
	in.PayloadType = websocket.BinaryFrame

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

	remoteOnce.Do(remote)

	var buf bytes.Buffer
	ctx := &proxy_ctx{
		w: in,
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
