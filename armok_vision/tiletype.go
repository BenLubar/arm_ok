package main

import (
	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
)

type (
	Tiletype    int32
	TiletypeDef struct {
		Name      string
		Caption   string
		Shape     RemoteFortressReader.TiletypeShape
		Special   RemoteFortressReader.TiletypeSpecial
		Material  RemoteFortressReader.TiletypeMaterial
		Variant   RemoteFortressReader.TiletypeVariant
		Direction string
	}
)

var Tiletypes map[Tiletype]TiletypeDef

func (tt Tiletype) Def() TiletypeDef { return Tiletypes[tt] }

func InitTiletypes(conn *dfhack.Conn) {
	list, _, err := conn.GetTiletypeList()
	if err != nil {
		panic(err)
	}

	Tiletypes = make(map[Tiletype]TiletypeDef)
	for _, tt := range list.TiletypeList {
		Tiletypes[Tiletype(tt.GetId())] = TiletypeDef{
			Name:      tt.GetName(),
			Caption:   tt.GetCaption(),
			Shape:     tt.GetShape(),
			Special:   tt.GetSpecial(),
			Material:  tt.GetMaterial(),
			Variant:   tt.GetVariant(),
			Direction: tt.GetDirection(),
		}
	}
}
