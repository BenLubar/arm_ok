package main

import (
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
	"github.com/go-gl/mathgl/mgl32"
)

func (block *MapBlock) Generate(pos [3]int32) (data []float32) {
	for x := int32(0); x < 16; x++ {
		for y := int32(0); y < 16; y++ {
			offset := func(dx, dy, dz int32) *MapTile {
				dx += x
				dy += y
				opos := pos
				for ; dx < 0; dx += 16 {
					opos[0]--
				}
				for ; dx >= 16; dx -= 16 {
					opos[0]++
				}
				for ; dy < 0; dy += 16 {
					opos[1]--
				}
				for ; dy >= 16; dy -= 16 {
					opos[1]++
				}
				opos[2] += dz

				b := block
				if opos != pos {
					b = Map[opos]
				}
				if b == nil {
					return nil
				}
				return &b[dx][dy]
			}
			tile := &block[x][y]

			x0 := float32(x)
			y0 := float32(y)
			z0 := float32(0)
			x1 := x0 + 1
			y1 := y0 + 1
			z1 := z0 + 1

			tt := tile.Tiletype.Def()
			mat := tile.Material.Def()
			c := mgl32.Vec3{
				mat.Color[0]/2 + 0.4,
				mat.Color[1]/2 + 0.4,
				mat.Color[2]/2 + 0.4,
			}

			switch tt.Shape {
			case RemoteFortressReader.TiletypeShape_FLOOR:
				z1 -= 0.9
				fallthrough
			case RemoteFortressReader.TiletypeShape_WALL:
				// Bottom
				data = append(data, x0, y0, z0, c[0], c[1], c[2], 0, 0, -1)
				data = append(data, x0, y1, z0, c[0], c[1], c[2], 0, 0, -1)
				data = append(data, x1, y0, z0, c[0], c[1], c[2], 0, 0, -1)
				data = append(data, x1, y0, z0, c[0], c[1], c[2], 0, 0, -1)
				data = append(data, x0, y1, z0, c[0], c[1], c[2], 0, 0, -1)
				data = append(data, x1, y1, z0, c[0], c[1], c[2], 0, 0, -1)

				// Top
				data = append(data, x0, y0, z1, c[0], c[1], c[2], 0, 0, 1)
				data = append(data, x1, y0, z1, c[0], c[1], c[2], 0, 0, 1)
				data = append(data, x0, y1, z1, c[0], c[1], c[2], 0, 0, 1)
				data = append(data, x1, y0, z1, c[0], c[1], c[2], 0, 0, 1)
				data = append(data, x1, y1, z1, c[0], c[1], c[2], 0, 0, 1)
				data = append(data, x0, y1, z1, c[0], c[1], c[2], 0, 0, 1)

				// Front
				if o := offset(0, 1, 0); o == nil || o.Tiletype != tile.Tiletype {
					data = append(data, x0, y1, z0, c[0], c[1], c[2], 0, 1, 0)
					data = append(data, x0, y1, z1, c[0], c[1], c[2], 0, 1, 0)
					data = append(data, x1, y1, z0, c[0], c[1], c[2], 0, 1, 0)
					data = append(data, x1, y1, z0, c[0], c[1], c[2], 0, 1, 0)
					data = append(data, x0, y1, z1, c[0], c[1], c[2], 0, 1, 0)
					data = append(data, x1, y1, z1, c[0], c[1], c[2], 0, 1, 0)
				}

				// Back
				if o := offset(0, -1, 0); o == nil || o.Tiletype != tile.Tiletype {
					data = append(data, x0, y0, z0, c[0], c[1], c[2], 0, -1, 0)
					data = append(data, x1, y0, z0, c[0], c[1], c[2], 0, -1, 0)
					data = append(data, x0, y0, z1, c[0], c[1], c[2], 0, -1, 0)
					data = append(data, x1, y0, z0, c[0], c[1], c[2], 0, -1, 0)
					data = append(data, x1, y0, z1, c[0], c[1], c[2], 0, -1, 0)
					data = append(data, x0, y0, z1, c[0], c[1], c[2], 0, -1, 0)
				}

				// Left
				if o := offset(-1, 0, 0); o == nil || o.Tiletype != tile.Tiletype {
					data = append(data, x0, y0, z1, c[0], c[1], c[2], -1, 0, 0)
					data = append(data, x0, y1, z0, c[0], c[1], c[2], -1, 0, 0)
					data = append(data, x0, y0, z0, c[0], c[1], c[2], -1, 0, 0)
					data = append(data, x0, y0, z1, c[0], c[1], c[2], -1, 0, 0)
					data = append(data, x0, y1, z1, c[0], c[1], c[2], -1, 0, 0)
					data = append(data, x0, y1, z0, c[0], c[1], c[2], -1, 0, 0)
				}

				// Right
				if o := offset(1, 0, 0); o == nil || o.Tiletype != tile.Tiletype {
					data = append(data, x1, y0, z1, c[0], c[1], c[2], 1, 0, 0)
					data = append(data, x1, y0, z0, c[0], c[1], c[2], 1, 0, 0)
					data = append(data, x1, y1, z0, c[0], c[1], c[2], 1, 0, 0)
					data = append(data, x1, y0, z1, c[0], c[1], c[2], 1, 0, 0)
					data = append(data, x1, y1, z0, c[0], c[1], c[2], 1, 0, 0)
					data = append(data, x1, y1, z1, c[0], c[1], c[2], 1, 0, 0)
				}
			}
		}
	}

	return
}
