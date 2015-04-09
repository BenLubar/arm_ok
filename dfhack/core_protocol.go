package dfhack

import (
	"github.com/BenLubar/arm_ok/dfhack/dfproto"
	"github.com/golang/protobuf/proto"
)

func (c *Conn) RoundTripBind(command string, plugin *string, in, out string, req, resp proto.Message) ([]*dfproto.CoreTextNotification, error) {
	var id int16
	var ok bool
	key := [3]string{command, in, out}

	c.mtx.Lock()
	if plugin == nil {
		id, ok = c.bound[key]
	} else {
		id, ok = c.plugin[*plugin][key]
	}
	c.mtx.Unlock()

	if ok {
		return c.roundTrip(id, req, resp)
	}

	bind, text, err := c.BindMethod(&dfproto.CoreBindRequest{
		Method:    &command,
		Plugin:    plugin,
		InputMsg:  &in,
		OutputMsg: &out,
	})
	if err != nil {
		return text, err
	}

	id = int16(bind)

	c.mtx.Lock()
	if plugin == nil {
		c.bound[key] = id
	} else {
		if c.plugin[*plugin] == nil {
			c.plugin[*plugin] = make(map[[3]string]int16)
		}
		c.plugin[*plugin][key] = id
	}
	c.mtx.Unlock()

	text2, err := c.roundTrip(id, req, resp)
	// Don't call append if there's a chance both slices are nil.
	if text == nil {
		return text2, err
	}
	return append(text, text2...), err
}

// RPC BindMethod : CoreBindRequest -> CoreBindReply
func (c *Conn) BindMethod(req *dfproto.CoreBindRequest) (int32, []*dfproto.CoreTextNotification, error) {
	var reply dfproto.CoreBindReply
	text, err := c.roundTrip(0, req, &reply)
	return reply.GetAssignedId(), text, err
}

// RPC RunCommand : CoreRunCommandRequest -> EmptyMessage
func (c *Conn) RunCommand(req *dfproto.CoreRunCommandRequest) ([]*dfproto.CoreTextNotification, error) {
	var reply dfproto.EmptyMessage
	text, err := c.RoundTripBind("RunCommand", nil, "CoreRunCommandRequest", "EmptyMessage", req, &reply)
	return text, err
}

// RPC CoreSuspend : EmptyMessage -> IntMessage
func (c *Conn) CoreSuspend() (int32, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.IntMessage
	text, err := c.RoundTripBind("CoreSuspend", nil, "EmptyMessage", "IntMessage", &req, &reply)
	return reply.GetValue(), text, err
}

// RPC CoreResume : EmptyMessage -> IntMessage
func (c *Conn) CoreResume() (int32, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.IntMessage
	text, err := c.RoundTripBind("CoreResume", nil, "EmptyMessage", "IntMessage", &req, &reply)
	return reply.GetValue(), text, err
}

// RPC RunLua : CoreRunLuaRequest -> StringListMessage
func (c *Conn) RunLua(req *dfproto.CoreRunLuaRequest) ([]string, []*dfproto.CoreTextNotification, error) {
	var reply dfproto.StringListMessage
	text, err := c.RoundTripBind("RunLua", nil, "CoreRunLuaRequest", "StringListMessage", req, &reply)
	return reply.GetValue(), text, err
}
