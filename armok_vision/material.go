package main

import (
	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/go-gl/mathgl/mgl32"
)

type (
	Material struct {
		Type  int32
		Index int32
	}
	MaterialDef struct {
		ID    string
		Name  string
		Color mgl32.Vec3
	}
)

var Materials map[Material]MaterialDef

func (mat Material) Def() MaterialDef { return Materials[mat] }

func InitMaterials(conn *dfhack.Conn) {
	list, _, err := conn.GetMaterialList()
	if err != nil {
		panic(err)
	}

	Materials = make(map[Material]MaterialDef)
	for _, mat := range list.MaterialList {
		Materials[Material{
			Type:  mat.GetMatPair().GetMatType(),
			Index: mat.GetMatPair().GetMatIndex(),
		}] = MaterialDef{
			ID:   mat.GetId(),
			Name: mat.GetName(),
			Color: mgl32.Vec3{
				float32(mat.GetStateColor().GetRed()) / 255,
				float32(mat.GetStateColor().GetGreen()) / 255,
				float32(mat.GetStateColor().GetBlue()) / 255,
			},
		}
	}
}
