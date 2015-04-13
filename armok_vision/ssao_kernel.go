package main

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
)

var Kernel []float32

func init() {
	for i := 0; i < 256; i++ {
		x := rand.Float32()
		v := mgl32.Vec3{
			rand.Float32()*2 - 1,
			rand.Float32()*2 - 1,
			rand.Float32(),
		}.Normalize().Mul(x*x*0.9 + 0.1)
		Kernel = append(Kernel, v[:]...)
	}
}
