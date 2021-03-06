package main

import (
	"sync"

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

var NotLoadedData = []float32{
	// Bottom
	0, 0, 0, 0.2, 0.3, 0.8, 0, 0, -1,
	0, 16, 0, 0.2, 0.3, 0.8, 0, 0, -1,
	16, 0, 0, 0.2, 0.3, 0.8, 0, 0, -1,
	16, 0, 0, 0.2, 0.3, 0.8, 0, 0, -1,
	0, 16, 0, 0.2, 0.3, 0.8, 0, 0, -1,
	16, 16, 0, 0.2, 0.3, 0.8, 0, 0, -1,

	// Top
	0, 0, 1, 0.2, 0.3, 0.8, 0, 0, 1,
	16, 0, 1, 0.2, 0.3, 0.8, 0, 0, 1,
	0, 16, 1, 0.2, 0.3, 0.8, 0, 0, 1,
	16, 0, 1, 0.2, 0.3, 0.8, 0, 0, 1,
	16, 16, 1, 0.2, 0.3, 0.8, 0, 0, 1,
	0, 16, 1, 0.2, 0.3, 0.8, 0, 0, 1,

	// Front
	0, 16, 0, 0.2, 0.3, 0.8, 0, -1, 0,
	0, 16, 1, 0.2, 0.3, 0.8, 0, -1, 0,
	16, 16, 0, 0.2, 0.3, 0.8, 0, -1, 0,
	16, 16, 0, 0.2, 0.3, 0.8, 0, -1, 0,
	0, 16, 1, 0.2, 0.3, 0.8, 0, -1, 0,
	16, 16, 1, 0.2, 0.3, 0.8, 0, -1, 0,

	// Back
	0, 0, 0, 0.2, 0.3, 0.8, 0, 1, 0,
	16, 0, 0, 0.2, 0.3, 0.8, 0, 1, 0,
	0, 0, 1, 0.2, 0.3, 0.8, 0, 1, 0,
	16, 0, 0, 0.2, 0.3, 0.8, 0, 1, 0,
	16, 0, 1, 0.2, 0.3, 0.8, 0, 1, 0,
	0, 0, 1, 0.2, 0.3, 0.8, 0, 1, 0,

	// Left
	0, 0, 1, 0.2, 0.3, 0.8, -1, 0, 0,
	0, 16, 0, 0.2, 0.3, 0.8, -1, 0, 0,
	0, 0, 0, 0.2, 0.3, 0.8, -1, 0, 0,
	0, 0, 1, 0.2, 0.3, 0.8, -1, 0, 0,
	0, 16, 1, 0.2, 0.3, 0.8, -1, 0, 0,
	0, 16, 0, 0.2, 0.3, 0.8, -1, 0, 0,

	// Right
	16, 0, 1, 0.2, 0.3, 0.8, 1, 0, 0,
	16, 0, 0, 0.2, 0.3, 0.8, 1, 0, 0,
	16, 16, 0, 0.2, 0.3, 0.8, 1, 0, 0,
	16, 0, 1, 0.2, 0.3, 0.8, 1, 0, 0,
	16, 16, 0, 0.2, 0.3, 0.8, 1, 0, 0,
	16, 16, 1, 0.2, 0.3, 0.8, 1, 0, 0,
}

func InitMap(conn *dfhack.Conn) {
	_, err := conn.ResetMapHashes()
	if err != nil {
		panic(err)
	}
}

var (
	Map       = make(map[[3]int32]*MapBlock)
	mapSame   int32
	Dirty     = make(map[[3]int32][]float32)
	dirtyLock sync.Mutex
)

const (
	rangeX      = 4
	rangeY      = 4
	rangeZup    = 0
	rangeZdown  = 25
	rangeZchunk = 5
)

func UpdateMap(conn *dfhack.Conn) {
	center := FindCenter()

	blocks, _, err := conn.GetBlockList(&RemoteFortressReader.BlockRequest{
		MinX: proto.Int32(center[0] - rangeX),
		MaxX: proto.Int32(center[0] + rangeX),
		MinY: proto.Int32(center[1] - rangeY),
		MaxY: proto.Int32(center[1] + rangeY),
		MinZ: proto.Int32(center[2] - rangeZchunk - mapSame),
		MaxZ: proto.Int32(center[2] + rangeZup + 1 - mapSame),

		BlocksNeeded: proto.Int32(1),
	})
	if err != nil {
		panic(err)
	}

	type dirty struct {
		pos  [3]int32
		data []float32
	}
	var next []dirty

	for _, block := range blocks.MapBlocks {
		pos := [3]int32{block.GetMapX() / 16, block.GetMapY() / 16, block.GetMapZ()}
		tiles, ok := Map[pos]
		if !ok {
			tiles = new(MapBlock)
			Map[pos] = tiles
		}
		any := false
		for i, tt := range block.Tiles {
			tiles[i%16][i/16].Tiletype = Tiletype(tt)
			any = true
		}
		for i, mat := range block.Materials {
			tiles[i%16][i/16].Material = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
			any = true
		}
		for i, mat := range block.BaseMaterials {
			tiles[i%16][i/16].Base = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
			any = true
		}
		for i, mat := range block.LayerMaterials {
			tiles[i%16][i/16].Layer = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
			any = true
		}
		for i, mat := range block.VeinMaterials {
			tiles[i%16][i/16].Vein = Material{
				Type:  mat.GetMatType(),
				Index: mat.GetMatIndex(),
			}
			any = true
		}
		for i, w := range block.Water {
			tiles[i%16][i/16].Water = uint8(w)
			any = true
		}
		for i, m := range block.Magma {
			tiles[i%16][i/16].Magma = uint8(m)
			any = true
		}

		if any {
			next = append(next, dirty{pos, tiles.Generate(pos)})
			checkAdjacent := func(dx, dy int32) {
				o := [3]int32{pos[0] + dx, pos[1] + dy, pos[2]}
				if b := Map[o]; b != nil {
					next = append(next, dirty{o, b.Generate(o)})
				}
			}
			checkAdjacent(-1, 0)
			checkAdjacent(0, 1)
			checkAdjacent(0, -1)
			checkAdjacent(0, 1)
		}
	}

	if len(next) == 0 {
		mapSame += rangeZchunk
		mapSame %= rangeZdown
	} else {
		mapSame = 0
	}

	dirtyLock.Lock()
	defer dirtyLock.Unlock()
	for _, d := range next {
		Dirty[d.pos] = d.data
	}
}
