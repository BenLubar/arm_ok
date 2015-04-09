package dfhack

import "github.com/BenLubar/arm_ok/dfhack/dfproto"

// RPC GetWorldInfo : EmptyMessage -> GetWorldInfoOut
func (c *Conn) GetWorldInfo() (*dfproto.GetWorldInfoOut, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.GetWorldInfoOut
	text, err := c.RoundTripBind("GetWorldInfo", nil, "EmptyMessage", "GetWorldInfoOut", &req, &reply)
	return &reply, text, err
}

// RPC ListEnums : EmptyMessage -> ListEnumsOut
func (c *Conn) ListEnums() (*dfproto.ListEnumsOut, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.ListEnumsOut
	text, err := c.RoundTripBind("ListEnums", nil, "EmptyMessage", "ListEnumsOut", &req, &reply)
	return &reply, text, err
}

// RPC ListJobSkills : EmptyMessage -> ListJobSkillsOut
func (c *Conn) ListJobSkills() (*dfproto.ListJobSkillsOut, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.ListJobSkillsOut
	text, err := c.RoundTripBind("ListJobSkills", nil, "EmptyMessage", "ListJobSkillsOut", &req, &reply)
	return &reply, text, err
}

// RPC ListMaterials : ListMaterialsIn -> ListMaterialsOut
func (c *Conn) ListMaterials(req *dfproto.ListMaterialsIn) (*dfproto.ListMaterialsOut, []*dfproto.CoreTextNotification, error) {
	var reply dfproto.ListMaterialsOut
	text, err := c.RoundTripBind("ListMaterials", nil, "ListMaterialsIn", "ListMaterialsOut", req, &reply)
	return &reply, text, err
}

// RPC ListUnits : ListUnitsIn -> ListUnitsOut
func (c *Conn) ListUnits(req *dfproto.ListUnitsIn) (*dfproto.ListUnitsOut, []*dfproto.CoreTextNotification, error) {
	var reply dfproto.ListUnitsOut
	text, err := c.RoundTripBind("ListUnits", nil, "ListUnitsIn", "ListUnitsOut", req, &reply)
	return &reply, text, err
}

// RPC ListSquads : ListSquadsIn -> ListSquadsOut
func (c *Conn) ListSquads(req *dfproto.ListSquadsIn) (*dfproto.ListSquadsOut, []*dfproto.CoreTextNotification, error) {
	var reply dfproto.ListSquadsOut
	text, err := c.RoundTripBind("ListSquads", nil, "ListSquadsIn", "ListSquadsOut", req, &reply)
	return &reply, text, err
}

// RPC SetUnitLabors : SetUnitLaborsIn -> EmptyMessage
func (c *Conn) SetUnitLabors(req *dfproto.SetUnitLaborsIn) ([]*dfproto.CoreTextNotification, error) {
	var reply dfproto.EmptyMessage
	text, err := c.RoundTripBind("SetUnitLabors", nil, "SetUnitLaborsIn", "EmptyMessage", req, &reply)
	return text, err
}
