// +build !js

package main

import (
	"os"
	"runtime"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/golang/protobuf/proto"
)

var window *glfw.Window

func InitGL() error {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		return err
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)

	var err error
	window, err = glfw.CreateWindow(800, 600, "arm_ok", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		return err
	}

	return nil
}

func CleanupGL() {
	glfw.Terminate()
}

func PositionCamera() {
	proto.MarshalText(os.Stdout, ViewInfo)
}

func UseBlockList() {
	proto.MarshalText(os.Stdout, BlockList)
}
