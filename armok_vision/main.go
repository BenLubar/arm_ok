package main

import (
	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
	"github.com/golang/protobuf/proto"
)

func main() {
	if err := InitGL(); err != nil {
		panic(err)
	}
	defer CleanupGL()

	conn, err := dfhack.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.ResetMapHashes()

	for {
		UpdateViewInfo(conn)
		UpdateBlockList(conn)
		PositionCamera()
		UseBlockList()

		// this would eventually be a render loop, but not yet.
		break
	}
}

var ViewInfo *RemoteFortressReader.ViewInfo

func UpdateViewInfo(conn *dfhack.Conn) {
	var err error
	ViewInfo, _, err = conn.GetViewInfo()
	if err != nil {
		panic(err)
	}
}

var BlockList *RemoteFortressReader.BlockList

func UpdateBlockList(conn *dfhack.Conn) {
	posX := (ViewInfo.GetViewPosX() + (ViewInfo.GetViewSizeX() / 2)) / 16
	posY := (ViewInfo.GetViewPosY() + (ViewInfo.GetViewSizeY() / 2)) / 16
	posZ := ViewInfo.GetViewPosZ()

	const rangeX = 4
	const rangeY = 3
	const rangeZup = 2
	const rangeZdown = 5

	var err error
	BlockList, _, err = conn.GetBlockList(&RemoteFortressReader.BlockRequest{
		MinX: proto.Int32(posX - rangeX),
		MaxX: proto.Int32(posX + rangeX),
		MinY: proto.Int32(posY - rangeY),
		MaxY: proto.Int32(posY + rangeY),
		MinZ: proto.Int32(posZ - rangeZdown),
		MaxZ: proto.Int32(posZ + rangeZup),
	})
	if err != nil {
		panic(err)
	}
}
