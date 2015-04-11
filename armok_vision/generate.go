package main

import "github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"

func (block *MapBlock) Generate(pos [3]int32) (data []float32) {
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			tile := &block[x][y]

			x0 := float32(pos[0])*16 + float32(x)
			y0 := float32(pos[1])*16 + float32(y)
			z0 := float32(pos[2])
			x1 := x0 + 1
			y1 := y0 + 1
			z1 := z0 + 1

			mat := tile.Material.Def()

			switch tile.Tiletype.Def().Shape {
			case RemoteFortressReader.TiletypeShape_FLOOR:
				z1 -= 0.99
				fallthrough
			case RemoteFortressReader.TiletypeShape_WALL:
				// Bottom
				data = append(data, x0, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, -1)
				data = append(data, x0, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, -1)
				data = append(data, x1, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, -1)
				data = append(data, x1, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, -1)
				data = append(data, x0, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, -1)
				data = append(data, x1, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, -1)

				// Top
				data = append(data, x0, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, 1)
				data = append(data, x1, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, 1)
				data = append(data, x0, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, 1)
				data = append(data, x1, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, 1)
				data = append(data, x1, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, 1)
				data = append(data, x0, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 0, 1)

				// Front
				data = append(data, x0, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, -1, 0)
				data = append(data, x0, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, -1, 0)
				data = append(data, x1, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, -1, 0)
				data = append(data, x1, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, -1, 0)
				data = append(data, x0, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, -1, 0)
				data = append(data, x1, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, -1, 0)

				// Back
				data = append(data, x0, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 1, 0)
				data = append(data, x1, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 1, 0)
				data = append(data, x0, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 1, 0)
				data = append(data, x1, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 0, 1, 0)
				data = append(data, x1, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 1, 0)
				data = append(data, x0, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 0, 1, 0)

				// Left
				data = append(data, x0, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], -1, 0, 0)
				data = append(data, x0, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], -1, 0, 0)
				data = append(data, x0, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], -1, 0, 0)
				data = append(data, x0, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], -1, 0, 0)
				data = append(data, x0, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], -1, 0, 0)
				data = append(data, x0, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], -1, 0, 0)

				// Right
				data = append(data, x1, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 1, 0, 0)
				data = append(data, x1, y0, z0, mat.Color[0], mat.Color[1], mat.Color[2], 1, 0, 0)
				data = append(data, x1, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 1, 0, 0)
				data = append(data, x1, y0, z1, mat.Color[0], mat.Color[1], mat.Color[2], 1, 0, 0)
				data = append(data, x1, y1, z0, mat.Color[0], mat.Color[1], mat.Color[2], 1, 0, 0)
				data = append(data, x1, y1, z1, mat.Color[0], mat.Color[1], mat.Color[2], 1, 0, 0)
			}
		}
	}

	return
}
