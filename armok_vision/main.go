package main

import (
	"log"
	"math"

	"github.com/BenLubar/arm_ok/dfhack"
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

	log.Println("connected")
	InitTiletypes(conn)
	log.Println("loaded tile types")
	InitMaterials(conn)
	log.Println("loaded materials")
	InitMap(conn)
	log.Println("starting to load map")

	for {
		select {
		case ch = <-stop:
			return
		default:
		}
		UpdateViewInfo(conn)
		UpdateMap(conn)
	}
}
