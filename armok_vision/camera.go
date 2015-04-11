package main

import (
	"log"
	"sync"
	"time"

	"github.com/BenLubar/arm_ok/dfhack"
	"github.com/BenLubar/arm_ok/dfhack/RemoteFortressReader"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/protobuf/proto"
)

var (
	viewInfo     *RemoteFortressReader.ViewInfo
	viewOverride *[3]int32
	viewLock     sync.Mutex
)

func UpdateViewInfo(conn *dfhack.Conn) {
	info, _, err := conn.GetViewInfo()
	if err != nil {
		panic(err)
	}

	viewLock.Lock()
	if viewOverride == nil && !proto.Equal(viewInfo, info) {
		log.Println("camera move:", viewInfo, "->", info)
	}
	viewInfo = info
	viewLock.Unlock()
}

func findCenter() (x, y, z int32) {
	viewLock.Lock()
	info := viewInfo
	override := viewOverride
	viewLock.Unlock()

	if override != nil {
		return override[0], override[1], override[2]
	}

	if info == nil {
		return
	}

	return info.GetViewPosX() + (info.GetViewSizeX() / 2),
		info.GetViewPosY() + (info.GetViewSizeY() / 2),
		info.GetViewPosZ()
}

func FindCenter() [3]int32 {
	x, y, z := findCenter()
	return [3]int32{x / 16, y / 16, z}
}

var (
	eyeLerpTime = time.Second / 10
	eyeStart    time.Time
	eyePos      mgl32.Vec3
	eyeTarget   mgl32.Vec3

	targetLerpTime = time.Second / 11
	targetStart    time.Time
	targetPos      mgl32.Vec3
	targetTarget   mgl32.Vec3
)

func lerp0(a, b mgl32.Vec3, cur, total time.Duration) mgl32.Vec3 {
	if cur <= 0 {
		return a
	}
	if cur >= total {
		return b
	}

	f := float32(cur) / float32(total)

	return mgl32.Vec3{
		a[0] + (b[0]-a[0])*f,
		a[1] + (b[1]-a[1])*f,
		a[2] + (b[2]-a[2])*f,
	}
}

func lerp(v *mgl32.Vec3, start *time.Time, pos, target *mgl32.Vec3, dur time.Duration, now time.Time) {
	if !v.ApproxEqual(*target) {
		if start.IsZero() {
			*pos = *v
		} else {
			*pos = lerp0(*pos, *target, now.Sub(*start), dur)
		}
		*start = now
		*target = *v
	}
	*v = lerp0(*pos, *target, now.Sub(*start), dur)
}

func CalculateCamera() mgl32.Mat4 {
	x, y, z := findCenter()

	now := time.Now()

	eye := mgl32.Vec3{float32(x) + 1, float32(y) + 5, float32(z) + 10}
	lerp(&eye, &eyeStart, &eyePos, &eyeTarget, eyeLerpTime, now)

	target := mgl32.Vec3{float32(x), float32(y), float32(z)}
	lerp(&target, &targetStart, &targetPos, &targetTarget, targetLerpTime, now)

	return mgl32.Scale3D(-1, 1, 1).Mul4(mgl32.LookAtV(eye, target, mgl32.Vec3{0, 0, 1}))
}
