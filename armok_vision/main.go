package main

import (
	"math"

	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	Width, Height   = 800, 600
	Width2, Height2 = powerOf2(Width), powerOf2(Height)

	Perspective = mgl32.Perspective(math.Pi/2, float32(Width)/float32(Height), 0.1, 100)
	Ambient     = mgl32.Vec3{0.1, 0.1, 0.1}
	Direction   = mgl32.Vec3{-2, 5, -20}.Normalize()
	Directional = mgl32.Vec3{1, 1, 1}
)

var ScreenData = []float32{
	0, 0,
	0, 1,
	1, 0,
	1, 1,
}

const float32_size = 4

func main() {
	if err := InitGL(); err != nil {
		panic(err)
	}
	defer CleanupGL()

	SetupGL()

	stopNetwork := make(chan chan struct{})
	go Network(stopNetwork)
	defer func() {
		ch := make(chan struct{})
		stopNetwork <- ch
		<-ch
	}()

	for !ShouldQuit() {
		Input()

		PositionCamera(CalculateCamera())
		CleanMap()

		Render(Ambient, Direction, Directional)
	}
}

func Network(stop chan chan struct{}) {
	var ch chan struct{}
	defer func() {
		if ch != nil {
			close(ch)
		}
	}()

	conn, err := dfhack.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	InitTiletypes(conn)
	InitMaterials(conn)
	InitMap(conn)

	for {
		select {
		case ch = <-stop:
			return
		default:
		}
		UpdateViewInfo(conn)
		UpdateMap(conn)
		UpdateUnits(conn)
	}
}

func powerOf2(x int) int {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
}
