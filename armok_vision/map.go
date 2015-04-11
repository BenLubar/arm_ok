package main

import (
	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
	"github.com/golang/protobuf/proto"
)

type (
	MapBlock [16][16]MapTile
	MapTile  struct {
		Tiletype Tiletype
		Material Material
		Base     Material
		Layer    Material
		Vein     Material
		Water    uint8 // [0, 7]
		Magma    uint8 // [0, 7]
	}
)

func InitMap(conn *dfhack.Conn) {
	_, err := conn.ResetMapHashes()
	if err != nil {
		panic(err)
	}
}

var Map = make(map[[3]int32]*MapBlock)
var Dirty [][3]int32

const rangeX = 4
const rangeY = 3
const rangeZup = 0
const rangeZdown = 5

func UpdateMap(conn *dfhack.Conn) {
	posX := (ViewInfo.GetViewPosX() + (ViewInfo.GetViewSizeX() / 2)) / 16
	posY := (ViewInfo.GetViewPosY() + (ViewInfo.GetViewSizeY() / 2)) / 16
	posZ := ViewInfo.GetViewPosZ()

	blocks, _, err := conn.GetBlockList(&RemoteFortressReader.BlockRequest{
		MinX: proto.Int32(posX - rangeX),
		MaxX: proto.Int32(posX + rangeX),
		MinY: proto.Int32(posY - rangeY),
		MaxY: proto.Int32(posY + rangeY),
		MinZ: proto.Int32(posZ - rangeZdown),
		MaxZ: proto.Int32(posZ + rangeZup),

		BlocksNeeded: proto.Int32(1),
	})
	if err != nil {
		panic(err)
	}

	for _, block := range blocks.MapBlocks {
		pos := [3]int32{block.GetMapX() / 16, block.GetMapY() / 16, block.GetMapZ()}
		Dirty = append(Dirty, pos)
		tiles, ok := Map[pos]
		if !ok {
			tiles = new(MapBlock)
			Map[pos] = tiles
		}
		for i, tt := range block.Tiles {
			tiles[i%16][i/16].Tiletype = Tiletype(tt)
		}
		for i, mat := range block.Materials {
			tiles[i%16][i/16].Material = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
		}
		for i, mat := range block.BaseMaterials {
			tiles[i%16][i/16].Base = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
		}
		for i, mat := range block.LayerMaterials {
			tiles[i%16][i/16].Layer = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
		}
		for i, mat := range block.VeinMaterials {
			tiles[i%16][i/16].Vein = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
		}
		for i, w := range block.Water {
			tiles[i%16][i/16].Water = uint8(w)
		}
		for i, m := range block.Magma {
			tiles[i%16][i/16].Magma = uint8(m)
		}
	}
}
