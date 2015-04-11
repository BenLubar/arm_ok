package main

import (
	"math"

	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	Perspective = mgl32.Perspective(math.Pi/2, 800.0/600.0, 0.1, 1000)
	Ambient     = mgl32.Vec3{0.1, 0.1, 0.1}
	Direction   = mgl32.Vec3{-2, 5, -20}.Normalize()
	Directional = mgl32.Vec3{1, 1, 1}
)

const float32_size = 4

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

	InitTiletypes(conn)
	InitMaterials(conn)
	InitMap(conn)

	SetupGL()

	for {
		UpdateViewInfo(conn)
		UpdateMap(conn)

		PositionCamera(CalculateCamera())
		CleanMap()

		Render(Ambient, Direction, Directional)
		DoEvents()

		if ShouldQuit() {
			break
		}
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

func CalculateCamera() mgl32.Mat4 {
	x := float32(ViewInfo.GetViewPosX() + (ViewInfo.GetViewSizeX() / 2))
	y := float32(ViewInfo.GetViewPosY() + (ViewInfo.GetViewSizeY() / 2))
	z := float32(ViewInfo.GetViewPosZ())
	return mgl32.Scale3D(-1, 1, 1).Mul4(mgl32.LookAtV(mgl32.Vec3{x + 1, y + 5, z + 10}, mgl32.Vec3{x, y, z}, mgl32.Vec3{0, 0, 1}))
}
