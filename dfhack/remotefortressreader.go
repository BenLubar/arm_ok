package dfhack

import (
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
	"github.com/BenLubar/arm_ok/dfhack/dfproto"
)

var pluginRemoteFortressReader = "RemoteFortressReader"

// RPC GetGrowthList : EmptyMessage -> MaterialList
func (c *Conn) GetGrowthList() (*RemoteFortressReader.MaterialList, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply RemoteFortressReader.MaterialList
	text, err := c.RoundTripBind("GetGrowthList", &pluginRemoteFortressReader, "EmptyMessage", "MaterialList", &req, &reply)
	return &reply, text, err
}

// RPC GetMaterialList : EmptyMessage -> MaterialList
func (c *Conn) GetMaterialList() (*RemoteFortressReader.MaterialList, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply RemoteFortressReader.MaterialList
	text, err := c.RoundTripBind("GetMaterialList", &pluginRemoteFortressReader, "EmptyMessage", "MaterialList", &req, &reply)
	return &reply, text, err
}

// RPC GetTiletypeList : EmptyMessage -> TiletypeList
func (c *Conn) GetTiletypeList() (*RemoteFortressReader.TiletypeList, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply RemoteFortressReader.TiletypeList
	text, err := c.RoundTripBind("GetTiletypeList", &pluginRemoteFortressReader, "EmptyMessage", "TiletypeList", &req, &reply)
	return &reply, text, err
}

// RPC GetBlockList : BlockRequest -> BlockList
func (c *Conn) GetBlockList(req *RemoteFortressReader.BlockRequest) (*RemoteFortressReader.BlockList, []*dfproto.CoreTextNotification, error) {
	var reply RemoteFortressReader.BlockList
	text, err := c.RoundTripBind("GetBlockList", &pluginRemoteFortressReader, "BlockRequest", "BlockList", req, &reply)
	return &reply, text, err
}

// RPC GetPlantList : BlockRequest -> PlantList
func (c *Conn) GetPlantList(req *RemoteFortressReader.BlockRequest) (*RemoteFortressReader.PlantList, []*dfproto.CoreTextNotification, error) {
	var reply RemoteFortressReader.PlantList
	text, err := c.RoundTripBind("GetPlantList", &pluginRemoteFortressReader, "BlockRequest", "PlantList", req, &reply)
	return &reply, text, err
}

// RPC CheckHashes : EmptyMessage -> EmptyMessage
func (c *Conn) CheckHashes() ([]*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.EmptyMessage
	text, err := c.RoundTripBind("CheckHashes", &pluginRemoteFortressReader, "EmptyMessage", "EmptyMessage", &req, &reply)
	return text, err
}

// RPC GetUnitList : EmptyMessage -> UnitList
func (c *Conn) GetUnitList() (*RemoteFortressReader.UnitList, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply RemoteFortressReader.UnitList
	text, err := c.RoundTripBind("GetUnitList", &pluginRemoteFortressReader, "EmptyMessage", "UnitList", &req, &reply)
	return &reply, text, err
}

// RPC GetViewInfo : EmptyMessage -> ViewInfo
func (c *Conn) GetViewInfo() (*RemoteFortressReader.ViewInfo, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply RemoteFortressReader.ViewInfo
	text, err := c.RoundTripBind("GetViewInfo", &pluginRemoteFortressReader, "EmptyMessage", "ViewInfo", &req, &reply)
	return &reply, text, err
}

// RPC GetMapInfo : EmptyMessage -> MapInfo
func (c *Conn) GetMapInfo() (*RemoteFortressReader.MapInfo, []*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply RemoteFortressReader.MapInfo
	text, err := c.RoundTripBind("GetMapInfo", &pluginRemoteFortressReader, "EmptyMessage", "MapInfo", &req, &reply)
	return &reply, text, err
}

// RPC ResetMapHashes : EmptyMessage -> EmptyMessage
func (c *Conn) ResetMapHashes() ([]*dfproto.CoreTextNotification, error) {
	var req dfproto.EmptyMessage
	var reply dfproto.EmptyMessage
	text, err := c.RoundTripBind("ResetMapHashes", &pluginRemoteFortressReader, "EmptyMessage", "EmptyMessage", &req, &reply)
	return text, err
}
