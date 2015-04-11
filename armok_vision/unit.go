package main

import (
	"math/rand"
	"sync"

	"github.com/BenLubar/arm_ok/dfhack"
)

var UnitData = []float32{
	// Bottom
	0.2, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, -1,
	0.2, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, -1,
	0.8, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, -1,
	0.8, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, -1,
	0.2, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, -1,
	0.8, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, -1,

	// Top
	0.2, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, 1,
	0.8, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, 1,
	0.2, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, 1,
	0.8, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, 1,
	0.8, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, 1,
	0.2, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 0, 1,

	// Front
	0.2, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, -1, 0,
	0.2, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, -1, 0,
	0.8, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, -1, 0,
	0.8, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, -1, 0,
	0.2, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, -1, 0,
	0.8, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, -1, 0,

	// Back
	0.2, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 1, 0,
	0.8, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 1, 0,
	0.2, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 1, 0,
	0.8, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 0, 1, 0,
	0.8, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 1, 0,
	0.2, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 0, 1, 0,

	// Left
	0.2, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), -1, 0, 0,
	0.2, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), -1, 0, 0,
	0.2, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), -1, 0, 0,
	0.2, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), -1, 0, 0,
	0.2, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), -1, 0, 0,
	0.2, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), -1, 0, 0,

	// Right
	0.8, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 1, 0, 0,
	0.8, 0.2, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 1, 0, 0,
	0.8, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 1, 0, 0,
	0.8, 0.2, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 1, 0, 0,
	0.8, 0.8, 0.2, rand.Float32(), rand.Float32(), rand.Float32(), 1, 0, 0,
	0.8, 0.8, 0.8, rand.Float32(), rand.Float32(), rand.Float32(), 1, 0, 0,
}

type Unit struct {
	Pos [3]int32
}

var Units map[int32]Unit
var unitLock sync.Mutex

func UpdateUnits(conn *dfhack.Conn) {
	list, _, err := conn.GetUnitList()
	if err != nil {
		panic(err)
	}

	units := make(map[int32]Unit)

	for _, u := range list.CreatureList {
		units[u.GetId()] = Unit{
			Pos: [3]int32{u.GetPosX(), u.GetPosY(), u.GetPosZ()},
		}
	}

	unitLock.Lock()
	Units = units
	unitLock.Unlock()
}
